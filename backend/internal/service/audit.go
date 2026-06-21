package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	stdlog "log"
	"time"
)

type AuditService struct {
	db *sql.DB
}

type AuditLog struct {
	ID        string                 `json:"id"`
	Timestamp time.Time              `json:"timestamp"`
	User      string                 `json:"user"`
	Action    string                 `json:"action"`
	Resource  string                 `json:"resource"`
	Name      string                 `json:"name"`
	Namespace string                 `json:"namespace,omitempty"`
	Status    string                 `json:"status"`
	Detail    map[string]interface{} `json:"detail,omitempty"`
	IP        string                 `json:"ip,omitempty"`
	UserAgent string                 `json:"userAgent,omitempty"`
}

type AuditLogQuery struct {
	User      string `json:"user,omitempty"`
	Action    string `json:"action,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Namespace string `json:"namespace,omitempty"`
	Status    string `json:"status,omitempty"`
	StartTime string `json:"startTime,omitempty"`
	EndTime   string `json:"endTime,omitempty"`
	Keyword   string `json:"keyword,omitempty"`
	Page      int    `json:"page"`
	PageSize  int    `json:"pageSize"`
}

type AuditLogStats struct {
	Total       int            `json:"total"`
	ByAction    map[string]int `json:"byAction"`
	ByResource  map[string]int `json:"byResource"`
	ByStatus    map[string]int `json:"byStatus"`
	ByUser      map[string]int `json:"byUser"`
	RecentCount int            `json:"recentCount"`
}

func NewAuditService(db *sql.DB) *AuditService {
	return &AuditService{db: db}
}

func (s *AuditService) Record(log AuditLog) {
	if log.ID == "" {
		log.ID = fmt.Sprintf("audit-%d", time.Now().UnixNano())
	}
	if log.Timestamp.IsZero() {
		log.Timestamp = time.Now()
	}

	detailJSON, _ := json.Marshal(log.Detail)

	_, err := s.db.Exec(
		`INSERT INTO audit_logs (log_id, user, action, resource, name, namespace, status, detail, ip, user_agent, created_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		log.ID, log.User, log.Action, log.Resource, log.Name, log.Namespace,
		log.Status, string(detailJSON), log.IP, log.UserAgent, log.Timestamp,
	)
	if err != nil {
		stdlog.Printf("Failed to record audit log: %v", err)
	}
}

func (s *AuditService) Query(ctx context.Context, query AuditLogQuery) ([]AuditLog, int) {
	if query.Page <= 0 {
		query.Page = 1
	}
	if query.PageSize <= 0 {
		query.PageSize = 20
	}

	where := "1=1"
	args := []interface{}{}

	if query.User != "" {
		where += " AND user = ?"
		args = append(args, query.User)
	}
	if query.Action != "" {
		where += " AND action = ?"
		args = append(args, query.Action)
	}
	if query.Resource != "" {
		where += " AND resource = ?"
		args = append(args, query.Resource)
	}
	if query.Namespace != "" {
		where += " AND namespace = ?"
		args = append(args, query.Namespace)
	}
	if query.Status != "" {
		where += " AND status = ?"
		args = append(args, query.Status)
	}
	if query.StartTime != "" {
		where += " AND created_at >= ?"
		args = append(args, query.StartTime)
	}
	if query.EndTime != "" {
		where += " AND created_at <= ?"
		args = append(args, query.EndTime)
	}
	if query.Keyword != "" {
		where += " AND (user LIKE ? OR action LIKE ? OR resource LIKE ? OR name LIKE ?)"
		keyword := "%" + query.Keyword + "%"
		args = append(args, keyword, keyword, keyword, keyword)
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs WHERE %s", where)
	s.db.QueryRow(countQuery, args...).Scan(&total)

	offset := (query.Page - 1) * query.PageSize
	selectQuery := fmt.Sprintf(
		"SELECT log_id, user, action, resource, name, namespace, status, detail, ip, user_agent, created_at FROM audit_logs WHERE %s ORDER BY created_at DESC LIMIT ? OFFSET ?",
		where,
	)
	args = append(args, query.PageSize, offset)

	rows, err := s.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		var detailStr sql.NullString
		var createdAt time.Time

		err := rows.Scan(&log.ID, &log.User, &log.Action, &log.Resource, &log.Name,
			&log.Namespace, &log.Status, &detailStr, &log.IP, &log.UserAgent, &createdAt)
		if err != nil {
			continue
		}

		log.Timestamp = createdAt
		if detailStr.Valid && detailStr.String != "" {
			json.Unmarshal([]byte(detailStr.String), &log.Detail)
		}
		logs = append(logs, log)
	}

	return logs, total
}

func (s *AuditService) GetByID(ctx context.Context, id string) (*AuditLog, error) {
	var log AuditLog
	var detailStr sql.NullString
	var createdAt time.Time

	err := s.db.QueryRow(
		"SELECT log_id, user, action, resource, name, namespace, status, detail, ip, user_agent, created_at FROM audit_logs WHERE log_id = ?",
		id,
	).Scan(&log.ID, &log.User, &log.Action, &log.Resource, &log.Name,
		&log.Namespace, &log.Status, &detailStr, &log.IP, &log.UserAgent, &createdAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("audit log not found: %s", id)
	}
	if err != nil {
		return nil, err
	}

	log.Timestamp = createdAt
	if detailStr.Valid && detailStr.String != "" {
		json.Unmarshal([]byte(detailStr.String), &log.Detail)
	}
	return &log, nil
}

func (s *AuditService) GetStats(ctx context.Context) *AuditLogStats {
	stats := &AuditLogStats{
		ByAction:   make(map[string]int),
		ByResource: make(map[string]int),
		ByStatus:   make(map[string]int),
		ByUser:     make(map[string]int),
	}

	s.db.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&stats.Total)
	s.db.QueryRow("SELECT COUNT(*) FROM audit_logs WHERE created_at >= datetime('now', '-1 hour')").Scan(&stats.RecentCount)

	rows, _ := s.db.Query("SELECT action, COUNT(*) FROM audit_logs GROUP BY action")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var action string
			var count int
			rows.Scan(&action, &count)
			stats.ByAction[action] = count
		}
	}

	rows, _ = s.db.Query("SELECT resource, COUNT(*) FROM audit_logs GROUP BY resource")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var resource string
			var count int
			rows.Scan(&resource, &count)
			stats.ByResource[resource] = count
		}
	}

	rows, _ = s.db.Query("SELECT status, COUNT(*) FROM audit_logs GROUP BY status")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			rows.Scan(&status, &count)
			stats.ByStatus[status] = count
		}
	}

	rows, _ = s.db.Query("SELECT user, COUNT(*) FROM audit_logs GROUP BY user")
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var user string
			var count int
			rows.Scan(&user, &count)
			stats.ByUser[user] = count
		}
	}

	return stats
}

func (s *AuditService) Export(ctx context.Context, query AuditLogQuery, format string) ([]byte, string, error) {
	logs, _ := s.Query(ctx, query)

	switch format {
	case "csv":
		return s.exportCSV(logs)
	case "json":
		data, err := json.MarshalIndent(logs, "", "  ")
		return data, "application/json", err
	default:
		return nil, "", fmt.Errorf("unsupported format: %s", format)
	}
}

func (s *AuditService) exportCSV(logs []AuditLog) ([]byte, string, error) {
	header := "ID,Timestamp,User,Action,Resource,Name,Namespace,Status,IP\n"
	var rows string
	for _, log := range logs {
		row := fmt.Sprintf("%s,%s,%s,%s,%s,%s,%s,%s,%s\n",
			log.ID,
			log.Timestamp.Format(time.RFC3339),
			log.User,
			log.Action,
			log.Resource,
			log.Name,
			log.Namespace,
			log.Status,
			log.IP,
		)
		rows += row
	}
	return []byte(header + rows), "text/csv", nil
}

func (s *AuditService) Cleanup(ctx context.Context, days int) (int, error) {
	result, err := s.db.Exec(
		"DELETE FROM audit_logs WHERE created_at < datetime('now', ?)",
		fmt.Sprintf("-%d days", days),
	)
	if err != nil {
		return 0, err
	}
	affected, _ := result.RowsAffected()
	return int(affected), nil
}
