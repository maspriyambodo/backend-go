package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"strconv"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/pkg/utils"

	"github.com/gin-gonic/gin"
)

// listMenuHandler GET /api/menu
func listMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, label, url, icon, parent_id, sort_order, created_at, updated_at, deleted_at, deleted_by FROM menu WHERE deleted_at IS NULL ORDER BY sort_order")
		if err != nil {
			log.Printf("Error querying menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menu"})
			return
		}
		defer rows.Close()

		var menus []models.Menu
		for rows.Next() {
			var m models.Menu
			if err := rows.Scan(&m.ID, &m.Label, &m.Url, &m.Icon, &m.ParentID, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.DeletedBy); err != nil {
				log.Printf("Error scanning menu row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menu"})
				return
			}
			menus = append(menus, m)
		}
		c.JSON(http.StatusOK, gin.H{"data": menus})
	}
}

// getMenuHandler GET /api/menu/:id
func getMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		menuID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var m models.Menu
		row := db.QueryRow("SELECT id, label, url, icon, parent_id, sort_order, created_at, updated_at, deleted_at, deleted_by FROM menu WHERE id = ? AND deleted_at IS NULL", menuID)
		err = row.Scan(&m.ID, &m.Label, &m.Url, &m.Icon, &m.ParentID, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.DeletedBy)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		} else if err != nil {
			log.Printf("Error querying menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": m})
	}
}

// CreateMenuRequest for creating a new menu
type CreateMenuRequest struct {
	Label     string  `json:"label" binding:"required,min=1,max=100"`
	Url       *string `json:"url,omitempty"`
	Icon      *string `json:"icon,omitempty"`
	ParentID  *uint   `json:"parent_id,omitempty"`
	SortOrder *uint16 `json:"sort_order,omitempty"`
}

// createMenuHandler POST /api/menu
func createMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		sortOrder := uint16(0) // default
		if req.SortOrder != nil {
			sortOrder = *req.SortOrder
		}

		now := time.Now()
		result, err := db.Exec("INSERT INTO menu (label, url, icon, parent_id, sort_order, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
			req.Label, req.Url, req.Icon, req.ParentID, sortOrder, now, now)
		if err != nil {
			log.Printf("Error inserting menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu"})
			return
		}

		menuID, _ := result.LastInsertId()
		c.JSON(http.StatusCreated, gin.H{"message": "Menu created", "id": menuID})
		createAuditLog(db, nil, "CREATE", "menu", uint64(menuID), nil, req)
	}
}

// UpdateMenuRequest for updating a menu
type UpdateMenuRequest struct {
	Label     *string `json:"label,omitempty" binding:"min=1,max=100"`
	Url       *string `json:"url,omitempty"`
	Icon      *string `json:"icon,omitempty"`
	ParentID  *uint   `json:"parent_id,omitempty"`
	SortOrder *uint16 `json:"sort_order,omitempty"`
}

// updateMenuHandler PUT /api/menu/:id
func updateMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		menuID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		var req UpdateMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// Check if exists
		var exists bool
		err = db.QueryRow("SELECT 1 FROM menu WHERE id = ? AND deleted_at IS NULL", menuID).Scan(&exists)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		} else if err != nil {
			log.Printf("Error checking menu existence: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Get old values
		var oldMenu struct {
			Label     string  `json:"label"`
			Url       *string `json:"url"`
			Icon      *string `json:"icon"`
			ParentID  *uint   `json:"parent_id"`
			SortOrder uint16  `json:"sort_order"`
		}
		err = db.QueryRow("SELECT label, url, icon, parent_id, sort_order FROM menu WHERE id = ?", menuID).Scan(&oldMenu.Label, &oldMenu.Url, &oldMenu.Icon, &oldMenu.ParentID, &oldMenu.SortOrder)
		if err != nil {
			log.Printf("Error getting old menu values: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		// Build update
		setParts := []string{}
		args := []interface{}{}

		if req.Label != nil {
			setParts = append(setParts, "label = ?")
			args = append(args, req.Label)
		}
		if req.Url != nil {
			setParts = append(setParts, "url = ?")
			args = append(args, req.Url)
		}
		if req.Icon != nil {
			setParts = append(setParts, "icon = ?")
			args = append(args, req.Icon)
		}
		if req.ParentID != nil {
			setParts = append(setParts, "parent_id = ?")
			args = append(args, req.ParentID)
		}
		if req.SortOrder != nil {
			setParts = append(setParts, "sort_order = ?")
			args = append(args, req.SortOrder)
		}

		if len(setParts) == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "No fields to update"})
			return
		}

		setParts = append(setParts, "updated_at = ?")
		args = append(args, time.Now())

		query := "UPDATE menu SET " + utils.JoinStrings(setParts, ", ") + " WHERE id = ? AND deleted_at IS NULL"
		args = append(args, menuID)

		_, err = db.Exec(query, args...)
		if err != nil {
			log.Printf("Error updating menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Update failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Menu updated"})
		createAuditLog(db, nil, "UPDATE", "menu", menuID, oldMenu, req)
	}
}

// deleteMenuHandler DELETE /api/menu/:id
func deleteMenuHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		menuID, err := strconv.ParseUint(id, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
			return
		}

		// Get old values before delete
		var oldMenu struct {
			Label     string  `json:"label"`
			Url       *string `json:"url"`
			Icon      *string `json:"icon"`
			ParentID  *uint   `json:"parent_id"`
			SortOrder uint16  `json:"sort_order"`
		}
		err = db.QueryRow("SELECT label, url, icon, parent_id, sort_order FROM menu WHERE id = ?", menuID).Scan(&oldMenu.Label, &oldMenu.Url, &oldMenu.Icon, &oldMenu.ParentID, &oldMenu.SortOrder)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		} else if err != nil {
			log.Printf("Error getting old menu values: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database query failed"})
			return
		}

		_, err = db.Exec("UPDATE menu SET deleted_at = ?, updated_at = ? WHERE id = ? AND deleted_at IS NULL", time.Now(), time.Now(), menuID)
		if err != nil {
			log.Printf("Error soft deleting menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Soft delete failed"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Menu deleted"})
		createAuditLog(db, nil, "DELETE", "menu", menuID, oldMenu, nil)
	}
}
