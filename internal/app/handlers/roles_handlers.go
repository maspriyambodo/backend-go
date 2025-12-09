package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/app/services"
	"adminbe/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

// listRolesHandler GET /api/roles
func listRolesHandler(roleService services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		roles, err := roleService.ListRoles()
		if err != nil {
			log.Printf("Error listing roles: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve roles"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": roles})
	}
}

// getRoleHandler GET /api/roles/:id
func getRoleHandler(roleService services.RoleService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		role, err := roleService.GetRole(id)
		if handleServiceError(c, err, "role") {
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": role})
	}
}

// createRoleHandler POST /api/roles
func createRoleHandler(roleService services.RoleService, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateRoleRequest
		if !bindJSONRequest(c, &req) {
			return
		}

		role, err := roleService.CreateRole(req)
		if err != nil {
			handleServiceError(c, err, "create role")
			return
		}

		// Audit logging
		logAuditEntry(c, "CREATE", "roles", uint64(role.ID), nil, req, db)

		c.JSON(http.StatusCreated, gin.H{"message": "Role created", "data": role})
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

		query := "UPDATE roles SET " + utils.JoinStrings(setParts, ", ") + " WHERE id = ? AND deleted_at IS NULL"
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

		// Audit logging
		logAuditEntry(c, "UPDATE", "roles", uint64(roleID), oldRole, req, db)

		c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
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

		// Perform soft delete
		_, err = db.Exec("UPDATE roles SET deleted_at = ?, updated_at = ?, deleted_by = ? WHERE id = ? AND deleted_at IS NULL",
			time.Now(), time.Now(), getUserIDFromContext(c), uint(roleID))
		if err != nil {
			log.Printf("Error soft deleting role: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Soft delete failed"})
			return
		}

		// Audit logging
		logAuditEntry(c, "DELETE", "roles", uint64(roleID), oldRole, nil, db)

		c.JSON(http.StatusOK, gin.H{"message": "Role deleted"})
	}
}
