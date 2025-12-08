package main

import (
	"log"
	"os"

	"adminbe/internal/app/handlers"
	"adminbe/internal/pkg/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("No .env file found, using environment variables: %v", err)
	}

	gin.SetMode(gin.ReleaseMode)

	r := gin.Default()
	r.Use(cors.Default())

	db := database.ConnectDB()
	defer func() {
		sqlDB, _ := db.DB()
		sqlDB.Close()
	}()

	// Initialize JasperServer client
	err = handlers.InitJasperClient("configs/config.yaml")
	if err != nil {
		log.Printf("Failed to initialize JasperServer client: %v", err)
	}

	// Start async audit logging system
	handlers.StartAuditLogger()
	defer handlers.StopAuditLogger()

	handlers.SetupRoutes(r, db)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Println("Server starting on port", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
