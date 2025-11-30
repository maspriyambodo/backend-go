package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"adminbe/internal/app/models"

	"github.com/gin-gonic/gin"
)

// listRolesHandler GET /api/roles
func listRolesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, name, description, created_at, updated_at, deleted_at, deleted_by FROM roles WHERE deleted_at IS NULL")
		if err != nil {
			log.Printf("Error querying roles: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve roles"})
			return
		}
		defer rows.Close()

		var roles []models.Role
		for rows.Next() {
			var r models.Role
			if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt, &r.DeletedBy); err != nil {
				log.Printf("Error scanning role row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve roles"})
				return
			}
			roles = append(roles, r)
		}
		c.JSON(http.StatusOK, gin.H{"data": roles})
	}
}

// getRoleHandler GET /api/roles/:id
func getRoleHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		roleID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var r models.Role
		row := db.QueryRow("SELECT id, name, description, created_at, updated_at, deleted_at, deleted_by FROM roles WHERE id = ? AND deleted_at IS NULL", uint(roleID))
		err = row.Scan(&r.ID, &r.Name, &r.Description, &r.CreatedAt, &r.UpdatedAt, &r.DeletedAt, &r.DeletedBy)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		} else if err != nil {
			log.Printf("Error querying role: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": r})
	}
}

// createRoleHandler POST /api/roles
func createRoleHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now()
		result, err := db.Exec("INSERT INTO roles (name, description, created_at, updated_at) VALUES (?, ?, ?, ?)",
			req.Name, req.Description, now, now)
		if err != nil {
			log.Printf("Error inserting role: %v", err)
			if strings.Contains(err.Error(), "1062") {
				c.JSON(http.StatusConflict, gin.H{"error": "Role name already exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role"})
			}
			return
		}

		roleID, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"message": "Role created", "id": roleID})
		createAuditLog(db, nil, "CREATE", "roles", uint64(roleID), nil, req)
	}
}

// updateRoleHandler PUT /api/roles/:id
func updateRoleHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		roleID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var req models.UpdateRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if exists
		var exists bool
		err = db.QueryRow("SELECT 1 FROM roles WHERE id = ? AND deleted_at IS NULL", uint(roleID)).Scan(&exists)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		} else if err != nil {
			log.Printf("Error checking role existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Get old values
		var oldRole struct {
			Name        string  `json:"name"`
			Description *string `json:"description"`
		}
		err = db.QueryRow("SELECT name, description FROM roles WHERE id = ?", uint(roleID)).Scan(&oldRole.Name, &oldRole.Description)
		if err != nil {
			log.Printf("Error getting old role values: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Build update
		setParts := []string{}
		args := []interface{}{}

		if req.Name != nil {
			setParts = append(setParts, "name = ?")
			args = append(args, req.Name)
		}
		if req.Description != nil {
			setParts = append(setParts, "description = ?")
			args = append(args, req.Description)
		}

		if len(setParts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		setParts = append(setParts, "updated_at = ?")
		args = append(args, time.Now())

		query := "UPDATE roles SET " + join(setParts, ", ") + " WHERE id = ? AND deleted_at IS NULL"
		args = append(args, uint(roleID))

		_, err = db.Exec(query, args...)
		if err != nil {
			log.Printf("Error updating role: %v", err)
			if strings.Contains(err.Error(), "1062") {
				c.JSON(http.StatusConflict, gin.H{"error": "Role name already exists"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			}
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
		createAuditLog(db, nil, "UPDATE", "roles", uint64(roleID), oldRole, req)
	}
}

// deleteRoleHandler DELETE /api/roles/:id
func deleteRoleHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		roleID, err := strconv.ParseUint(id, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		// Get current user ID
		userIDVal, exists := c.Get("user_id")
		var userID *uint64
		if exists {
			uid := userIDVal.(uint64)
			userID = &uid
		}

		// Get old values before delete
		var oldRole struct {
			Name        string  `json:"name"`
			Description *string `json:"description"`
		}
		err = db.QueryRow("SELECT name, description FROM roles WHERE id = ?", uint(roleID)).Scan(&oldRole.Name, &oldRole.Description)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role not found"})
			return
		} else if err != nil {
			log.Printf("Error getting old role values: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		_, err = db.Exec("UPDATE roles SET deleted_at = ?, updated_at = ?, deleted_by = ? WHERE id = ? AND deleted_at IS NULL", time.Now(), time.Now(), userID, uint(roleID))
		if err != nil {
			log.Printf("Error soft deleting role: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Soft delete failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Role deleted"})
		createAuditLog(db, userID, "DELETE", "roles", uint64(roleID), oldRole, nil)
	}
}
