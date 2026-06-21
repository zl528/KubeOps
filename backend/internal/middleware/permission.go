package middleware

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

type PermissionChecker struct {
	authSvc interface {
		GetUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error)
	}
}

func NewPermissionChecker(authSvc interface {
	GetUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error)
}) *PermissionChecker {
	return &PermissionChecker{authSvc: authSvc}
}

// RequirePermission checks if user has the required module permission
func (pc *PermissionChecker) RequirePermission(module, action string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Admin bypasses all permission checks
			role, _ := r.Context().Value(UserRoleKey).(string)
			if role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			userID, _ := r.Context().Value(UserIDKey).(int64)
			if userID == 0 {
				http.Error(w, `{"code":401,"message":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			ctx := r.Context()
			permissions, err := pc.authSvc.GetUserPermissions(ctx, userID)
			if err != nil {
				http.Error(w, `{"code":500,"message":"failed to get permissions"}`, http.StatusInternalServerError)
				return
			}

			// Check module permission
			modules, ok := permissions["modules"].(map[string]interface{})
			if !ok {
				http.Error(w, `{"code":403,"message":"forbidden: no permissions"}`, http.StatusForbidden)
				return
			}

			modulePerms, ok := modules[module].(map[string]interface{})
			if !ok {
				http.Error(w, `{"code":403,"message":"forbidden: module not found"}`, http.StatusForbidden)
				return
			}

			allowed, ok := modulePerms[action].(bool)
			if !ok || !allowed {
				http.Error(w, `{"code":403,"message":"forbidden: insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Map URL path to module
func GetModuleFromPath(path string) string {
	path = strings.TrimPrefix(path, "/api/")
	parts := strings.Split(path, "/")
	if len(parts) == 0 {
		return ""
	}

	resource := parts[0]
	switch resource {
	case "pods", "deployments", "statefulsets", "daemonsets", "cronjobs", "jobs":
		return "workloads"
	case "services", "ingresses", "networkpolicies", "endpoints":
		return "network"
	case "configmaps", "secrets":
		return "storage"
	case "persistentvolumes", "persistentvolumeclaims", "storageclasses":
		return "storage"
	case "roles", "clusterroles", "rolebindings", "clusterrolebindings", "serviceaccounts":
		return "rbac"
	case "limitranges", "hpas", "resourcequotas":
		return "workloads"
	case "namespaces", "nodes":
		return "workloads"
	case "events":
		return "workloads"
	default:
		return ""
	}
}

// Map HTTP method to action
func GetActionFromMethod(method string) string {
	switch method {
	case "GET":
		return "view"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "edit"
	case "DELETE":
		return "delete"
	default:
		return "view"
	}
}

// AutoPermission automatically checks permission based on URL and method
func (pc *PermissionChecker) AutoPermission() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Admin bypasses all permission checks
			role, _ := r.Context().Value(UserRoleKey).(string)
			if role == "admin" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip permission check for auth endpoints
			if strings.HasPrefix(r.URL.Path, "/api/auth/") {
				next.ServeHTTP(w, r)
				return
			}

			// Skip permission check for health endpoint
			if r.URL.Path == "/api/health" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip permission check for websocket/exec endpoints
			if strings.HasPrefix(r.URL.Path, "/api/pods/exec") {
				// Terminal requires workloads.edit permission
				ctx := r.Context()
				userID, _ := ctx.Value(UserIDKey).(int64)
				if userID == 0 {
					http.Error(w, `{"code":401,"message":"unauthorized"}`, http.StatusUnauthorized)
					return
				}
				permissions, err := pc.authSvc.GetUserPermissions(ctx, userID)
				if err != nil {
					http.Error(w, `{"code":500,"message":"failed to get permissions"}`, http.StatusInternalServerError)
					return
				}
				modules, ok := permissions["modules"].(map[string]interface{})
				if !ok {
					http.Error(w, `{"code":403,"message":"forbidden"}`, http.StatusForbidden)
					return
				}
				wl, ok := modules["workloads"].(map[string]interface{})
				editOk, _ := wl["edit"].(bool)
				if !ok || !editOk {
					http.Error(w, `{"code":403,"message":"forbidden: terminal requires workloads.edit permission"}`, http.StatusForbidden)
					return
				}
				next.ServeHTTP(w, r)
				return
			}

			// Skip permission check for logs endpoint
			if strings.HasPrefix(r.URL.Path, "/api/pods/logs") {
				next.ServeHTTP(w, r)
				return
			}

			// Skip permission check for clusters endpoint
			if r.URL.Path == "/api/clusters" {
				next.ServeHTTP(w, r)
				return
			}

			userID, _ := r.Context().Value(UserIDKey).(int64)
			if userID == 0 {
				http.Error(w, `{"code":401,"message":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			module := GetModuleFromPath(r.URL.Path)
			if module == "" {
				// Unknown module, allow access
				next.ServeHTTP(w, r)
				return
			}

			action := GetActionFromMethod(r.Method)

			log.Printf("[PERM] Checking permission for user %d: module=%s, action=%s, path=%s", userID, module, action, r.URL.Path)

			ctx := r.Context()
			permissions, err := pc.authSvc.GetUserPermissions(ctx, userID)
			if err != nil {
				log.Printf("[PERM] Error getting permissions: %v", err)
				http.Error(w, `{"code":500,"message":"failed to get permissions"}`, http.StatusInternalServerError)
				return
			}

			log.Printf("[PERM] User permissions: %v", permissions)

			// Check module permission
			modules, ok := permissions["modules"].(map[string]interface{})
			if !ok {
				log.Printf("[PERM] No modules found in permissions")
				http.Error(w, `{"code":403,"message":"forbidden: no permissions"}`, http.StatusForbidden)
				return
			}

			modulePerms, ok := modules[module].(map[string]interface{})
			if !ok {
				log.Printf("[PERM] Module %s not found in permissions", module)
				http.Error(w, `{"code":403,"message":"forbidden: module not found"}`, http.StatusForbidden)
				return
			}

			allowed, ok := modulePerms[action].(bool)
			log.Printf("[PERM] Permission check: module=%s, action=%s, allowed=%v, ok=%v", module, action, allowed, ok)
			if !ok || !allowed {
				http.Error(w, `{"code":403,"message":"forbidden: insufficient permissions"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func writeJSONPerm(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}
