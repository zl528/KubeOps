package service

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type LogService struct {
	client kubernetes.Interface
}

func NewLogService(client kubernetes.Interface) *LogService {
	return &LogService{client: client}
}

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Level     string `json:"level,omitempty"`
	Source    string `json:"source,omitempty"`
	Message   string `json:"message"`
}

type LogQuery struct {
	Namespace   string `json:"namespace"`
	Pod         string `json:"pod"`
	Container   string `json:"container"`
	Follow      bool   `json:"follow"`
	TailLines   int64  `json:"tailLines"`
	SinceTime   string `json:"sinceTime,omitempty"`
	Search      string `json:"search,omitempty"`
	Level       string `json:"level,omitempty"`
}

type LogResponse struct {
	Lines    []LogEntry `json:"lines"`
	Total    int        `json:"total"`
	HasMore  bool       `json:"hasMore"`
}

func (s *LogService) GetPodLogs(ctx context.Context, query LogQuery) (*LogResponse, error) {
	if query.TailLines == 0 {
		query.TailLines = 1000
	}

	opts := &corev1.PodLogOptions{
		Container:  query.Container,
		TailLines:  &query.TailLines,
		Follow:     query.Follow,
	}

	if query.SinceTime != "" {
		t, err := time.Parse(time.RFC3339, query.SinceTime)
		if err == nil {
			sinceSeconds := int64(time.Since(t).Seconds())
			opts.SinceSeconds = &sinceSeconds
		}
	}

	req := s.client.CoreV1().Pods(query.Namespace).GetLogs(query.Pod, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("get pod logs: %w", err)
	}
	defer stream.Close()

	var lines []LogEntry
	scanner := bufio.NewScanner(stream)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		entry := parseLogLine(line)

		if query.Search != "" && !strings.Contains(strings.ToLower(line), strings.ToLower(query.Search)) {
			continue
		}

		if query.Level != "" && entry.Level != "" && !strings.EqualFold(entry.Level, query.Level) {
			continue
		}

		lines = append(lines, entry)
	}

	return &LogResponse{
		Lines:   lines,
		Total:   len(lines),
		HasMore: false,
	}, nil
}

func (s *LogService) GetPodLogsFromTime(ctx context.Context, query LogQuery, since time.Time) (*LogResponse, error) {
	sinceSeconds := int64(time.Since(since).Seconds())
	opts := &corev1.PodLogOptions{
		Container:  query.Container,
		SinceSeconds: &sinceSeconds,
	}

	req := s.client.CoreV1().Pods(query.Namespace).GetLogs(query.Pod, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("get pod logs: %w", err)
	}
	defer stream.Close()

	var lines []LogEntry
	scanner := bufio.NewScanner(stream)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		entry := parseLogLine(line)
		lines = append(lines, entry)
	}

	return &LogResponse{
		Lines:   lines,
		Total:   len(lines),
		HasMore: false,
	}, nil
}

func (s *LogService) StreamPodLogs(ctx context.Context, query LogQuery, writer io.Writer) error {
	opts := &corev1.PodLogOptions{
		Container: query.Container,
		Follow:    true,
	}

	req := s.client.CoreV1().Pods(query.Namespace).GetLogs(query.Pod, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return fmt.Errorf("get pod logs: %w", err)
	}
	defer stream.Close()

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			writer.Write(buf[:n])
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
	}

	return nil
}

func (s *LogService) DownloadPodLogs(ctx context.Context, query LogQuery) (*bytes.Buffer, error) {
	opts := &corev1.PodLogOptions{
		Container: query.Container,
	}

	req := s.client.CoreV1().Pods(query.Namespace).GetLogs(query.Pod, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("get pod logs: %w", err)
	}
	defer stream.Close()

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, stream); err != nil {
		return nil, fmt.Errorf("read logs: %w", err)
	}

	return &buf, nil
}

func (s *LogService) ListPodContainers(ctx context.Context, namespace, podName string) ([]string, error) {
	pod, err := s.client.CoreV1().Pods(namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	var containers []string
	for _, c := range pod.Spec.Containers {
		containers = append(containers, c.Name)
	}

	return containers, nil
}

func (s *LogService) SearchLogs(ctx context.Context, query LogQuery, keyword string) (*LogResponse, error) {
	query.Search = keyword
	return s.GetPodLogs(ctx, query)
}

func (s *LogService) GetLogsByLevel(ctx context.Context, query LogQuery, level string) (*LogResponse, error) {
	query.Level = level
	return s.GetPodLogs(ctx, query)
}

func (s *LogService) GetPreviousPodLogs(ctx context.Context, query LogQuery) (*LogResponse, error) {
	opts := &corev1.PodLogOptions{
		Container: query.Container,
		Previous:  true,
		TailLines: &query.TailLines,
	}

	req := s.client.CoreV1().Pods(query.Namespace).GetLogs(query.Pod, opts)
	stream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("get pod logs: %w", err)
	}
	defer stream.Close()

	var lines []LogEntry
	scanner := bufio.NewScanner(stream)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		entry := parseLogLine(line)
		lines = append(lines, entry)
	}

	return &LogResponse{
		Lines:   lines,
		Total:   len(lines),
		HasMore: false,
	}, nil
}

func parseLogLine(line string) LogEntry {
	entry := LogEntry{
		Message: line,
	}

	lowerLine := strings.ToLower(line)

	if strings.Contains(lowerLine, "error") || strings.Contains(lowerLine, "err") {
		entry.Level = "error"
	} else if strings.Contains(lowerLine, "warn") || strings.Contains(lowerLine, "warning") {
		entry.Level = "warn"
	} else if strings.Contains(lowerLine, "info") {
		entry.Level = "info"
	} else if strings.Contains(lowerLine, "debug") {
		entry.Level = "debug"
	}

	if len(line) > 30 && line[10] == 'T' {
		entry.Timestamp = line[:30]
		if len(line) > 31 {
			rest := line[31:]
			if idx := strings.Index(rest, " "); idx > 0 {
				entry.Source = rest[:idx]
				entry.Message = rest[idx+1:]
			}
		}
	}

	return entry
}

// Unused imports
var _ = io.Reader(nil)
var _ = time.Now()
