package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"adminbe/internal/app/models"

	"github.com/gin-gonic/gin"
)

// listUserMenusHandler GET /api/user_menu
func listUserMenusHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT user_id, menu_id, deleted_at, deleted_by FROM user_menu WHERE deleted_at IS NULL")
		if err != nil {
			log.Printf("Error querying user_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user-menu assignments"})
			return
		}
		defer rows.Close()

		var userMenus []models.UserMenu
		for rows.Next() {
			var um models.UserMenu
			if err := rows.Scan(&um.UserID, &um.MenuID, &um.DeletedAt, &um.DeletedBy); err != nil {
				log.Printf("Error scanning user_menu row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve user-menu assignments"})
				return
			}
			userMenus = append(userMenus, um)
		}
		c.JSON(http.StatusOK, gin.H{"data": userMenus})
	}
}

// getUserMenuHandler GET /api/user_menu/:userId/:menuId
func getUserMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		menuIDStr := c.Param("menuId")
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		menuID, err := strconv.ParseUint(menuIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
			return
		}

		var um models.UserMenu
		row := db.QueryRow("SELECT user_id, menu_id, deleted_at, deleted_by FROM user_menu WHERE user_id = ? AND menu_id = ? AND deleted_at IS NULL", userID, uint(menuID))
		err = row.Scan(&um.UserID, &um.MenuID, &um.DeletedAt, &um.DeletedBy)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User-menu assignment not found"})
			return
		} else if err != nil {
			log.Printf("Error querying user_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": um})
	}
}

// createUserMenuHandler POST /api/user_menu
func createUserMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateUserMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if already exists active
		var exists bool
		err := db.QueryRow("SELECT 1 FROM user_menu WHERE user_id = ? AND menu_id = ? AND deleted_at IS NULL", req.UserID, req.MenuID).Scan(&exists)
		if err != nil && err != sql.ErrNoRows {
			log.Printf("Error checking existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		if exists {
			c.JSON(http.StatusConflict, gin.H{"error": "User-menu assignment already exists"})
			return
		}

		_, err = db.Exec("INSERT INTO user_menu (user_id, menu_id, deleted_at, deleted_by) VALUES (?, ?, ?, ?)",
			req.UserID, req.MenuID, nil, nil)
		if err != nil {
			log.Printf("Error inserting user_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user-menu assignment"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "User-menu assignment created"})
		createAuditLog(db, nil, "CREATE", "user_menu", uint64(req.UserID), nil, req)
	}
}

// updateUserMenuHandler PUT /api/user_menu/:userId/:menuId
func updateUserMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		menuIDStr := c.Param("menuId")
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		menuID, err := strconv.ParseUint(menuIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
			return
		}

		var req models.UpdateUserMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if exists
		var exists bool
		err = db.QueryRow("SELECT 1 FROM user_menu WHERE user_id = ? AND menu_id = ? AND deleted_at IS NULL", userID, uint(menuID)).Scan(&exists)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "User-menu assignment not found"})
			return
		} else if err != nil {
			log.Printf("Error checking existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Get old values
		var oldUserMenu struct {
			UserID uint64 `json:"user_id"`
			MenuID uint   `json:"menu_id"`
		}
		oldUserMenu.UserID = userID
		oldUserMenu.MenuID = uint(menuID)

		// Build update
		setParts := []string{}
		args := []interface{}{}

		if req.UserID != nil {
			setParts = append(setParts, "user_id = ?")
			args = append(args, req.UserID)
		}
		if req.MenuID != nil {
			setParts = append(setParts, "menu_id = ?")
			args = append(args, req.MenuID)
		}

		if len(setParts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		query := "UPDATE user_menu SET " + join(setParts, ", ") + " WHERE user_id = ? AND menu_id = ? AND deleted_at IS NULL"
		args = append(args, userID, uint(menuID))

		_, err = db.Exec(query, args...)
		if err != nil {
			log.Printf("Error updating user_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User-menu assignment updated"})
		createAuditLog(db, nil, "UPDATE", "user_menu", userID, oldUserMenu, req)
	}
}

// deleteUserMenuHandler DELETE /api/user_menu/:userId/:menuId
func deleteUserMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userIDStr := c.Param("userId")
		menuIDStr := c.Param("menuId")
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			return
		}
		menuID, err := strconv.ParseUint(menuIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid menu ID"})
			return
		}

		// Get old values before delete
		var oldUserMenu struct {
			UserID uint64 `json:"user_id"`
			MenuID uint   `json:"menu_id"`
		}
		oldUserMenu.UserID = userID
		oldUserMenu.MenuID = uint(menuID)

		_, err = db.Exec("UPDATE user_menu SET deleted_at = ? WHERE user_id = ? AND menu_id = ? AND deleted_at IS NULL", time.Now(), userID, uint(menuID))
		if err != nil {
			log.Printf("Error soft deleting user_menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Soft delete failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "User-menu assignment deleted"})
		createAuditLog(db, nil, "DELETE", "user_menu", userID, oldUserMenu, nil)
	}
}
