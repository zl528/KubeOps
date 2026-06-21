package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/kubeops/ops-kubernetes/internal/service"
)

type AuditMiddleware struct {
	auditSvc *service.AuditService
}

func NewAuditMiddleware(auditSvc *service.AuditService) *AuditMiddleware {
	return &AuditMiddleware{auditSvc: auditSvc}
}

func (am *AuditMiddleware) RecordAudit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only record write operations (POST, PUT, DELETE)
		if r.Method == "GET" || r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth, health, monitor, exec, and audit endpoints
		if r.URL.Path == "/api/health" ||
			strings.HasPrefix(r.URL.Path, "/api/auth/") ||
			strings.HasPrefix(r.URL.Path, "/api/pods/exec") ||
			strings.HasPrefix(r.URL.Path, "/api/monitor") ||
			strings.HasPrefix(r.URL.Path, "/api/audit") {
			next.ServeHTTP(w, r)
			return
		}

		// Get user info from context
		username, _ := r.Context().Value(UsernameKey).(string)

		// Determine action from method
		action := getActionFromMethod(r.Method)

		// Parse resource type from URL path
		resource := parseResourceType(r.URL.Path)

		// Get name and namespace from query parameters first
		name := r.URL.Query().Get("name")
		namespace := r.URL.Query().Get("namespace")

		// If not in query, try to get from request body
		if name == "" || namespace == "" {
			bodyName, bodyNs := extractNameNamespace(r)
			if name == "" {
				name = bodyName
			}
			if namespace == "" {
				namespace = bodyNs
			}
		}

		// Record the audit log
		go func() {
			am.auditSvc.Record(service.AuditLog{
				User:      username,
				Action:    action,
				Resource:  resource,
				Name:      name,
				Namespace: namespace,
				Status:    "success",
				IP:        getClientIP(r),
				UserAgent: r.UserAgent(),
				Timestamp: time.Now(),
			})
		}()

		next.ServeHTTP(w, r)
	})
}

func getActionFromMethod(method string) string {
	switch method {
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return "unknown"
	}
}

func parseResourceType(path string) string {
	path = strings.TrimPrefix(path, "/api/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func extractNameNamespace(r *http.Request) (name, namespace string) {
	// Read body
	if r.Body == nil {
		return "", ""
	}
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "", ""
	}
	// Restore body for downstream handlers
	r.Body = io.NopCloser(bytes.NewBuffer(body))

	// Try to parse JSON body
	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		return "", ""
	}

	if n, ok := data["name"].(string); ok {
		name = n
	}
	if ns, ok := data["namespace"].(string); ok {
		namespace = ns
	}

	return name, namespace
}

func getClientIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}
	return r.RemoteAddr
}
