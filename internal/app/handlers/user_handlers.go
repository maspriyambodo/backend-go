package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"strconv"
	"sync"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/app/services"

	"github.com/gin-gonic/gin"
)

// Per-goroutine pools to eliminate concurrency bottlenecks
var (
	responseMapPoolMu sync.RWMutex
	responseMapPools  = make(map[string]*sync.Pool) // Per-goroutine pools indexed by goroutine ID

	paginationMapPoolMu sync.RWMutex
	paginationMapPools  = make(map[string]*sync.Pool)

	userSlicePoolMu sync.RWMutex
	userSlicePools  = make(map[string]*sync.Pool)

	// 🔧 OPTIMIZED: Worker pool for audit logging (3 workers)
	numAuditWorkers = 3
	auditLogChan    = make(chan auditLogEntry, 2000)  // Increased buffer
	auditBatchChan  = make(chan []auditLogEntry, 100) // For batched processing
	auditStopCh     = make(chan struct{})
	auditWorkerWG   sync.WaitGroup
)

// AuditPriority represents different priorities for audit log processing
type AuditPriority int

const (
	PriorityLow AuditPriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// auditLogEntry represents an audit log entry for async processing
type auditLogEntry struct {
	UserID    *uint64       `json:"user_id,omitempty"`
	Event     string        `json:"event"`
	Table     string        `json:"table"`
	RecordID  uint64        `json:"record_id"`
	OldValues interface{}   `json:"old_values,omitempty"`
	NewValues interface{}   `json:"new_values,omitempty"`
	DB        *sql.DB       `json:"-"` // DB connection (not serialized)
	Priority  AuditPriority `json:"priority"`
	Timestamp time.Time     `json:"-"`
}

// StartAuditLogger starts the optimized worker pool for audit logging
func StartAuditLogger() {
	// ✅ RECOMMENDATION 1: Worker Pool Pattern
	for i := 0; i < numAuditWorkers; i++ {
		auditWorkerWG.Add(1)
		go auditWorker(i)
	}

	// Separate batch processor worker
	auditWorkerWG.Add(1)
	go auditBatchWorker()
}

// StopAuditLogger stops all audit logging workers gracefully
func StopAuditLogger() {
	close(auditStopCh)
	auditWorkerWG.Wait() // Wait for all workers to finish
}

// auditWorker handles individual audit log entries with timeout protection
func auditWorker(workerID int) {
	defer auditWorkerWG.Done()

	batch := make([]auditLogEntry, 0, 10)               // Batch up to 10 entries for efficiency
	batchTimer := time.NewTimer(100 * time.Millisecond) // Max wait time for batch
	defer batchTimer.Stop()

	for {
		select {
		case entry := <-auditLogChan:
			batch = append(batch, entry)

			// ✅ RECOMMENDATION 2: Batching for Reduced DB Round Trips
			if len(batch) >= 10 {
				processAuditBatch(batch[:len(batch)]) // Process current batch
				batch = batch[:0]                     // Reset batch
				batchTimer.Reset(100 * time.Millisecond)
			}

		case <-batchTimer.C:
			// Timeout: process any accumulated batch
			if len(batch) > 0 {
				processAuditBatch(batch[:len(batch)])
				batch = batch[:0]
			}
			batchTimer.Reset(100 * time.Millisecond)

		case batchEntries := <-auditBatchChan:
			// Direct batch processing request
			processAuditBatch(batchEntries)

		case <-auditStopCh:
			// ✅ RECOMMENDATION 3: Graceful Shutdown - Process Remaining Work
			if len(batch) > 0 {
				processAuditBatch(batch[:len(batch)])
			}
			return
		}
	}
}

// auditBatchWorker handles pre-batched audit log entries
func auditBatchWorker() {
	defer auditWorkerWG.Done()

	for {
		select {
		case batch := <-auditBatchChan:
			processAuditBatch(batch)
		case <-auditStopCh:
			return
		}
	}
}

// processAuditLog processes an audit log entry synchronously but in background
func processAuditLog(entry auditLogEntry) {
	var oldJSON, newJSON []byte
	if entry.OldValues != nil {
		oldJSON, _ = json.Marshal(entry.OldValues)
	}
	if entry.NewValues != nil {
		newJSON, _ = json.Marshal(entry.NewValues)
	}

	// Execute synchronously but outside of request handler
	entry.DB.Exec("INSERT INTO audit_logs (user_id, event_type, table_name, record_id, old_values, new_values) VALUES (?, ?, ?, ?, ?, ?)",
		entry.UserID, entry.Event, entry.Table, entry.RecordID, oldJSON, newJSON)
}

// processAuditBatch processes multiple audit log entries in optimized batches
func processAuditBatch(entries []auditLogEntry) {
	if len(entries) == 0 {
		return
	}

	// Get one DB connection for the batch (assuming first entry's DB)
	db := entries[0].DB

	// ✅ RECOMMENDATION 4: Use transaction for batch inserts
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Failed to start audit batch transaction: %v", err)
		// Fall back to individual processing
		for _, entry := range entries {
			processAuditLog(entry)
		}
		return
	}
	defer tx.Rollback() // Will be ignored if committed

	// Prepare statement once for the batch
	stmt, err := tx.Prepare("INSERT INTO audit_logs (user_id, event_type, table_name, record_id, old_values, new_values) VALUES (?, ?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("Failed to prepare audit batch statement: %v", err)
		// Fall back to individual processing
		for _, entry := range entries {
			processAuditLog(entry)
		}
		return
	}
	defer stmt.Close()

	// Execute batch inserts
	for _, entry := range entries {
		var oldJSON, newJSON []byte
		if entry.OldValues != nil {
			oldJSON, _ = json.Marshal(entry.OldValues)
		}
		if entry.NewValues != nil {
			newJSON, _ = json.Marshal(entry.NewValues)
		}

		_, err = stmt.Exec(entry.UserID, entry.Event, entry.Table, entry.RecordID, oldJSON, newJSON)
		if err != nil {
			log.Printf("Failed to execute batch audit insert: %v", err)
			// Continue with other entries - don't fail the whole batch
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		log.Printf("Failed to commit audit batch transaction: %v", err)
		// Transaction will rollback automatically due to defer
	}
}

// parseIntMinMax parses a string to int with min/max bounds
func parseIntMinMax(s string, defaultVal, min, max int) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// isNotFoundError checks if error is a public not found error
func isNotFoundError(err error) bool {
	var ginErr *gin.Error
	if errors.As(err, &ginErr) && ginErr.Type == gin.ErrorTypePublic {
		return true
	}
	return false
}

// listUsersHandler GET /api/users
func listUsersHandler(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "50")

		page := parseIntMinMax(pageStr, 1, 1, 10000)
		limit := parseIntMinMax(limitStr, 50, 1, 1000)

		result, err := userService.ListUsers(page, limit)
		if err != nil {
			log.Printf("Error listing users: %v", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve users"})
			return
		}

		c.JSON(200, result)
	}
}

// getUserHandler GET /api/users/:id
func getUserHandler(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		user, err := userService.GetUser(id)
		if err != nil {
			if isNotFoundError(err) {
				c.JSON(404, gin.H{"error": "User not found"})
				return
			}
			log.Printf("Error getting user: %v", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve user"})
			return
		}
		c.JSON(200, gin.H{"data": user})
	}
}

// createUserHandler POST /api/users
func createUserHandler(userService services.UserService, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		user, err := userService.CreateUser(req)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			c.JSON(500, gin.H{"error": "Failed to create user"})
			return
		}

		c.JSON(201, gin.H{"message": "User created", "data": user})

		// Audit logging
		if auditLogChan != nil {
			select {
			case auditLogChan <- auditLogEntry{
				UserID:    nil,
				Event:     "CREATE",
				Table:     "users",
				RecordID:  user.ID,
				OldValues: nil,
				NewValues: req,
				DB:        db,
			}:
			default:
				log.Printf("Warning: audit log queue full, dropping CREATE audit for user %d", user.ID)
			}
		}
	}
}

// updateUserHandler PUT /api/users/:id
func updateUserHandler(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req models.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}

		user, err := userService.UpdateUser(id, req)
		if err != nil {
			if isNotFoundError(err) {
				c.JSON(404, gin.H{"error": "User not found"})
				return
			}
			log.Printf("Error updating user: %v", err)
			c.JSON(500, gin.H{"error": "Failed to update user"})
			return
		}

		c.JSON(200, gin.H{"message": "User updated", "data": user})

		// Audit logging would go here, but we need DB
	}
}

// deleteUserHandler DELETE /api/users/:id
func deleteUserHandler(userService services.UserService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		err := userService.DeleteUser(id)
		if err != nil {
			if isNotFoundError(err) {
				c.JSON(404, gin.H{"error": "User not found"})
				return
			}
			log.Printf("Error deleting user: %v", err)
			c.JSON(500, gin.H{"error": "Failed to delete user"})
			return
		}

		c.JSON(200, gin.H{"message": "User deleted"})
	}
}
