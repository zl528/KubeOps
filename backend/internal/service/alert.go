package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"k8s.io/client-go/kubernetes"
)

type AlertService struct {
	client kubernetes.Interface
	db     *sql.DB
}

func NewAlertService(client kubernetes.Interface, db *sql.DB) *AlertService {
	svc := &AlertService{
		client: client,
		db:     db,
	}
	svc.initDefaultRules()
	return svc
}

type AlertRule struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Metric      string            `json:"metric"`
	Condition   string            `json:"condition"`
	Threshold   float64           `json:"threshold"`
	Severity    string            `json:"severity"`
	Enabled     bool              `json:"enabled"`
	Namespace   string            `json:"namespace,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
}

type AlertHistory struct {
	ID         string     `json:"id"`
	RuleID     string     `json:"ruleId"`
	RuleName   string     `json:"ruleName"`
	Severity   string     `json:"severity"`
	Message    string     `json:"message"`
	Status     string     `json:"status"`
	Value      float64    `json:"value"`
	Threshold  float64    `json:"threshold"`
	FiredAt    time.Time  `json:"firedAt"`
	ResolvedAt *time.Time `json:"resolvedAt,omitempty"`
}

func (s *AlertService) initDefaultRules() {
	// Check if rules already exist
	var count int
	s.db.QueryRow("SELECT COUNT(*) FROM alert_rules").Scan(&count)
	if count > 0 {
		return
	}

	defaults := []struct {
		ruleID      string
		name        string
		description string
		metric      string
		condition   string
		threshold   float64
		severity    string
	}{
		{"high-cpu", "CPU 使用率过高", "节点 CPU 使用率超过 80%", "cpu_usage", "greater_than", 80, "warning"},
		{"high-memory", "内存使用率过高", "节点内存使用率超过 85%", "memory_usage", "greater_than", 85, "warning"},
		{"high-disk", "磁盘使用率过高", "节点磁盘使用率超过 90%", "disk_usage", "greater_than", 90, "critical"},
		{"pod-crashlooping", "Pod CrashLoopBackOff", "Pod 处于 CrashLoopBackOff 状态", "pod_status", "equals", 0, "critical"},
		{"pod-pending", "Pod 长时间 Pending", "Pod 处于 Pending 状态超过 5 分钟", "pod_status", "equals", 0, "warning"},
	}

	for _, d := range defaults {
		s.db.Exec(`INSERT OR IGNORE INTO alert_rules (rule_id, name, description, metric, condition_type, threshold, severity, enabled) 
			VALUES (?, ?, ?, ?, ?, ?, ?, 1)`,
			d.ruleID, d.name, d.description, d.metric, d.condition, d.threshold, d.severity)
	}
}

func (s *AlertService) ListRules(ctx context.Context) ([]AlertRule, error) {
	rows, err := s.db.Query("SELECT rule_id, name, description, metric, condition_type, threshold, severity, enabled, created_at, updated_at FROM alert_rules ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rules []AlertRule
	for rows.Next() {
		var r AlertRule
		var enabled int
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Metric, &r.Condition, &r.Threshold, &r.Severity, &enabled, &r.CreatedAt, &r.UpdatedAt); err != nil {
			continue
		}
		r.Enabled = enabled == 1
		rules = append(rules, r)
	}
	return rules, nil
}

func (s *AlertService) CreateRule(ctx context.Context, rule AlertRule) error {
	_, err := s.db.Exec(`INSERT INTO alert_rules (rule_id, name, description, metric, condition_type, threshold, severity, enabled) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		rule.ID, rule.Name, rule.Description, rule.Metric, rule.Condition, rule.Threshold, rule.Severity, boolToInt(rule.Enabled))
	return err
}

func (s *AlertService) UpdateRule(ctx context.Context, rule AlertRule) error {
	_, err := s.db.Exec(`UPDATE alert_rules SET name=?, description=?, metric=?, condition_type=?, threshold=?, severity=?, updated_at=CURRENT_TIMESTAMP 
		WHERE rule_id=?`,
		rule.Name, rule.Description, rule.Metric, rule.Condition, rule.Threshold, rule.Severity, rule.ID)
	return err
}

func (s *AlertService) DeleteRule(ctx context.Context, ruleID string) error {
	_, err := s.db.Exec("DELETE FROM alert_rules WHERE rule_id = ?", ruleID)
	return err
}

func (s *AlertService) ToggleRule(ctx context.Context, ruleID string, enabled bool) error {
	_, err := s.db.Exec("UPDATE alert_rules SET enabled = ?, updated_at = CURRENT_TIMESTAMP WHERE rule_id = ?", boolToInt(enabled), ruleID)
	return err
}

func (s *AlertService) ListHistory(ctx context.Context, severity string) ([]AlertHistory, error) {
	query := "SELECT alert_id, rule_id, rule_name, severity, message, status, value, threshold, fired_at, resolved_at FROM alert_history"
	args := []interface{}{}
	if severity != "" {
		query += " WHERE severity = ?"
		args = append(args, severity)
	}
	query += " ORDER BY fired_at DESC LIMIT 100"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []AlertHistory
	for rows.Next() {
		var h AlertHistory
		var resolvedAt sql.NullTime
		if err := rows.Scan(&h.ID, &h.RuleID, &h.RuleName, &h.Severity, &h.Message, &h.Status, &h.Value, &h.Threshold, &h.FiredAt, &resolvedAt); err != nil {
			continue
		}
		if resolvedAt.Valid {
			h.ResolvedAt = &resolvedAt.Time
		}
		history = append(history, h)
	}
	return history, nil
}

func (s *AlertService) ResolveAlert(ctx context.Context, alertID string) error {
	_, err := s.db.Exec("UPDATE alert_history SET status = 'resolved', resolved_at = CURRENT_TIMESTAMP WHERE alert_id = ?", alertID)
	return err
}

func (s *AlertService) DeleteAlert(ctx context.Context, alertID string) error {
	_, err := s.db.Exec("DELETE FROM alert_history WHERE alert_id = ?", alertID)
	return err
}

func (s *AlertService) GetStats(ctx context.Context) (map[string]int, error) {
	var total, firing, resolved int
	s.db.QueryRow("SELECT COUNT(*) FROM alert_history").Scan(&total)
	s.db.QueryRow("SELECT COUNT(*) FROM alert_history WHERE status = 'firing'").Scan(&firing)
	s.db.QueryRow("SELECT COUNT(*) FROM alert_history WHERE status = 'resolved'").Scan(&resolved)

	return map[string]int{
		"total":    total,
		"firing":   firing,
		"resolved": resolved,
	}, nil
}

func (s *AlertService) FireAlert(ctx context.Context, ruleID, ruleName, severity, message string, value, threshold float64) error {
	alertID := fmt.Sprintf("alert-%d", time.Now().UnixNano())
	_, err := s.db.Exec(`INSERT INTO alert_history (alert_id, rule_id, rule_name, severity, message, status, value, threshold, fired_at)
		VALUES (?, ?, ?, ?, ?, 'firing', ?, ?, CURRENT_TIMESTAMP)`,
		alertID, ruleID, ruleName, severity, message, value, threshold)
	return err
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

func (s *AlertService) GetRule(ctx context.Context, ruleID string) (*AlertRule, error) {
	var r AlertRule
	var enabled int
	err := s.db.QueryRow("SELECT rule_id, name, description, metric, condition_type, threshold, severity, enabled, created_at, updated_at FROM alert_rules WHERE rule_id = ?", ruleID).
		Scan(&r.ID, &r.Name, &r.Description, &r.Metric, &r.Condition, &r.Threshold, &r.Severity, &enabled, &r.CreatedAt, &r.UpdatedAt)
	if err != nil {
		return nil, err
	}
	r.Enabled = enabled == 1
	return &r, nil
}

func (s *AlertService) EnableRule(ctx context.Context, ruleID string, enabled bool) error {
	return s.ToggleRule(ctx, ruleID, enabled)
}

func (s *AlertService) ListAlerts(ctx context.Context, severity string) ([]AlertHistory, error) {
	return s.ListHistory(ctx, severity)
}

func (s *AlertService) GetAlert(ctx context.Context, alertID string) (*AlertHistory, error) {
	var h AlertHistory
	var resolvedAt sql.NullTime
	err := s.db.QueryRow("SELECT alert_id, rule_id, rule_name, severity, message, status, value, threshold, fired_at, resolved_at FROM alert_history WHERE alert_id = ?", alertID).
		Scan(&h.ID, &h.RuleID, &h.RuleName, &h.Severity, &h.Message, &h.Status, &h.Value, &h.Threshold, &h.FiredAt, &resolvedAt)
	if err != nil {
		return nil, err
	}
	if resolvedAt.Valid {
		h.ResolvedAt = &resolvedAt.Time
	}
	return &h, nil
}

func (s *AlertService) GetAlertStats(ctx context.Context) (map[string]int, error) {
	return s.GetStats(ctx)
}

type NotificationConfig struct {
	ID       string   `json:"id"`
	Type     string   `json:"type"`
	Endpoint string   `json:"endpoint"`
	Enabled  bool     `json:"enabled"`
	Events   []string `json:"events"`
}

func (s *AlertService) ListNotifications(ctx context.Context) ([]NotificationConfig, error) {
	return []NotificationConfig{}, nil
}

func (s *AlertService) CreateNotification(ctx context.Context, config NotificationConfig) error {
	return nil
}

func (s *AlertService) DeleteNotification(ctx context.Context, id string) error {
	return nil
}
