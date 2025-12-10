package handlers

import (
	"adminbe/internal/app/middleware"
	"adminbe/internal/app/repositories"
	"adminbe/internal/app/services"
	"adminbe/internal/pkg/database"
	"adminbe/internal/pkg/utils"
	"database/sql"
	"encoding/json"
	"log"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// handleServiceError handles common service error patterns
func handleServiceError(c *gin.Context, err error, operation string) bool {
	return utils.HandleError(c, err, operation)
}

// bindJSONRequest binds JSON request and handles validation errors
func bindJSONRequest(c *gin.Context, req interface{}) bool {
	if err := c.ShouldBindJSON(req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return false
	}
	return true
}

// getUserIDFromContext extracts user ID from Gin context
func getUserIDFromContext(c *gin.Context) *uint64 {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		return nil
	}

	if userID, ok := userIDVal.(uint64); ok {
		return &userID
	}

	return nil
}

// logAuditEntry creates an audit log entry (helper for consistency)
func logAuditEntry(c *gin.Context, eventType, tableName string, recordID uint64, oldValues, newValues interface{}, db *sql.DB) {
	if auditLogChan == nil {
		return
	}

	userIDPtr := getUserIDFromContext(c)
	if userIDPtr == nil {
		log.Printf("Warning: cannot create audit log without user ID for %s %s %d", eventType, tableName, recordID)
		return
	}

	select {
	case auditLogChan <- auditLogEntry{
		UserID:    *userIDPtr,
		Event:     eventType,
		Table:     tableName,
		RecordID:  recordID,
		OldValues: oldValues,
		NewValues: newValues,
		DB:        db,
	}:
	default:
		log.Printf("Warning: audit log queue full, dropping %s audit for %s %d", eventType, tableName, recordID)
	}
}

// isNotFoundError checks if error is a public not found error
func isNotFoundError(err error) bool {
	return utils.IsNotFound(err)
}

func SetupRoutes(r *gin.Engine, db *gorm.DB) {
	sqlDB, _ := db.DB()

	// Dependency injection setup
	userRepo := repositories.NewUserRepository(sqlDB)
	userService := services.NewUserService(userRepo)

	menuRepo := repositories.NewMenuRepository(sqlDB)
	menuService := services.NewMenuService(menuRepo)

	roleRepo := repositories.NewRoleRepository(sqlDB)
	roleService := services.NewRoleService(roleRepo)

	roleInheritanceRepo := repositories.NewRoleInheritanceRepository(sqlDB)
	services.NewRoleInheritanceService(roleInheritanceRepo)

	roleMenuRepo := repositories.NewRoleMenuRepository(sqlDB)
	services.NewRoleMenuService(roleMenuRepo)

	userMenuRepo := repositories.NewUserMenuRepository(sqlDB)
	services.NewUserMenuService(userMenuRepo)

	userRoleRepo := repositories.NewUserRoleRepository(sqlDB)
	services.NewUserRoleService(userRoleRepo)

	// Shalat services
	shalatRepo := repositories.NewShalatRepository(sqlDB)
	shalatService := services.NewShalatService(shalatRepo)

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
			userGroup.PUT("/:id", updateUserHandler(userService, sqlDB))
			userGroup.DELETE("/:id", deleteUserHandler(userService, sqlDB))
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
			menuGroup.GET("", listMenuHandler(menuService))
			menuGroup.GET("/:id", getMenuHandler(menuService))
			menuGroup.POST("", createMenuHandler(menuService, sqlDB))
			menuGroup.PUT("/:id", updateMenuHandler(menuService, sqlDB))
			menuGroup.DELETE("/:id", deleteMenuHandler(menuService, sqlDB))
		}

		// Roles CRUD
		rolesGroup := apiGroup.Group("/roles")
		{
			rolesGroup.GET("", listRolesHandler(roleService))
			rolesGroup.GET("/:id", getRoleHandler(roleService))
			rolesGroup.POST("", createRoleHandler(roleService, sqlDB))
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

		// Shalat group (public for prayer times)
		shalatGroup := r.Group("/api/v1")
		{
			shalatGroup.POST("/getShalat", getShalatHandler(shalatService))
			shalatGroup.POST("/getsholatbln", getShalatBlnHandler(shalatService))
			shalatGroup.POST("/getShalat30", getShalat30Handler(shalatService))
			shalatGroup.POST("/getimsakiyah", getImsakiyahHandler(shalatService))
			shalatGroup.POST("/getTahunimsak", getTahunImsakHandler(shalatService))
			shalatGroup.POST("/getProv", getProvHandler(shalatService))
			shalatGroup.POST("/getKabko", getKabkoHandler(shalatService))
			shalatGroup.POST("/getApiProv", getApiProvHandler(shalatService))
			shalatGroup.POST("/getApiKabko", getApiKabkoHandler(shalatService))
			shalatGroup.POST("/getApiSholatbln", getApiSholatblnHandler(shalatService))
			shalatGroup.POST("/getApiimsakiyah", getApiImsakiyahHandler(shalatService))
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

// createAuditLog creates an audit log entry (deprecated - use logAuditEntry with Gin context instead)
func createAuditLog(db *sql.DB, userIDPtr *uint64, eventType string, tableName string, recordID uint64, oldValues interface{}, newValues interface{}) {
	userID := uint64(0) // Default fallback ID
	if userIDPtr != nil {
		userID = *userIDPtr
	}

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
