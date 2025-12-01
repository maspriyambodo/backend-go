package database

import (
	"database/sql"
	"sync"
)

// PreparedStmts holds cached prepared statements to reduce compilation overhead
type PreparedStmts struct {
	mu     sync.RWMutex
	stmts  map[string]*sql.Stmt
	db     *sql.DB
	closed bool
}

// NewPreparedStmts creates a new prepared statements cache
func NewPreparedStmts(db *sql.DB) *PreparedStmts {
	ps := &PreparedStmts{
		stmts:  make(map[string]*sql.Stmt),
		db:     db,
		closed: false,
	}

	// Pre-prepare commonly used queries
	queries := []string{
		// Users
		"SELECT id, username, email, status, created_at, updated_at, deleted_at, deleted_by FROM users WHERE deleted_at IS NULL",
		"SELECT id, username, email, status, created_at, updated_at, deleted_at, deleted_by FROM users WHERE id = ? AND deleted_at IS NULL",
		"SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL",
		"INSERT INTO users (username, email, password_hash, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		"UPDATE users SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL",

		// Audit logs
		"SELECT id, user_id, event_type, table_name, record_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE deleted_at IS NULL",
		"INSERT INTO audit_logs (user_id, event_type, table_name, record_id, old_values, new_values) VALUES (?, ?, ?, ?, ?, ?)",

		// Roles
		"SELECT id, name, description, created_at, updated_at, deleted_at, deleted_by FROM roles WHERE deleted_at IS NULL",
		"SELECT id, name, description, created_at, updated_at, deleted_at, deleted_by FROM roles WHERE id = ? AND deleted_at IS NULL",
		"INSERT INTO roles (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)",
	}

	for _, query := range queries {
		if stmt, err := db.Prepare(query); err == nil {
			ps.stmts[query] = stmt
		}
	}

	return ps
}

// Get gets or creates a prepared statement
func (ps *PreparedStmts) Get(query string) *sql.Stmt {
	ps.mu.RLock()
	stmt, exists := ps.stmts[query]
	ps.mu.RUnlock()

	if exists && !ps.closed {
		return stmt
	}

	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return nil
	}

	// Double-check after acquiring write lock
	if stmt, exists := ps.stmts[query]; exists {
		return stmt
	}

	// Prepare new statement
	if stmt, err := ps.db.Prepare(query); err == nil {
		ps.stmts[query] = stmt
		return stmt
	}

	return nil
}

// Close closes all prepared statements
func (ps *PreparedStmts) Close() error {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ps.closed = true

	var lastErr error
	for _, stmt := range ps.stmts {
		if err := stmt.Close(); err != nil {
			lastErr = err
		}
	}

	// Clear the map
	clear(ps.stmts)

	return lastErr
}
