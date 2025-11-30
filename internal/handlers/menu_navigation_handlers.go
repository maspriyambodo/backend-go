package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"adminbe/internal/models"

	"github.com/gin-gonic/gin"
)

// listMenuNavigationHandler GET /api/menu_navigation
func listMenuNavigationHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := db.Query("SELECT id, label, url, icon, children FROM menu_navigation")
		if err != nil {
			log.Printf("Error querying menu_navigation: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menu navigation"})
			return
		}
		defer rows.Close()

		var navigations []models.MenuNavigation
		for rows.Next() {
			var mn models.MenuNavigation
			if err := rows.Scan(&mn.ID, &mn.Label, &mn.URL, &mn.Icon, &mn.Children); err != nil {
				log.Printf("Error scanning menu_navigation row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menu navigation"})
				return
			}
			navigations = append(navigations, mn)
		}
		c.JSON(http.StatusOK, gin.H{"data": navigations})
	}
}
