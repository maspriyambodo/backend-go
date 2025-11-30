package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"adminbe/internal/app/models"

	"github.com/gin-gonic/gin"
)

// listAuditLogsHandler GET /api/audit_logs
func listAuditLogsHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, user_id, event_type, table_name, record_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs ORDER BY created_at DESC")
		if err != nil {
			log.Printf("Error querying audit logs: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
			return
		}
		defer rows.Close()

		var logs []models.AuditLog
		for rows.Next() {
			var a models.AuditLog
			if err := rows.Scan(&a.ID, &a.UserID, &a.EventType, &a.TableName, &a.RecordID, &a.OldValues, &a.NewValues, &a.IPAddress, &a.UserAgent, &a.CreatedAt); err != nil {
				log.Printf("Error scanning audit log row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve audit logs"})
				return
			}
			logs = append(logs, a)
		}
		c.JSON(http.StatusOK, gin.H{"data": logs})
	}
}

// getAuditLogHandler GET /api/audit_logs/:id
func getAuditLogHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		aID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var a models.AuditLog
		row := db.QueryRow("SELECT id, user_id, event_type, table_name, record_id, old_values, new_values, ip_address, user_agent, created_at FROM audit_logs WHERE id = ?", aID)
		err = row.Scan(&a.ID, &a.UserID, &a.EventType, &a.TableName, &a.RecordID, &a.OldValues, &a.NewValues, &a.IPAddress, &a.UserAgent, &a.CreatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Audit log not found"})
			return
		} else if err != nil {
			log.Printf("Error querying audit log: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": a})
	}
}

// createAuditLogHandler POST /api/audit_logs
func createAuditLogHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req struct {
			UserID    *uint64     `json:"user_id"`
			EventType string      `json:"event_type" binding:"required"`
			TableName string      `json:"table_name" binding:"required"`
			RecordID  uint64      `json:"record_id" binding:"required"`
			OldValues interface{} `json:"old_values"`
			NewValues interface{} `json:"new_values"`
			IPAddress string      `json:"ip_address"`
			UserAgent *string     `json:"user_agent"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var oldJSON, newJSON []byte
		if req.OldValues != nil {
			oldJSON, _ = json.Marshal(req.OldValues)
		}
		if req.NewValues != nil {
			newJSON, _ = json.Marshal(req.NewValues)
		}

		result, err := db.Exec("INSERT INTO audit_logs (user_id, event_type, table_name, record_id, old_values, new_values, ip_address, user_agent) VALUES (?, ?, ?, ?, ?, ?, INET6_ATON(?), ?)",
			req.UserID, req.EventType, req.TableName, req.RecordID, oldJSON, newJSON, req.IPAddress, req.UserAgent)
		if err != nil {
			log.Printf("Error inserting audit log: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create audit log"})
			return
		}

		logID, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"message": "Audit log created", "id": logID})
	}
}

// updateAuditLogHandler PUT /api/audit_logs/:id (not recommended, but for CRUD)
func updateAuditLogHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Update not allowed for audit logs"})
	}
}

// deleteAuditLogHandler DELETE /api/audit_logs/:id (not recommended, but for CRUD)
func deleteAuditLogHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Delete not allowed for audit logs"})
	}
}
