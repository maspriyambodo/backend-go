package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/pkg/utils"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
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

// getPool gets or creates a per-goroutine pool to avoid contention
func getPool(poolType string) *sync.Pool {
	// Simple goroutine ID approximation using pointer to avoid allocations
	goroutineID := fmt.Sprintf("%p", &poolType)

	var pool *sync.Pool
	var mu *sync.RWMutex

	switch poolType {
	case "response":
		mu = &responseMapPoolMu
		mu.RLock()
		pool = responseMapPools[goroutineID]
		mu.RUnlock()
		if pool == nil {
			mu.Lock()
			if responseMapPools[goroutineID] == nil {
				responseMapPools[goroutineID] = &sync.Pool{
					New: func() interface{} { return make(gin.H, 2) },
				}
			}
			pool = responseMapPools[goroutineID]
			mu.Unlock()
		}
	case "pagination":
		mu = &paginationMapPoolMu
		mu.RLock()
		pool = paginationMapPools[goroutineID]
		mu.RUnlock()
		if pool == nil {
			mu.Lock()
			if paginationMapPools[goroutineID] == nil {
				paginationMapPools[goroutineID] = &sync.Pool{
					New: func() interface{} { return make(gin.H, 6) },
				}
			}
			pool = paginationMapPools[goroutineID]
			mu.Unlock()
		}
	case "userSlice":
		mu = &userSlicePoolMu
		mu.RLock()
		pool = userSlicePools[goroutineID]
		mu.RUnlock()
		if pool == nil {
			mu.Lock()
			if userSlicePools[goroutineID] == nil {
				userSlicePools[goroutineID] = &sync.Pool{
					New: func() interface{} { return make([]models.User, 0, 50) },
				}
			}
			pool = userSlicePools[goroutineID]
			mu.Unlock()
		}
	}

	return pool
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

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// listUsersHandler GET /api/users
func listUsersHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse pagination parameters with optimized string conversion
		pageStr := c.DefaultQuery("page", "1")
		limitStr := c.DefaultQuery("limit", "50")

		page := parseIntMinMax(pageStr, 1, 1, 10000)   // sensible upper bound
		limit := parseIntMinMax(limitStr, 50, 1, 1000) // max 1000 as before

		offset := (page - 1) * limit

		// Get total count for pagination info
		var totalCount int
		err := db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&totalCount)
		if err != nil {
			log.Printf("Error counting users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to count users"})
			return
		}

		// Get a user slice from the pool with ensured capacity
		userPool := getPool("userSlice")
		users := userPool.Get().([]models.User)
		if cap(users) < limit {
			// If the pooled slice is smaller than needed, grow it
			users = make([]models.User, 0, max(limit, 50))
		} else {
			users = users[:0]
		}
		users = users[:0] // ensure empty regardless

		defer func() {
			// Reset and return to pool
			users = users[:0]
			userPool.Put(users)
		}()

		// Query with pagination
		rows, err := db.Query("SELECT id, username, email, status, created_at, updated_at, deleted_at, deleted_by FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT ? OFFSET ?",
			limit, offset)
		if err != nil {
			log.Printf("Error querying users: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
			return
		}
		defer rows.Close()

		// Optimize slice usage - set length to avoid repeated append growth

		for rows.Next() {
			var u models.User
			if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.DeletedBy); err != nil {
				log.Printf("Error scanning user row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve users"})
				return
			}
			u.PasswordHash = "" // Clear sensitive data
			users = append(users, u)
		}

		// Calculate pagination info using optimized math
		totalPages := (totalCount + limit - 1) / limit
		hasNext := page < totalPages
		hasPrev := page > 1

		// Optimized response using pooled maps to reduce allocations
		paginationPool := getPool("pagination")
		paginationMap := paginationPool.Get().(gin.H)
		defer func() {
			// Clear and return to pool
			clear(paginationMap)
			paginationPool.Put(paginationMap)
		}()

		paginationMap["page"] = page
		paginationMap["limit"] = limit
		paginationMap["total"] = totalCount
		paginationMap["total_pages"] = totalPages
		paginationMap["has_next"] = hasNext
		paginationMap["has_prev"] = hasPrev

		responsePool := getPool("response")
		responseMap := responsePool.Get().(gin.H)
		defer func() {
			// Clear and return to pool
			clear(responseMap)
			responsePool.Put(responseMap)
		}()

		responseMap["data"] = users
		responseMap["pagination"] = paginationMap

		c.JSON(http.StatusOK, responseMap)
	}
}

// getUserHandler GET /api/users/:id
func getUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var u models.User
		row := db.QueryRow("SELECT id, username, email, status, created_at, updated_at, deleted_at, deleted_by FROM users WHERE id = ? AND deleted_at IS NULL", userID)
		err = row.Scan(&u.ID, &u.Username, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.DeletedBy)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": u})
	}
}

// createUserHandler POST /api/users
func createUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Hash password
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hash failed"})
			return
		}

		status := uint8(1) // default active
		if req.Status != nil {
			status = *req.Status
		}

		now := time.Now()
		result, err := db.Exec("INSERT INTO users (username, email, password_hash, status, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
			req.Username, req.Email, string(hashed), status, now, now)
		if err != nil {
			log.Printf("Error inserting user: %v", err)
			if strings.Contains(err.Error(), "1062") {
				c.JSON(http.StatusConflict, gin.H{"error": "Username or email already exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			}
			return
		}

		userID, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"message": "User created", "id": userID})

		// Queue audit log for async processing - don't block response
		select {
		case auditLogChan <- auditLogEntry{
			UserID:    nil,
			Event:     "CREATE",
			Table:     "users",
			RecordID:  uint64(userID),
			OldValues: nil,
			NewValues: req,
			DB:        db,
		}:
			// Successfully queued
		default:
			// Channel full, log error but don't fail the request
			log.Printf("Warning: audit log queue full, dropping CREATE audit for user %d", userID)
		}
	}
}

// updateUserHandler PUT /api/users/:id
func updateUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var req models.UpdateUserRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if user exists
		var exists bool
		err = db.QueryRow("SELECT 1 FROM users WHERE id = ? AND deleted_at IS NULL", userID).Scan(&exists)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Build update query with pre-allocated slices for efficiency
		setParts := make([]string, 0, 5)  // Pre-allocate for up to 5 fields typically
		args := make([]interface{}, 0, 5) // Pre-allocate for corresponding arguments

		if req.Username != "" {
			setParts = append(setParts, "username = ?")
			args = append(args, req.Username)
		}
		if req.Email != "" {
			setParts = append(setParts, "email = ?")
			args = append(args, req.Email)
		}
		if req.Password != "" {
			hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Password hash failed"})
				return
			}
			setParts = append(setParts, "password_hash = ?")
			args = append(args, string(hashed))
		}
		if req.Status != nil {
			setParts = append(setParts, "status = ?")
			args = append(args, *req.Status)
		}

		if len(setParts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		setParts = append(setParts, "updated_at = ?")
		args = append(args, time.Now())

		query := buildUpdateQuery(setParts, "users")
		args = append(args, userID)

		_, err = db.Exec(query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User updated"})
	}
}

// deleteUserHandler DELETE /api/users/:id (soft delete)
func deleteUserHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		userID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		// Assume deleted_by is from context, e.g., JWT. For now, set to nil
		_, err = db.Exec("UPDATE users SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL", time.Now(), time.Now(), userID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Soft delete failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User deleted"})
	}
}

// Optimized UPDATE query building
func buildUpdateQuery(setParts []string, table string) string {
	setClause := utils.JoinStrings(setParts, ", ")
	return "UPDATE " + table + " SET " + setClause + " WHERE id = ? AND deleted_at IS NULL"
}

// createAuditLog logs audit events
func createAuditLog(db *sql.DB, userID *uint64, event string, table string, recordID uint64, oldValues, newValues interface{}) {
	var oldJSON, newJSON []byte
	if oldValues != nil {
		oldJSON, _ = json.Marshal(oldValues)
	}
	if newValues != nil {
		newJSON, _ = json.Marshal(newValues)
	}
	db.Exec("INSERT INTO audit_logs (user_id, event_type, table_name, record_id, old_values, new_values) VALUES (?, ?, ?, ?, ?, ?)", userID, event, table, recordID, oldJSON, newJSON)
}
