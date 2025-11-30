package handlers

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"adminbe/internal/database"
	"adminbe/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

// listMenuNavigationHandler GET /api/menu_navigation
func listMenuNavigationHandler(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		cacheKey := "menu_navigation_cache"
		rdb := database.RedisClient

		// Try to get from Redis
		cachedData, err := rdb.Get(rdb.Context(), cacheKey).Result()
		if err == nil {
			var navigations []models.MenuNavigation
			if err := json.Unmarshal([]byte(cachedData), &navigations); err == nil {
				c.JSON(http.StatusOK, gin.H{"data": navigations, "cached": true})
				return
			}
		} else if err != redis.Nil {
			log.Printf("Error getting from Redis: %v", err)
		}

		// Query from DB
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

		// Cache in Redis for 10 minutes
		if rdb != nil {
			if data, err := json.Marshal(navigations); err == nil {
				rdb.Set(rdb.Context(), cacheKey, data, 10*time.Minute)
			}
		}

		c.JSON(http.StatusOK, gin.H{"data": navigations, "cached": false})
	}
}
