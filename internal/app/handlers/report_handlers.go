package handlers

import (
	"adminbe/internal/app/models"
	"adminbe/pkg/jasper"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

// JasperClient global instance
var jasperClient *jasper.Client

// InitJasperClient initializes the JasperServer client
func InitJasperClient(configPath string) error {
	config := &models.JasperServerConfig{
		BaseURL:      getEnvOrDefault("JASPER_BASE_URL", "http://localhost:8080/jasperserver"),
		Username:     getEnvOrDefault("JASPER_USERNAME", "jasperadmin"),
		Password:     getEnvOrDefault("JASPER_PASSWORD", "password"),
		Organization: getEnvOrDefault("JASPER_ORGANIZATION", ""),
	}

	jasperClient = jasper.NewClient(config)
	log.Printf("JasperServer client initialized with base URL: %s", config.BaseURL)
	return nil
}

// getEnvOrDefault gets environment variable or returns default value
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// runReportHandler handles report execution requests
func runReportHandler(c *gin.Context) {
	var req models.JasperReportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request format", "details": err.Error()})
		return
	}

	// Execute report
	response, reportData, err := jasperClient.RunReport(&req)
	if err != nil {
		log.Printf("Error running JasperServer report: %v", err)
		c.JSON(500, gin.H{"error": "Failed to run report", "details": err.Error()})
		return
	}

	// For binary content, return the file directly
	if req.OutputFormat == "pdf" || req.OutputFormat == "excel" || req.OutputFormat == "pptx" ||
		req.OutputFormat == "rtf" || req.OutputFormat == "docx" || req.OutputFormat == "xlsx" ||
		req.OutputFormat == "xls" || req.OutputFormat == "png" {

		contentType := "application/octet-stream"
		filename := "report." + req.OutputFormat

		switch req.OutputFormat {
		case "pdf":
			contentType = "application/pdf"
		case "excel", "xlsx", "xls":
			contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		case "pptx":
			contentType = "application/vnd.openxmlformats-officedocument.presentationml.presentation"
		case "docx":
			contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
		case "rtf":
			contentType = "application/rtf"
		case "png":
			contentType = "image/png"
		}

		c.Header("Content-Disposition", "attachment; filename="+filename)
		c.Header("Content-Type", contentType)
		c.Data(200, contentType, reportData)
		return
	}

	// For HTML/JSON content, return JSON response
	c.JSON(200, response)
}

// getServerInfoHandler retrieves JasperServer server information
func getServerInfoHandler(c *gin.Context) {
	info, err := jasperClient.GetServerInfo()
	if err != nil {
		log.Printf("Error getting JasperServer info: %v", err)
		c.JSON(500, gin.H{"error": "Failed to get server info", "details": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"server_info": info,
		"status":      "success",
	})
}

// health check for JasperServer
func jasperHealthHandler(c *gin.Context) {
	_, err := jasperClient.GetServerInfo()
	if err != nil {
		log.Printf("JasperServer health check failed: %v", err)
		c.JSON(500, gin.H{
			"status":  "error",
			"message": "JasperServer connection failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(200, gin.H{
		"status":  "ok",
		"message": "JasperServer is healthy",
	})
}
