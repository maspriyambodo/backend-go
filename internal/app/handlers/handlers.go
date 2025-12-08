package handlers

import (
	"adminbe/internal/app/middleware"
	"adminbe/internal/app/repositories"
	"adminbe/internal/app/services"
	"adminbe/internal/pkg/database"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	sqlDB, _ := db.DB()

	// Dependency injection setup
	userRepo := repositories.NewUserRepository(sqlDB)
	userService := services.NewUserService(userRepo)

	// Global middleware
	r.Use(middleware.CustomRecoveryMiddleware())
	r.Use(middleware.RequestLoggerMiddleware(sqlDB))
	r.Use(middleware.SecurityHeadersMiddleware())

	r.GET("/ping", pingHandler)
	r.GET("/health", func(c *gin.Context) { healthHandler(c, db) })

	// Auth routes (public)
	authGroup := r.Group("/api/auth")
	{
		authGroup.POST("/login", loginHandler(db))
	}

	// Protected API routes
	apiGroup := r.Group("/api")
	apiGroup.Use(middleware.AuthMiddleware())
	{
		// User CRUD
		userGroup := apiGroup.Group("/users")
		{
			userGroup.GET("", listUsersHandler(userService))
			userGroup.GET("/:id", getUserHandler(userService))
			userGroup.POST("", createUserHandler(userService, sqlDB))
			userGroup.PUT("/:id", updateUserHandler(userService))
			userGroup.DELETE("/:id", deleteUserHandler(userService))
		}

		// Audit Logs CRUD
		auditGroup := apiGroup.Group("/audit_logs")
		{
			auditGroup.GET("", listAuditLogsHandler(sqlDB))
			auditGroup.GET("/:id", getAuditLogHandler(sqlDB))
			auditGroup.POST("", createAuditLogHandler(sqlDB))
			auditGroup.PUT("/:id", updateAuditLogHandler(sqlDB))
			auditGroup.DELETE("/:id", deleteAuditLogHandler(sqlDB))
		}

		// Menu CRUD
		menuGroup := apiGroup.Group("/menu")
		{
			menuGroup.GET("", listMenuHandler(sqlDB))
			menuGroup.GET("/:id", getMenuHandler(sqlDB))
			menuGroup.POST("", createMenuHandler(sqlDB))
			menuGroup.PUT("/:id", updateMenuHandler(sqlDB))
			menuGroup.DELETE("/:id", deleteMenuHandler(sqlDB))
		}

		// Roles CRUD
		rolesGroup := apiGroup.Group("/roles")
		{
			rolesGroup.GET("", listRolesHandler(sqlDB))
			rolesGroup.GET("/:id", getRoleHandler(sqlDB))
			rolesGroup.POST("", createRoleHandler(sqlDB))
			rolesGroup.PUT("/:id", updateRoleHandler(sqlDB))
			rolesGroup.DELETE("/:id", deleteRoleHandler(sqlDB))
		}

		// Role Inheritances CRUD
		inheritancesGroup := apiGroup.Group("/role_inheritances")
		{
			inheritancesGroup.GET("", listRoleInheritancesHandler(sqlDB))
			inheritancesGroup.GET("/:id", getRoleInheritanceHandler(sqlDB))
			inheritancesGroup.POST("", createRoleInheritanceHandler(sqlDB))
			inheritancesGroup.PUT("/:id", updateRoleInheritanceHandler(sqlDB))
			inheritancesGroup.DELETE("/:id", deleteRoleInheritanceHandler(sqlDB))
		}

		// V Roles (view for role hierarchies)
		vRolesGroup := apiGroup.Group("/v_roles")
		{
			vRolesGroup.GET("", listVRolesHandler(sqlDB))
		}

		// Role Menu CRUD
		roleMenuGroup := apiGroup.Group("/role_menu")
		{
			roleMenuGroup.GET("", listRoleMenusHandler(sqlDB))
			roleMenuGroup.GET("/:roleId/:menuId", getRoleMenuHandler(sqlDB))
			roleMenuGroup.POST("", createRoleMenuHandler(sqlDB))
			roleMenuGroup.PUT("/:roleId/:menuId", updateRoleMenuHandler(sqlDB))
			roleMenuGroup.DELETE("/:roleId/:menuId", deleteRoleMenuHandler(sqlDB))
		}

		// Menu Navigation (view for menu tree)
		menuNavigationGroup := apiGroup.Group("/menu_navigation")
		{
			menuNavigationGroup.GET("", listMenuNavigationHandler(sqlDB))
		}

		// User Menu CRUD
		userMenuGroup := apiGroup.Group("/user_menu")
		{
			userMenuGroup.GET("", listUserMenusHandler(sqlDB))
			userMenuGroup.GET("/:userId/:menuId", getUserMenuHandler(sqlDB))
			userMenuGroup.POST("", createUserMenuHandler(sqlDB))
			userMenuGroup.PUT("/:userId/:menuId", updateUserMenuHandler(sqlDB))
			userMenuGroup.DELETE("/:userId/:menuId", deleteUserMenuHandler(sqlDB))
		}

		// User Roles CRUD
		userRolesGroup := apiGroup.Group("/user_roles")
		{
			userRolesGroup.GET("", listUserRolesHandler(sqlDB))
			userRolesGroup.GET("/:userId/:roleId", getUserRoleHandler(sqlDB))
			userRolesGroup.POST("", createUserRoleHandler(sqlDB))
			userRolesGroup.PUT("/:userId/:roleId", updateUserRoleHandler(sqlDB))
			userRolesGroup.DELETE("/:userId/:roleId", deleteUserRoleHandler(sqlDB))
		}

		// Reports group
		reportsGroup := apiGroup.Group("/reports")
		{
			reportsGroup.POST("/run", runReportHandler)
			reportsGroup.GET("/server-info", getServerInfoHandler)
			reportsGroup.GET("/health", jasperHealthHandler)
		}
	}
}

func pingHandler(c *gin.Context) {
	c.JSON(200, map[string]string{
		"message": "pong",
	})
}

func healthHandler(c *gin.Context, db *gorm.DB) {
	dbHealthy := true
	redisHealthy := false

	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		log.Printf("Database health check failed: %v", err)
		dbHealthy = false
	}

	if database.RedisClient != nil {
		_, err := database.RedisClient.Ping(database.RedisClient.Context()).Result()
		if err == nil {
			redisHealthy = true
		}
	}

	if !dbHealthy {
		c.JSON(500, gin.H{"status": "error", "message": "DB connection failed", "redis": redisHealthy})
		return
	}

	c.JSON(200, gin.H{"status": "ok", "message": "Service is healthy", "redis": redisHealthy})
}

// createAuditLog creates an audit log entry
func createAuditLog(db *sql.DB, userID *uint64, eventType string, tableName string, recordID uint64, oldValues interface{}, newValues interface{}) {
	var oldJSON, newJSON []byte
	if oldValues != nil {
		oldJSON, _ = json.Marshal(oldValues)
	}
	if newValues != nil {
		newJSON, _ = json.Marshal(newValues)
	}

	// Insert audit log
	_, err := db.Exec("INSERT INTO audit_logs (user_id, event_type, table_name, record_id, old_values, new_values) VALUES (?, ?, ?, ?, ?, ?)",
		userID, eventType, tableName, recordID, oldJSON, newJSON)
	if err != nil {
		log.Printf("Error creating audit log: %v", err)
	}
}
