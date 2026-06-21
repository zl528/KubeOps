package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/kubeops/ops-kubernetes/internal/middleware"
	"github.com/kubeops/ops-kubernetes/internal/model"
	"github.com/kubeops/ops-kubernetes/internal/service"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req service.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	ctx := context.Background()
	resp, err := h.authSvc.Login(ctx, req)
	if err != nil {
		writeError(w, http.StatusUnauthorized, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: resp})
}

func (h *AuthHandler) HandleGetCurrentUser(w http.ResponseWriter, r *http.Request) {
	userID, username, role := GetUserFromContext(r.Context())

	ctx := context.Background()
	user, err := h.authSvc.GetUserByID(ctx, userID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}

	_ = username
	_ = role

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: user})
}

func (h *AuthHandler) HandleListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	users, err := h.authSvc.ListUsers(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: users})
}

func (h *AuthHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Role     string `json:"role"`
		Email    string `json:"email"`
		RoleID   int64  `json:"roleId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "username and password are required")
		return
	}

	ctx := context.Background()
	userID, err := h.authSvc.CreateUser(ctx, req.Username, req.Password, req.Role, req.Email, req.RoleID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "user created", Data: map[string]int64{"userId": userID}})
}

func (h *AuthHandler) HandleUpdatePassword(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	userID, _, _ := GetUserFromContext(r.Context())

	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		writeError(w, http.StatusBadRequest, "old and new passwords are required")
		return
	}

	ctx := context.Background()
	if err := h.authSvc.UpdatePassword(ctx, userID, req.OldPassword, req.NewPassword); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "password updated"})
}

func (h *AuthHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		UserID int64 `json:"userId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	currentUserID, _, _ := GetUserFromContext(r.Context())
	if currentUserID == req.UserID {
		writeError(w, http.StatusBadRequest, "cannot delete yourself")
		return
	}

	// Prevent deleting admin user
	ctx := context.Background()
	user, err := h.authSvc.GetUserByID(ctx, req.UserID)
	if err != nil {
		writeError(w, http.StatusNotFound, "user not found")
		return
	}
	if user.Username == "admin" {
		writeError(w, http.StatusBadRequest, "cannot delete admin user")
		return
	}

	if err := h.authSvc.HardDeleteUser(ctx, req.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "user deleted"})
}

func (h *AuthHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		UserID      int64  `json:"userId"`
		Username    string `json:"username"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
		RoleID      *int64 `json:"roleId"`
		Password    string `json:"password"`
		Status      *int   `json:"status"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	log.Printf("[DEBUG] HandleUpdateUser: userId=%d, status=%v, roleId=%v", req.UserID, req.Status, req.RoleID)

	ctx := context.Background()

	// Prevent disabling admin user
	if req.Status != nil && *req.Status == 0 {
		user, err := h.authSvc.GetUserByID(ctx, req.UserID)
		if err != nil {
			writeError(w, http.StatusNotFound, "user not found")
			return
		}
		if user.Username == "admin" {
			writeError(w, http.StatusBadRequest, "cannot disable admin user")
			return
		}
	}

	// Update role_id if provided
	if req.RoleID != nil {
		log.Printf("[DEBUG] HandleUpdateUser: updating role_id to %d", *req.RoleID)
		if err := h.authSvc.UpdateUserRole(ctx, req.UserID, *req.RoleID); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Update display_name and email if provided
	if req.DisplayName != "" || req.Email != "" {
		log.Printf("[DEBUG] HandleUpdateUser: updating profile")
		if err := h.authSvc.UpdateUserProfile(ctx, req.UserID, req.Email, req.DisplayName); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Update status if provided
	if req.Status != nil {
		log.Printf("[DEBUG] HandleUpdateUser: updating status to %d", *req.Status)
		if err := h.authSvc.UpdateUserStatus(ctx, req.UserID, *req.Status); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	// Update password if provided
	if req.Password != "" {
		if err := h.authSvc.AdminResetPassword(ctx, req.UserID, req.Password); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "user updated"})
}

func (h *AuthHandler) HandleListRoles(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	roles, err := h.authSvc.ListRoles(ctx)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: roles})
}

func (h *AuthHandler) HandleGetRole(w http.ResponseWriter, r *http.Request) {
	roleIDStr := r.URL.Query().Get("id")
	if roleIDStr == "" {
		writeError(w, http.StatusBadRequest, "role id is required")
		return
	}

	var roleID int64
	if _, err := fmt.Sscanf(roleIDStr, "%d", &roleID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid role id")
		return
	}

	ctx := context.Background()
	role, err := h.authSvc.GetRole(ctx, roleID)
	if err != nil {
		writeError(w, http.StatusNotFound, "role not found")
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: role})
}

func (h *AuthHandler) HandleCreateRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Permissions string `json:"permissions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "role name is required")
		return
	}

	ctx := context.Background()
	if err := h.authSvc.CreateRole(ctx, req.Name, req.Description, req.Permissions); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "role created"})
}

func (h *AuthHandler) HandleUpdateRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		RoleID      int64  `json:"roleId"`
		Name        string `json:"name"`
		Description string `json:"description"`
		Permissions string `json:"permissions"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	if err := h.authSvc.UpdateRole(ctx, req.RoleID, req.Name, req.Description, req.Permissions); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "role updated"})
}

func (h *AuthHandler) HandleDeleteRole(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		RoleID int64 `json:"roleId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	if err := h.authSvc.DeleteRole(ctx, req.RoleID); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "role deleted"})
}

func (h *AuthHandler) HandleListAuditLogs(w http.ResponseWriter, r *http.Request) {
	userID, _, role := GetUserFromContext(r.Context())

	// Non-admin users can only see their own logs
	if role == "admin" {
		userID = 0 // 0 means all users
	}

	ctx := context.Background()
	logs, err := h.authSvc.ListAuditLogs(ctx, userID, 100, 0)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: logs})
}

func (h *AuthHandler) HandleGetUserClusters(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("userId")
	if userIDStr == "" {
		writeError(w, http.StatusBadRequest, "userId is required")
		return
	}

	var userID int64
	if _, err := fmt.Sscanf(userIDStr, "%d", &userID); err != nil {
		writeError(w, http.StatusBadRequest, "invalid userId")
		return
	}

	ctx := context.Background()
	clusters, err := h.authSvc.GetUserClusters(ctx, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: clusters})
}

func (h *AuthHandler) HandleSetUserClusters(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var req struct {
		UserID      int64    `json:"userId"`
		ClusterNames []string `json:"clusterNames"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := context.Background()
	if err := h.authSvc.SetUserClusters(ctx, req.UserID, req.ClusterNames); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "clusters updated"})
}

func (h *AuthHandler) HandleCheckClusterAccess(w http.ResponseWriter, r *http.Request) {
	userID, _, _ := GetUserFromContext(r.Context())
	clusterName := r.URL.Query().Get("cluster")

	if clusterName == "" {
		writeError(w, http.StatusBadRequest, "cluster name is required")
		return
	}

	ctx := context.Background()
	hasAccess, err := h.authSvc.HasClusterAccess(ctx, userID, clusterName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: map[string]bool{"hasAccess": hasAccess}})
}

func (h *AuthHandler) HandleGetUserPermissions(w http.ResponseWriter, r *http.Request) {
	userID, _, _ := GetUserFromContext(r.Context())

	ctx := context.Background()
	permissions, err := h.authSvc.GetUserPermissions(ctx, userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, model.Response{Code: 0, Message: "ok", Data: permissions})
}

func GetUserFromContext(ctx context.Context) (int64, string, string) {
	userID, _ := ctx.Value(middleware.UserIDKey).(int64)
	username, _ := ctx.Value(middleware.UsernameKey).(string)
	role, _ := ctx.Value(middleware.UserRoleKey).(string)
	return userID, username, role
}
