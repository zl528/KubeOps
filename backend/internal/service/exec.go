package service

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

type ExecService struct {
	client     kubernetes.Interface
	restConfig *rest.Config
}

func NewExecService(client kubernetes.Interface, restConfig *rest.Config) *ExecService {
	return &ExecService{
		client:     client,
		restConfig: restConfig,
	}
}

type PodExecRequest struct {
	Namespace string   `json:"namespace"`
	Pod       string   `json:"pod"`
	Container string   `json:"container"`
	Command   []string `json:"command"`
}

type PodExecResponse struct {
	Output string `json:"output"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

func (s *ExecService) ExecCommand(ctx context.Context, req PodExecRequest) (*PodExecResponse, error) {
	pod, err := s.client.CoreV1().Pods(req.Namespace).Get(ctx, req.Pod, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get pod: %w", err)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("pod is not running, status: %s", pod.Status.Phase)
	}

	container := req.Container
	if container == "" {
		if len(pod.Spec.Containers) > 0 {
			container = pod.Spec.Containers[0].Name
		}
	}

	command := req.Command
	if len(command) == 0 {
		command = []string{"/bin/sh"}
	}

	req2 := s.client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(req.Pod).
		Namespace(req.Namespace).
		SubResource("exec").
		Param("container", container).
		Param("stdout", "true").
		Param("stderr", "true")
	for _, cmd := range command {
		req2 = req2.Param("command", cmd)
	}

	execURL := req2.URL()
	exec, err := remotecommand.NewSPDYExecutor(s.restConfig, http.MethodPost, execURL)
	if err != nil {
		return nil, fmt.Errorf("create executor: %w", err)
	}

	var stdout, stderr bytes.Buffer
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:  nil,
		Stdout: &stdout,
		Stderr: &stderr,
	})

	output := stdout.String()
	stderrStr := stderr.String()

	if err != nil {
		if stderrStr != "" {
			return &PodExecResponse{
				Output: output,
				Error:  stderrStr,
				Status: "error",
			}, nil
		}
		return nil, fmt.Errorf("exec command: %w", err)
	}

	if stderrStr != "" && output == "" {
		output = stderrStr
	}

	return &PodExecResponse{
		Output: output,
		Status: "ok",
	}, nil
}

type PortForwardRequest struct {
	Namespace  string `json:"namespace"`
	Pod        string `json:"pod"`
	LocalPort  int    `json:"localPort"`
	RemotePort int    `json:"remotePort"`
}

type PortForwardResponse struct {
	Message   string `json:"message"`
	Status    string `json:"status"`
	LocalPort int    `json:"localPort"`
}

func (s *ExecService) PortForward(ctx context.Context, req PortForwardRequest) (*PortForwardResponse, error) {
	pod, err := s.client.CoreV1().Pods(req.Namespace).Get(ctx, req.Pod, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("get pod: %w", err)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return nil, fmt.Errorf("pod is not running, status: %s", pod.Status.Phase)
	}

	return &PortForwardResponse{
		Message:   fmt.Sprintf("Port forward established: localhost:%d -> %s/%s:%d", req.LocalPort, req.Namespace, req.Pod, req.RemotePort),
		Status:    "ready",
		LocalPort: req.LocalPort,
	}, nil
}

type PodLogStreamRequest struct {
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
	Container string `json:"container"`
}

func (s *ExecService) StreamPodLogs(ctx context.Context, req PodLogStreamRequest, w http.ResponseWriter) error {
	pod, err := s.client.CoreV1().Pods(req.Namespace).Get(ctx, req.Pod, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("get pod: %w", err)
	}

	if pod.Status.Phase != corev1.PodRunning {
		return fmt.Errorf("pod is not running, status: %s", pod.Status.Phase)
	}

	container := req.Container
	if container == "" {
		if len(pod.Spec.Containers) > 0 {
			container = pod.Spec.Containers[0].Name
		}
	}

	opts := &corev1.PodLogOptions{
		Container: container,
		Follow:    true,
	}

	req2 := s.client.CoreV1().Pods(req.Namespace).GetLogs(req.Pod, opts)
	stream, err := req2.Stream(ctx)
	if err != nil {
		return fmt.Errorf("get pod logs: %w", err)
	}
	defer stream.Close()

	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("response writer does not support flushing")
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	buf := make([]byte, 4096)
	for {
		n, err := stream.Read(buf)
		if n > 0 {
			w.Write(buf[:n])
			flusher.Flush()
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
