package handlers

import (
	"database/sql"
	"log"
	"net/http"

	"adminbe/internal/app/models"
	"adminbe/internal/pkg/cache"
	"adminbe/internal/pkg/database"

	"github.com/gin-gonic/gin"
)

// listMenuNavigationHandler GET /api/menu_navigation
func listMenuNavigationHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try to get from Redis cache first
		var navigations []models.MenuNavigation
		err := database.Cache.Get(cache.CacheKeyMenuNavigation, &navigations)
		if err == nil {
			// Cache hit
			c.JSON(http.StatusOK, gin.H{"data": navigations, "cached": true})
			return
		}

		// Cache miss - query from DB
		rows, err := db.Query("SELECT id, label, url, icon, children FROM menu_navigation")
		if err != nil {
			log.Printf("Error querying menu_navigation: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menu navigation"})
			return
		}
		defer rows.Close()

		navigations = []models.MenuNavigation{}
		for rows.Next() {
			var mn models.MenuNavigation
			if err := rows.Scan(&mn.ID, &mn.Label, &mn.URL, &mn.Icon, &mn.Children); err != nil {
				log.Printf("Error scanning menu_navigation row: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve menu navigation"})
				return
			}
			navigations = append(navigations, mn)
		}

		// Cache the result
		cacheErr := database.Cache.Set(cache.CacheKeyMenuNavigation, navigations, cache.DefaultNavigationExpiration)
		if cacheErr != nil {
			log.Printf("Warning: Failed to cache menu navigation: %v", cacheErr)
		}

		c.JSON(http.StatusOK, gin.H{"data": navigations, "cached": false})
	}
}
