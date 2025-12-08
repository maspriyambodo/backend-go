package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"adminbe/internal/app/models"
	"adminbe/internal/app/services"

	"github.com/gin-gonic/gin"
)

// listMenuHandler GET /api/menu
func listMenuHandler(menuService services.MenuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		menus, err := menuService.ListMenus()
		if err != nil {
			log.Printf("Error listing menus: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menu"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": menus})
	}
}

// getMenuHandler GET /api/menu/:id
func getMenuHandler(menuService services.MenuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")
		menu, err := menuService.GetMenu(id)
		if err != nil && isNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		}
		if err != nil {
			log.Printf("Error getting menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menu"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": menu})
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
func createMenuHandler(menuService services.MenuService, db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		menu := models.Menu{
			Label:    req.Label,
			Url:      req.Url,
			Icon:     req.Icon,
			ParentID: req.ParentID,
			SortOrder: func() uint16 {
				if req.SortOrder != nil {
					return *req.SortOrder
				}
				return 0
			}(),
		}

		createdMenu, err := menuService.CreateMenu(menu)
		if err != nil {
			log.Printf("Error creating menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create menu"})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"message": "Menu created", "data": createdMenu})

		// Audit logging
		if auditLogChan != nil {
			select {
			case auditLogChan <- auditLogEntry{
				UserID:    nil,
				Event:     "CREATE",
				Table:     "menu",
				RecordID:  uint64(createdMenu.ID),
				OldValues: nil,
				NewValues: req,
				DB:        db,
			}:
			default:
				log.Printf("Warning: audit log queue full, dropping CREATE audit for menu %d", createdMenu.ID)
			}
		}
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
func updateMenuHandler(menuService services.MenuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		var req UpdateMenuRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		updateData := make(map[string]interface{})
		if req.Label != nil {
			updateData["label"] = *req.Label
		}
		if req.Url != nil {
			updateData["url"] = req.Url
		}
		if req.Icon != nil {
			updateData["icon"] = req.Icon
		}
		if req.ParentID != nil {
			updateData["parent_id"] = *req.ParentID
		}
		if req.SortOrder != nil {
			updateData["sort_order"] = *req.SortOrder
		}

		updatedMenu, err := menuService.UpdateMenu(id, updateData)
		if err != nil && isNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		}
		if err != nil {
			log.Printf("Error updating menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update menu"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Menu updated", "data": updatedMenu})
	}
}

// deleteMenuHandler DELETE /api/menu/:id
func deleteMenuHandler(menuService services.MenuService) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		err := menuService.DeleteMenu(id)
		if err != nil && isNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Menu not found"})
			return
		}
		if err != nil {
			log.Printf("Error deleting menu: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete menu"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Menu deleted"})
	}
}
