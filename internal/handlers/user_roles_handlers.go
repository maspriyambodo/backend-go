package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"adminbe/internal/models"

	"github.com/gin-gonic/gin"
)

// listUserRolesHandler GET /api/user_roles
func listUserRolesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT user_id, role_id, deleted_at, deleted_by FROM user_roles WHERE deleted_at IS NULL")
		if err != nil {
			log.Printf("Error querying user_roles: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user-role assignments"})
			return
		}
		defer rows.Close()

		var userRoles []models.UserRole
		for rows.Next() {
			var ur models.UserRole
			if err := rows.Scan(&ur.UserID, &ur.RoleID, &ur.DeletedAt, &ur.DeletedBy); err != nil {
				log.Printf("Error scanning user_role row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user-role assignments"})
				return
			}
			userRoles = append(userRoles, ur)
		}
		c.JSON(http.StatusOK, gin.H{"data": userRoles})
	}
}

// getUserRoleHandler GET /api/user_roles/:userId/:roleId
func getUserRoleHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		roleIDStr := c.Param("roleId")
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}

		var ur models.UserRole
		row := db.QueryRow("SELECT user_id, role_id, deleted_at, deleted_by FROM user_roles WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL", userID, uint(roleID))
		err = row.Scan(&ur.UserID, &ur.RoleID, &ur.DeletedAt, &ur.DeletedBy)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User-role assignment not found"})
			return
		} else if err != nil {
			log.Printf("Error querying user_role: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": ur})
	}
}

// createUserRoleHandler POST /api/user_roles
func createUserRoleHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateUserRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if already exists active
		var exists bool
		err := db.QueryRow("SELECT 1 FROM user_roles WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL", req.UserID, req.RoleID).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error checking existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "User-role assignment already exists"})
			return
		}

		_, err = db.Exec("INSERT INTO user_roles (user_id, role_id, deleted_at, deleted_by) VALUES (?, ?, ?, ?)",
			req.UserID, req.RoleID, nil, nil)
		if err != nil {
			log.Printf("Error inserting user_role: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user-role assignment"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User-role assignment created"})
		createAuditLog(db, nil, "CREATE", "user_roles", uint64(req.UserID), nil, req)
	}
}

// updateUserRoleHandler PUT /api/user_roles/:userId/:roleId
func updateUserRoleHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		roleIDStr := c.Param("roleId")
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}

		var req models.UpdateUserRoleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if exists
		var exists bool
		err = db.QueryRow("SELECT 1 FROM user_roles WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL", userID, uint(roleID)).Scan(&exists)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User-role assignment not found"})
			return
		} else if err != nil {
			log.Printf("Error checking existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Get old values
		var oldUserRole struct {
			UserID uint64 `json:"user_id"`
			RoleID uint   `json:"role_id"`
		}
		oldUserRole.UserID = userID
		oldUserRole.RoleID = uint(roleID)

		// Build update
		setParts := []string{}
		args := []interface{}{}

		if req.UserID != nil {
			setParts = append(setParts, "user_id = ?")
			args = append(args, req.UserID)
		}
		if req.RoleID != nil {
			setParts = append(setParts, "role_id = ?")
			args = append(args, req.RoleID)
		}

		if len(setParts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		query := "UPDATE user_roles SET " + join(setParts, ", ") + " WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL"
		args = append(args, userID, uint(roleID))

		_, err = db.Exec(query, args...)
		if err != nil {
			log.Printf("Error updating user_role: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User-role assignment updated"})
		createAuditLog(db, nil, "UPDATE", "user_roles", userID, oldUserRole, req)
	}
}

// deleteUserRoleHandler DELETE /api/user_roles/:userId/:roleId
func deleteUserRoleHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		roleIDStr := c.Param("roleId")
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}

		// Get old values before delete
		var oldUserRole struct {
			UserID uint64 `json:"user_id"`
			RoleID uint   `json:"role_id"`
		}
		oldUserRole.UserID = userID
		oldUserRole.RoleID = uint(roleID)

		_, err = db.Exec("UPDATE user_roles SET deleted_at = ? WHERE user_id = ? AND role_id = ? AND deleted_at IS NULL", time.Now(), userID, uint(roleID))
		if err != nil {
			log.Printf("Error soft deleting user_role: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Soft delete failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User-role assignment deleted"})
		createAuditLog(db, nil, "DELETE", "user_roles", userID, oldUserRole, nil)
	}
}
