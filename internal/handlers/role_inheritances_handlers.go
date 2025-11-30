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

// listRoleInheritancesHandler GET /api/role_inheritances
func listRoleInheritancesHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, role_id, parent_role_id, created_at FROM role_inheritances")
		if err != nil {
			log.Printf("Error querying role inheritances: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve role inheritances"})
			return
		}
		defer rows.Close()

		var inheritances []models.RoleInheritance
		for rows.Next() {
			var ri models.RoleInheritance
			if err := rows.Scan(&ri.ID, &ri.RoleID, &ri.ParentRoleID, &ri.CreatedAt); err != nil {
				log.Printf("Error scanning role inheritance row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve role inheritances"})
				return
			}
			inheritances = append(inheritances, ri)
		}
		c.JSON(http.StatusOK, gin.H{"data": inheritances})
	}
}

// getRoleInheritanceHandler GET /api/role_inheritances/:id
func getRoleInheritanceHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		inheritanceID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var ri models.RoleInheritance
		row := db.QueryRow("SELECT id, role_id, parent_role_id, created_at FROM role_inheritances WHERE id = ?", inheritanceID)
		err = row.Scan(&ri.ID, &ri.RoleID, &ri.ParentRoleID, &ri.CreatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role inheritance not found"})
			return
		} else if err != nil {
			log.Printf("Error querying role inheritance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": ri})
	}
}

// createRoleInheritanceHandler POST /api/role_inheritances
func createRoleInheritanceHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateRoleInheritanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		now := time.Now()
		result, err := db.Exec("INSERT INTO role_inheritances (role_id, parent_role_id, created_at) VALUES (?, ?, ?)",
			req.RoleID, req.ParentRoleID, now)
		if err != nil {
			log.Printf("Error inserting role inheritance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role inheritance"})
			return
		}

		inheritanceID, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"message": "Role inheritance created", "id": inheritanceID})
		createAuditLog(db, nil, "CREATE", "role_inheritances", uint64(inheritanceID), nil, req)
	}
}

// updateRoleInheritanceHandler PUT /api/role_inheritances/:id
func updateRoleInheritanceHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		inheritanceID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var req models.UpdateRoleInheritanceRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if exists
		var exists bool
		err = db.QueryRow("SELECT 1 FROM role_inheritances WHERE id = ?", inheritanceID).Scan(&exists)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role inheritance not found"})
			return
		} else if err != nil {
			log.Printf("Error checking role inheritance existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Get old values
		var oldInheritance struct {
			RoleID       uint `json:"role_id"`
			ParentRoleID uint `json:"parent_role_id"`
		}
		err = db.QueryRow("SELECT role_id, parent_role_id FROM role_inheritances WHERE id = ?", inheritanceID).Scan(&oldInheritance.RoleID, &oldInheritance.ParentRoleID)
		if err != nil {
			log.Printf("Error getting old role inheritance values: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Build update
		setParts := []string{}
		args := []interface{}{}

		if req.RoleID != nil {
			setParts = append(setParts, "role_id = ?")
			args = append(args, req.RoleID)
		}
		if req.ParentRoleID != nil {
			setParts = append(setParts, "parent_role_id = ?")
			args = append(args, req.ParentRoleID)
		}

		if len(setParts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		query := "UPDATE role_inheritances SET " + join(setParts, ", ") + " WHERE id = ?"
		args = append(args, inheritanceID)

		_, err = db.Exec(query, args...)
		if err != nil {
			log.Printf("Error updating role inheritance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Role inheritance updated"})
		createAuditLog(db, nil, "UPDATE", "role_inheritances", inheritanceID, oldInheritance, req)
	}
}

// deleteRoleInheritanceHandler DELETE /api/role_inheritances/:id
func deleteRoleInheritanceHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		inheritanceID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		// Get old values before delete
		var oldInheritance struct {
			RoleID       uint `json:"role_id"`
			ParentRoleID uint `json:"parent_role_id"`
		}
		err = db.QueryRow("SELECT role_id, parent_role_id FROM role_inheritances WHERE id = ?", inheritanceID).Scan(&oldInheritance.RoleID, &oldInheritance.ParentRoleID)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role inheritance not found"})
			return
		} else if err != nil {
			log.Printf("Error getting old role inheritance values: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		_, err = db.Exec("DELETE FROM role_inheritances WHERE id = ?", inheritanceID)
		if err != nil {
			log.Printf("Error deleting role inheritance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Delete failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Role inheritance deleted"})
		createAuditLog(db, nil, "DELETE", "role_inheritances", inheritanceID, oldInheritance, nil)
	}
}
