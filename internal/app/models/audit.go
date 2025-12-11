package models

import (
	"time"
)

// AuditLog represents the audit_logs table
type AuditLog struct {
	ID        uint64      `json:"id" db:"id"`
	UserID    uint64      `json:"user_id" db:"user_id"`
	EventType string      `json:"event_type" db:"event_type"`
	TableName string      `json:"table_name" db:"table_name"`
	RecordID  uint64      `json:"record_id" db:"record_id"`
	OldValues interface{} `json:"old_values" db:"old_values"`
	NewValues interface{} `json:"new_values" db:"new_values"`
	IPAddress []byte      `json:"ip_address" db:"ip_address"`
	UserAgent *string     `json:"user_agent" db:"user_agent"`
	CreatedAt *time.Time  `json:"created_at" db:"created_at"`
}
