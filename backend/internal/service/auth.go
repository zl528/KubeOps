package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db        *sql.DB
	jwtSecret []byte
}

type User struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Role        string    `json:"role"`
	RoleID      int64     `json:"roleId"`
	RoleName    string    `json:"roleName"`
	DisplayName string    `json:"displayName"`
	Email       string    `json:"email,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	LastLogin   time.Time `json:"lastLogin,omitempty"`
	Status      int       `json:"status"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	User      User   `json:"user"`
	ExpiresAt int64  `json:"expiresAt"`
}

type Claims struct {
	UserID   int64  `json:"userId"`
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func NewAuthService(db *sql.DB, jwtSecret string) *AuthService {
	if jwtSecret == "" {
		jwtSecret = "kubeops-secret-key-change-in-production"
	}
	return &AuthService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (*LoginResponse, error) {
	var user User
	var password string
	var email sql.NullString

	err := s.db.QueryRow(
		"SELECT id, username, password, role, email, created_at, status FROM users WHERE username = ? AND status = 1",
		req.Username,
	).Scan(&user.ID, &user.Username, &password, &user.Role, &email, &user.CreatedAt, &user.Status)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("invalid username or password")
	}
	if err != nil {
		return nil, fmt.Errorf("database error: %w", err)
	}

	if email.Valid {
		user.Email = email.String
	}

	if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid username or password")
	}

	// Update last login
	s.db.Exec("UPDATE users SET last_login = ? WHERE id = ?", time.Now(), user.ID)

	// Generate JWT
	expiresAt := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}

	return &LoginResponse{
		Token:     tokenString,
		User:      user,
		ExpiresAt: expiresAt.Unix(),
	}, nil
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return s.jwtSecret, nil
	})

	if err != nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	return claims, nil
}

func (s *AuthService) ListUsers(ctx context.Context) ([]User, error) {
	rows, err := s.db.Query(`SELECT u.id, u.username, u.role, u.role_id, COALESCE(r.name, '') as role_name, 
		COALESCE(u.display_name, '') as display_name, u.email, u.created_at, u.last_login, u.status 
		FROM users u LEFT JOIN roles r ON u.role_id = r.id ORDER BY u.created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		var email, displayName, roleName sql.NullString
		var lastLogin sql.NullTime
		var roleID sql.NullInt64
		if err := rows.Scan(&u.ID, &u.Username, &u.Role, &roleID, &roleName, &displayName, &email, &u.CreatedAt, &lastLogin, &u.Status); err != nil {
			continue
		}
		if email.Valid {
			u.Email = email.String
		}
		if displayName.Valid {
			u.DisplayName = displayName.String
		}
		if roleName.Valid {
			u.RoleName = roleName.String
		}
		if roleID.Valid {
			u.RoleID = roleID.Int64
		}
		if lastLogin.Valid {
			u.LastLogin = lastLogin.Time
		}
		users = append(users, u)
	}
	return users, nil
}

func (s *AuthService) CreateUser(ctx context.Context, username, password, role, email string, roleID int64) (int64, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return 0, fmt.Errorf("hash password: %w", err)
	}

	if role == "" {
		role = "user"
	}

	result, err := s.db.Exec(
		"INSERT INTO users (username, password, role, email, role_id) VALUES (?, ?, ?, ?, ?)",
		username, string(hashedPassword), role, email, roleID,
	)
	if err != nil {
		return 0, fmt.Errorf("create user: %w", err)
	}
	userID, _ := result.LastInsertId()
	return userID, nil
}

func (s *AuthService) UpdateUser(ctx context.Context, userID int64, email, displayName string, roleID int64) error {
	_, err := s.db.Exec(
		"UPDATE users SET email = ?, display_name = ?, role_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		email, displayName, roleID, userID,
	)
	return err
}

func (s *AuthService) UpdateUserRole(ctx context.Context, userID int64, roleID int64) error {
	_, err := s.db.Exec(
		"UPDATE users SET role_id = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		roleID, userID,
	)
	return err
}

func (s *AuthService) UpdateUserProfile(ctx context.Context, userID int64, email, displayName string) error {
	_, err := s.db.Exec(
		"UPDATE users SET email = ?, display_name = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		email, displayName, userID,
	)
	return err
}

func (s *AuthService) UpdateUserStatus(ctx context.Context, userID int64, status int) error {
	_, err := s.db.Exec(
		"UPDATE users SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?",
		status, userID,
	)
	return err
}

func (s *AuthService) UpdatePassword(ctx context.Context, userID int64, oldPassword, newPassword string) error {
	var password string
	err := s.db.QueryRow("SELECT password FROM users WHERE id = ?", userID).Scan(&password)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if oldPassword != "" {
		if err := bcrypt.CompareHashAndPassword([]byte(password), []byte(oldPassword)); err != nil {
			return fmt.Errorf("incorrect old password")
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = s.db.Exec("UPDATE users SET password = ?, updated_at = ? WHERE id = ?", string(hashedPassword), time.Now(), userID)
	return err
}

func (s *AuthService) AdminResetPassword(ctx context.Context, userID int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	_, err = s.db.Exec("UPDATE users SET password = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", string(hashedPassword), userID)
	return err
}

func (s *AuthService) HardDeleteUser(ctx context.Context, userID int64) error {
	_, err := s.db.Exec("DELETE FROM users WHERE id = ?", userID)
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	_, _ = s.db.Exec("DELETE FROM user_clusters WHERE user_id = ?", userID)
	return nil
}

func (s *AuthService) DeleteUser(ctx context.Context, userID int64) error {
	_, err := s.db.Exec("UPDATE users SET status = 0 WHERE id = ?", userID)
	return err
}

func (s *AuthService) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	var user User
	var email sql.NullString
	var displayName sql.NullString
	var roleName sql.NullString
	err := s.db.QueryRow(
		`SELECT u.id, u.username, u.role, u.role_id, COALESCE(r.name, '') as role_name, 
		COALESCE(u.display_name, '') as display_name, u.email, u.created_at, u.status 
		FROM users u LEFT JOIN roles r ON u.role_id = r.id WHERE u.id = ?`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Role, &user.RoleID, &roleName, &displayName, &email, &user.CreatedAt, &user.Status)
	if err != nil {
		return nil, err
	}
	if email.Valid {
		user.Email = email.String
	}
	if displayName.Valid {
		user.DisplayName = displayName.String
	}
	if roleName.Valid {
		user.RoleName = roleName.String
	}
	return &user, nil
}

func (s *AuthService) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	var email sql.NullString
	var displayName sql.NullString
	var roleName sql.NullString
	err := s.db.QueryRow(
		`SELECT u.id, u.username, u.role, u.role_id, COALESCE(r.name, '') as role_name, 
		COALESCE(u.display_name, '') as display_name, u.email, u.created_at, u.status 
		FROM users u LEFT JOIN roles r ON u.role_id = r.id WHERE u.username = ?`,
		username,
	).Scan(&user.ID, &user.Username, &user.Role, &user.RoleID, &roleName, &displayName, &email, &user.CreatedAt, &user.Status)
	if err != nil {
		return nil, err
	}
	if email.Valid {
		user.Email = email.String
	}
	if displayName.Valid {
		user.DisplayName = displayName.String
	}
	if roleName.Valid {
		user.RoleName = roleName.String
	}
	return &user, nil
}

type Role struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsPreset    int       `json:"isPreset"`
	Permissions string    `json:"permissions"`
	CreatedAt   time.Time `json:"createdAt"`
}

func (s *AuthService) ListRoles(ctx context.Context) ([]Role, error) {
	rows, err := s.db.Query("SELECT id, name, description, is_preset, permissions, created_at FROM roles ORDER BY is_preset DESC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []Role
	for rows.Next() {
		var r Role
		var description sql.NullString
		if err := rows.Scan(&r.ID, &r.Name, &description, &r.IsPreset, &r.Permissions, &r.CreatedAt); err != nil {
			continue
		}
		if description.Valid {
			r.Description = description.String
		}
		roles = append(roles, r)
	}
	return roles, nil
}

func (s *AuthService) GetRole(ctx context.Context, roleID int64) (*Role, error) {
	var r Role
	var description sql.NullString
	err := s.db.QueryRow(
		"SELECT id, name, description, is_preset, permissions, created_at FROM roles WHERE id = ?",
		roleID,
	).Scan(&r.ID, &r.Name, &description, &r.IsPreset, &r.Permissions, &r.CreatedAt)
	if err != nil {
		return nil, err
	}
	if description.Valid {
		r.Description = description.String
	}
	return &r, nil
}

func (s *AuthService) CreateRole(ctx context.Context, name, description, permissions string) error {
	_, err := s.db.Exec(
		"INSERT INTO roles (name, description, is_preset, permissions) VALUES (?, ?, 0, ?)",
		name, description, permissions,
	)
	return err
}

func (s *AuthService) UpdateRole(ctx context.Context, roleID int64, name, description, permissions string) error {
	_, err := s.db.Exec(
		"UPDATE roles SET name = ?, description = ?, permissions = ? WHERE id = ? AND is_preset = 0",
		name, description, permissions, roleID,
	)
	return err
}

func (s *AuthService) DeleteRole(ctx context.Context, roleID int64) error {
	_, err := s.db.Exec("DELETE FROM roles WHERE id = ? AND is_preset = 0", roleID)
	return err
}

func (s *AuthService) ListAuditLogs(ctx context.Context, userID int64, limit, offset int) ([]AuditLogEntry, error) {
	var rows *sql.Rows
	var err error

	if userID > 0 {
		rows, err = s.db.Query(
			"SELECT id, log_id, user_id, user, action, resource, name, namespace, status, detail, ip, created_at FROM audit_logs WHERE user_id = ? ORDER BY created_at DESC LIMIT ? OFFSET ?",
			userID, limit, offset,
		)
	} else {
		rows, err = s.db.Query(
			"SELECT id, log_id, user_id, user, action, resource, name, namespace, status, detail, ip, created_at FROM audit_logs ORDER BY created_at DESC LIMIT ? OFFSET ?",
			limit, offset,
		)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []AuditLogEntry
	for rows.Next() {
		var l AuditLogEntry
		var userIDVal sql.NullInt64
		var namespace, detail, ip sql.NullString
		if err := rows.Scan(&l.ID, &l.LogID, &userIDVal, &l.Username, &l.Action, &l.Resource, &l.Name, &namespace, &l.Status, &detail, &ip, &l.CreatedAt); err != nil {
			continue
		}
		if userIDVal.Valid {
			l.UserID = userIDVal.Int64
		}
		if namespace.Valid {
			l.Namespace = namespace.String
		}
		if detail.Valid {
			l.Detail = detail.String
		}
		if ip.Valid {
			l.IP = ip.String
		}
		logs = append(logs, l)
	}
	return logs, nil
}

func (s *AuthService) CreateAuditLog(ctx context.Context, userID int64, username, action, resource, name, namespace, status, detail, ip string) error {
	logID := fmt.Sprintf("audit-%d-%d", time.Now().UnixNano(), userID)
	_, err := s.db.Exec(
		"INSERT INTO audit_logs (log_id, user_id, user, action, resource, name, namespace, status, detail, ip) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		logID, userID, username, action, resource, name, namespace, status, detail, ip,
	)
	return err
}

type AuditLogEntry struct {
	ID        int64     `json:"id"`
	LogID     string    `json:"logId"`
	UserID    int64     `json:"userId"`
	Username  string    `json:"username"`
	Action    string    `json:"action"`
	Resource  string    `json:"resource"`
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Status    string    `json:"status"`
	Detail    string    `json:"detail"`
	IP        string    `json:"ip"`
	CreatedAt time.Time `json:"createdAt"`
}

// UserCluster represents a user's cluster access
type UserCluster struct {
	ID          int64     `json:"id"`
	UserID      int64     `json:"userId"`
	ClusterName string    `json:"clusterName"`
	CreatedAt   time.Time `json:"createdAt"`
}

// GetUserClusters returns the list of clusters a user has access to
func (s *AuthService) GetUserClusters(ctx context.Context, userID int64) ([]string, error) {
	rows, err := s.db.Query("SELECT cluster_name FROM user_clusters WHERE user_id = ?", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clusters []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			continue
		}
		clusters = append(clusters, name)
	}
	return clusters, nil
}

// SetUserClusters sets the clusters a user has access to
func (s *AuthService) SetUserClusters(ctx context.Context, userID int64, clusterNames []string) error {
	// Delete existing clusters
	_, err := s.db.Exec("DELETE FROM user_clusters WHERE user_id = ?", userID)
	if err != nil {
		return err
	}

	// Insert new clusters
	for _, name := range clusterNames {
		_, err := s.db.Exec("INSERT INTO user_clusters (user_id, cluster_name) VALUES (?, ?)", userID, name)
		if err != nil {
			return err
		}
	}
	return nil
}

// HasClusterAccess checks if a user has access to a specific cluster
func (s *AuthService) HasClusterAccess(ctx context.Context, userID int64, clusterName string) (bool, error) {
	// Admin users have access to all clusters
	var role string
	err := s.db.QueryRow("SELECT role FROM users WHERE id = ?", userID).Scan(&role)
	if err != nil {
		return false, err
	}
	if role == "admin" {
		return true, nil
	}

	// Check if user has explicit cluster access
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM user_clusters WHERE user_id = ? AND cluster_name = ?", userID, clusterName).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetUserPermissions returns the permissions for a user based on their role
func (s *AuthService) GetUserPermissions(ctx context.Context, userID int64) (map[string]interface{}, error) {
	var roleID sql.NullInt64
	err := s.db.QueryRow("SELECT role_id FROM users WHERE id = ?", userID).Scan(&roleID)
	if err != nil {
		return nil, err
	}

	if !roleID.Valid {
		return nil, nil
	}

	var permissionsStr sql.NullString
	err = s.db.QueryRow("SELECT permissions FROM roles WHERE id = ?", roleID.Int64).Scan(&permissionsStr)
	if err != nil {
		return nil, err
	}

	if !permissionsStr.Valid {
		return nil, nil
	}

	var permissions map[string]interface{}
	if err := json.Unmarshal([]byte(permissionsStr.String), &permissions); err != nil {
		return nil, err
	}

	return permissions, nil
}
