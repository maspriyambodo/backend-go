package handlers

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine, db *sql.DB) {
	r.GET("/ping", pingHandler)
	r.GET("/health", func(c *gin.Context) { healthHandler(c, db) })

	// User CRUD
	userGroup := r.Group("/api/users")
	{
		userGroup.GET("", listUsersHandler(db))
		userGroup.GET("/:id", getUserHandler(db))
		userGroup.POST("", createUserHandler(db))
		userGroup.PUT("/:id", updateUserHandler(db))
		userGroup.DELETE("/:id", deleteUserHandler(db))
	}

	// Audit Logs CRUD
	auditGroup := r.Group("/api/audit_logs")
	{
		auditGroup.GET("", listAuditLogsHandler(db))
		auditGroup.GET("/:id", getAuditLogHandler(db))
		auditGroup.POST("", createAuditLogHandler(db))
		auditGroup.PUT("/:id", updateAuditLogHandler(db))
		auditGroup.DELETE("/:id", deleteAuditLogHandler(db))
	}

	// Menu CRUD
	menuGroup := r.Group("/api/menu")
	{
		menuGroup.GET("", listMenuHandler(db))
		menuGroup.GET("/:id", getMenuHandler(db))
		menuGroup.POST("", createMenuHandler(db))
		menuGroup.PUT("/:id", updateMenuHandler(db))
		menuGroup.DELETE("/:id", deleteMenuHandler(db))
	}

}

func pingHandler(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

func healthHandler(c *gin.Context, db *sql.DB) {
	if err := db.Ping(); err != nil {
		log.Printf("Database health check failed: %v", err)
		c.JSON(500, gin.H{"status": "error", "message": "DB connection failed"})
		return
	}
	c.JSON(200, gin.H{"status": "ok", "message": "Service is healthy"})
}
