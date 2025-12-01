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

// listRoleMenusHandler GET /api/role_menu
func listRoleMenusHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT role_id, menu_id, deleted_at, deleted_by FROM role_menu WHERE deleted_at IS NULL")
		if err != nil {
			log.Printf("Error querying role_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve role-menu assignments"})
			return
		}
		defer rows.Close()

		var roleMenus []models.RoleMenu
		for rows.Next() {
			var rm models.RoleMenu
			if err := rows.Scan(&rm.RoleID, &rm.MenuID, &rm.DeletedAt, &rm.DeletedBy); err != nil {
				log.Printf("Error scanning role_menu row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve role-menu assignments"})
				return
			}
			roleMenus = append(roleMenus, rm)
		}
		c.JSON(http.StatusOK, gin.H{"data": roleMenus})
	}
}

// getRoleMenuHandler GET /api/role_menu/:roleId/:menuId
func getRoleMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIDStr := c.Param("roleId")
		menuIDStr := c.Param("menuId")
		roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}
		menuID, err := strconv.ParseUint(menuIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
			return
		}

		var rm models.RoleMenu
		row := db.QueryRow("SELECT role_id, menu_id, deleted_at, deleted_by FROM role_menu WHERE role_id = ? AND menu_id = ? AND deleted_at IS NULL", uint(roleID), uint(menuID))
		err = row.Scan(&rm.RoleID, &rm.MenuID, &rm.DeletedAt, &rm.DeletedBy)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role-menu assignment not found"})
			return
		} else if err != nil {
			log.Printf("Error querying role_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": rm})
	}
}

// createRoleMenuHandler POST /api/role_menu
func createRoleMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateRoleMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if already exists active
		var exists bool
		err := db.QueryRow("SELECT 1 FROM role_menu WHERE role_id = ? AND menu_id = ? AND deleted_at IS NULL", req.RoleID, req.MenuID).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error checking existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "Role-menu assignment already exists"})
			return
		}

		_, err = db.Exec("INSERT INTO role_menu (role_id, menu_id, deleted_at, deleted_by) VALUES (?, ?, ?, ?)",
			req.RoleID, req.MenuID, nil, nil)
		if err != nil {
			log.Printf("Error inserting role_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create role-menu assignment"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Role-menu assignment created"})
		createAuditLog(db, nil, "CREATE", "role_menu", uint64(req.RoleID), nil, req)
	}
}

// updateRoleMenuHandler PUT /api/role_menu/:roleId/:menuId
func updateRoleMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIDStr := c.Param("roleId")
		menuIDStr := c.Param("menuId")
		roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}
		menuID, err := strconv.ParseUint(menuIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
			return
		}

		var req models.UpdateRoleMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if exists
		var exists bool
		err = db.QueryRow("SELECT 1 FROM role_menu WHERE role_id = ? AND menu_id = ? AND deleted_at IS NULL", uint(roleID), uint(menuID)).Scan(&exists)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Role-menu assignment not found"})
			return
		} else if err != nil {
			log.Printf("Error checking existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Get old values
		var oldRoleMenu struct {
			RoleID uint `json:"role_id"`
			MenuID uint `json:"menu_id"`
		}
		oldRoleMenu.RoleID = uint(roleID)
		oldRoleMenu.MenuID = uint(menuID)

		// Build update
		setParts := []string{}
		args := []interface{}{}

		if req.RoleID != nil {
			setParts = append(setParts, "role_id = ?")
			args = append(args, req.RoleID)
		}
		if req.MenuID != nil {
			setParts = append(setParts, "menu_id = ?")
			args = append(args, req.MenuID)
		}

		if len(setParts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		query := "UPDATE role_menu SET " + strings.Join(setParts, ", ") + " WHERE role_id = ? AND menu_id = ? AND deleted_at IS NULL"
		args = append(args, uint(roleID), uint(menuID))

		_, err = db.Exec(query, args...)
		if err != nil {
			log.Printf("Error updating role_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Role-menu assignment updated"})
		createAuditLog(db, nil, "UPDATE", "role_menu", uint64(roleID), oldRoleMenu, req)
	}
}

// deleteRoleMenuHandler DELETE /api/role_menu/:roleId/:menuId
func deleteRoleMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleIDStr := c.Param("roleId")
		menuIDStr := c.Param("menuId")
		roleID, err := strconv.ParseUint(roleIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role ID"})
			return
		}
		menuID, err := strconv.ParseUint(menuIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
			return
		}

		// Get old values before delete
		var oldRoleMenu struct {
			RoleID uint `json:"role_id"`
			MenuID uint `json:"menu_id"`
		}
		oldRoleMenu.RoleID = uint(roleID)
		oldRoleMenu.MenuID = uint(menuID)

		_, err = db.Exec("UPDATE role_menu SET deleted_at = ? WHERE role_id = ? AND menu_id = ? AND deleted_at IS NULL", time.Now(), uint(roleID), uint(menuID))
		if err != nil {
			log.Printf("Error soft deleting role_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Soft delete failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Role-menu assignment deleted"})
		createAuditLog(db, nil, "DELETE", "role_menu", uint64(roleID), oldRoleMenu, nil)
	}
}
