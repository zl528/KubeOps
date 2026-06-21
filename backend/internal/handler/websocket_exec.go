package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/remotecommand"
)

var wsUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type wsExecRequest struct {
	Namespace string   `json:"namespace"`
	Pod       string   `json:"pod"`
	Container string   `json:"container"`
	Command   []string `json:"command"`
	Columns   int      `json:"columns"`
	Rows      int      `json:"rows"`
}

func (h *ExecHandler) HandleWebSocketExec(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("WebSocket exec panic: %v\n", r)
		}
	}()

	conn, err := wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("WebSocket upgrade error: %v\n", err)
		return
	}
	defer conn.Close()

	_, msg, err := conn.ReadMessage()
	if err != nil {
		fmt.Printf("WS read error: %v\n", err)
		return
	}
	fmt.Printf("WS received: %s\n", string(msg))

	var req wsExecRequest
	if err := json.Unmarshal(msg, &req); err != nil {
		conn.WriteJSON(map[string]string{"error": "invalid request"})
		return
	}

	if req.Namespace == "" || req.Pod == "" {
		conn.WriteJSON(map[string]string{"error": "namespace and pod are required"})
		return
	}

	if len(req.Command) == 0 {
		req.Command = []string{"/bin/sh", "-il"}
	}

	client := h.clientGetter.GetK8sClient()
	restConfig := h.clientGetter.GetRestConfig()

	pod, err := client.CoreV1().Pods(req.Namespace).Get(r.Context(), req.Pod, metav1.GetOptions{})
	if err != nil {
		fmt.Printf("WS get pod error: %v\n", err)
		conn.WriteJSON(map[string]string{"error": err.Error()})
		return
	}
	fmt.Printf("WS got pod: %s/%s phase=%s\n", req.Namespace, req.Pod, pod.Status.Phase)

	if pod.Status.Phase != corev1.PodRunning {
		conn.WriteJSON(map[string]string{"error": "pod is not running, status: " + string(pod.Status.Phase)})
		return
	}

	container := req.Container
	if container == "" {
		if len(pod.Spec.Containers) > 0 {
			container = pod.Spec.Containers[0].Name
		}
	}

	execReq := client.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(req.Pod).
		Namespace(req.Namespace).
		SubResource("exec").
		Param("container", container).
		Param("stdout", "true").
		Param("stderr", "true").
		Param("stdin", "true").
		Param("tty", "true")
	for _, cmd := range req.Command {
		execReq = execReq.Param("command", cmd)
	}

	exec, err := remotecommand.NewSPDYExecutor(restConfig, http.MethodPost, execReq.URL())
	if err != nil {
		fmt.Printf("WS create executor error: %v\n", err)
		conn.WriteJSON(map[string]string{"error": "create executor: " + err.Error()})
		return
	}
	fmt.Printf("WS executor created, starting stream...\n")

	stdinReader, stdinWriter := io.Pipe()
	stdoutReader, stdoutWriter := io.Pipe()
	stderrReader, stderrWriter := io.Pipe()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		exec.StreamWithContext(r.Context(), remotecommand.StreamOptions{
			Stdin:  stdinReader,
			Stdout: stdoutWriter,
			Stderr: stderrWriter,
			Tty:    true,
		})
		stdoutWriter.Close()
		stderrWriter.Close()
		stdinWriter.Close()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 8192)
		for {
			n, err := stdoutReader.Read(buf)
			if n > 0 {
				fmt.Printf("WS stdout: %d bytes\n", n)
				if writeErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
					fmt.Printf("WS write error: %v\n", writeErr)
					return
				}
			}
			if err != nil {
				fmt.Printf("WS stdout read error: %v\n", err)
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		buf := make([]byte, 8192)
		for {
			n, err := stderrReader.Read(buf)
			if n > 0 {
				if writeErr := conn.WriteMessage(websocket.BinaryMessage, buf[:n]); writeErr != nil {
					return
				}
			}
			if err != nil {
				return
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		defer stdinWriter.Close()
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				return
			}
			var input struct {
				Type string `json:"type"`
				Data string `json:"data"`
				Cols int    `json:"cols"`
				Rows int    `json:"rows"`
			}
			if err := json.Unmarshal(msg, &input); err != nil {
				continue
			}
			switch input.Type {
			case "input":
				stdinWriter.Write([]byte(input.Data))
			case "resize":
			}
		}
	}()

	wg.Wait()
}
