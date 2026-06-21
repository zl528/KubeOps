package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

var (
	globalDB   *sql.DB
	globalOnce sync.Once
)

type Database struct {
	DB *sql.DB
}

func GetDB() *Database {
	globalOnce.Do(func() {
		dbPath := os.Getenv("DB_PATH")
		if dbPath == "" {
			home := os.Getenv("HOME")
			if home == "" {
				home = "/tmp"
			}
			dbPath = filepath.Join(home, ".kubeops", "kubeops.db")
		}

		dir := filepath.Dir(dbPath)
		os.MkdirAll(dir, 0755)

		var err error
		globalDB, err = sql.Open("sqlite", dbPath+"?_journal_mode=WAL&_busy_timeout=5000")
		if err != nil {
			panic(fmt.Sprintf("failed to open database: %v", err))
		}

		globalDB.SetMaxOpenConns(1)
		globalDB.SetMaxIdleConns(1)

		initTables(globalDB)
	})

	return &Database{DB: globalDB}
}

func initTables(db *sql.DB) {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS roles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			description TEXT,
			is_preset INTEGER DEFAULT 0,
			permissions TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			password TEXT NOT NULL,
			role TEXT DEFAULT 'user',
			role_id INTEGER,
			display_name TEXT,
			email TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME,
			status INTEGER DEFAULT 1,
			FOREIGN KEY (role_id) REFERENCES roles(id)
		)`,
		`CREATE TABLE IF NOT EXISTS user_clusters (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			cluster_name TEXT NOT NULL,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, cluster_name),
			FOREIGN KEY (user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE IF NOT EXISTS audit_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			log_id TEXT UNIQUE NOT NULL,
			user_id INTEGER,
			user TEXT,
			action TEXT,
			resource TEXT,
			name TEXT,
			namespace TEXT,
			status TEXT,
			detail TEXT,
			ip TEXT,
			user_agent TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alert_rules (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			rule_id TEXT UNIQUE NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			metric TEXT,
			condition_type TEXT,
			threshold REAL,
			severity TEXT,
			enabled INTEGER DEFAULT 1,
			namespace TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS alert_history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			alert_id TEXT UNIQUE NOT NULL,
			rule_id TEXT,
			rule_name TEXT,
			severity TEXT,
			message TEXT,
			status TEXT,
			value REAL,
			threshold REAL,
			fired_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			resolved_at DATETIME
		)`,
		`CREATE TABLE IF NOT EXISTS system_config (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			key TEXT UNIQUE NOT NULL,
			value TEXT,
			description TEXT,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS cluster_connections (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT UNIQUE NOT NULL,
			server TEXT,
			token TEXT,
			kubeconfig TEXT,
			status TEXT DEFAULT 'disconnected',
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
	}

	for _, q := range queries {
		if _, err := db.Exec(q); err != nil {
			panic(fmt.Sprintf("failed to create table: %v", err))
		}
	}

	// Insert default admin user if not exists
	var count int
	db.QueryRow("SELECT COUNT(*) FROM users WHERE username = 'admin'").Scan(&count)
	if count == 0 {
		db.Exec("INSERT INTO users (username, password, role) VALUES (?, ?, ?)",
			"admin", "$2a$10$AHbY6RCweg9jMEju87wTae2Hme1oS/joNK1ZYdJHz5lFvh8nQxUc6", "admin")
	}

	// Insert preset roles if not exists
	var roleCount int
	db.QueryRow("SELECT COUNT(*) FROM roles WHERE is_preset = 1").Scan(&roleCount)
	if roleCount == 0 {
		presetRoles := []struct {
			name        string
			description string
			permissions string
		}{
			{
				name:        "管理员",
				description: "系统管理员，拥有所有权限",
				permissions: `{"modules":{"workloads":{"view":true,"create":true,"edit":true,"delete":true},"network":{"view":true,"create":true,"edit":true,"delete":true},"storage":{"view":true,"create":true,"edit":true,"delete":true},"rbac":{"view":true,"create":true,"edit":true,"delete":true},"usercenter":{"view":true,"create":true,"edit":true,"delete":true}}}`,
			},
			{
				name:        "开发者",
				description: "开发人员，可管理工作负载，其他模块只读",
				permissions: `{"modules":{"workloads":{"view":true,"create":true,"edit":true,"delete":true},"network":{"view":true,"create":false,"edit":false,"delete":false},"storage":{"view":true,"create":false,"edit":false,"delete":false},"rbac":{"view":false,"create":false,"edit":false,"delete":false},"usercenter":{"view":true,"create":false,"edit":true,"delete":false}}}`,
			},
			{
				name:        "只读用户",
				description: "只读用户，仅可查看资源",
				permissions: `{"modules":{"workloads":{"view":true,"create":false,"edit":false,"delete":false},"network":{"view":true,"create":false,"edit":false,"delete":false},"storage":{"view":true,"create":false,"edit":false,"delete":false},"rbac":{"view":false,"create":false,"edit":false,"delete":false},"usercenter":{"view":true,"create":false,"edit":true,"delete":false}}}`,
			},
		}
		for _, r := range presetRoles {
			db.Exec("INSERT INTO roles (name, description, is_preset, permissions) VALUES (?, ?, 1, ?)",
				r.name, r.description, r.permissions)
		}
	}

	// Ensure admin user has role_id=1 (管理员)
	db.Exec("UPDATE users SET role_id = 1 WHERE username = 'admin' AND (role_id IS NULL OR role_id = 0)")
}

func (d *Database) Close() {
	if globalDB != nil {
		globalDB.Close()
	}
}
