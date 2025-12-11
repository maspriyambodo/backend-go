package handlers

import (
	"adminbe/internal/app/models"
	"adminbe/internal/app/services"
	"crypto/md5"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

// getShalatHandler handles POST /api/apiv1/getShalat - Prayer schedule API
func getShalatHandler(prayerService services.PrayerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse and validate request
		var req models.ShalatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "Invalid request format: " + err.Error()})
			return
		}

		// Input validation
		if req.Prov == "" {
			c.JSON(400, gin.H{"error": "Province (prov) parameter is required"})
			return
		}
		if req.Tgl == "" {
			c.JSON(400, gin.H{"error": "Date (tgl) parameter is required"})
			return
		}

		// Get prayer schedule from service
		response, err := prayerService.GetPrayerSchedule(c.Request.Context(), req.Prov, req.Kabko, req.Tgl)
		if err != nil {
			log.Printf("Error getting prayer schedule: %v", err)
			c.JSON(500, gin.H{"error": "Failed to calculate prayer times"})
			return
		}

		c.JSON(200, response)
	}
}

// getApiProvHandler handles POST /api/apiv1/getApiProv - Get all provinces API
func getApiProvHandler(prayerService services.PrayerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get all provinces from service
		response, err := prayerService.GetAllProvinces(c.Request.Context())
		if err != nil {
			log.Printf("Error getting provinces: %v", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve provinces"})
			return
		}

		c.JSON(200, response)
	}
}

// getApiKabkoHandler handles POST /api/apiv1/getApiKabko - Get cities/regencies by province API
func getApiKabkoHandler(prayerService services.PrayerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse request parameters
		provinceHash := c.PostForm("x")
		if provinceHash == "" {
			// Use default Jakarta hash if no province provided
			provinceHash = fmt.Sprintf("%x", md5.Sum([]byte("13")))
		}

		// Get cities from service
		response, err := prayerService.GetCitiesByProvince(c.Request.Context(), provinceHash)
		if err != nil {
			log.Printf("Error getting cities: %v", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve cities"})
			return
		}

		c.JSON(200, response)
	}
}

// getApiSholatblnHandler handles POST /api/apiv1/getApiSholatbln - Get monthly prayer schedule API
func getApiSholatblnHandler(prayerService services.PrayerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse request parameters (matching PHP POST format)
		year := c.PostForm("thn")
		month := c.PostForm("bln")
		provinceHash := c.PostForm("prov")
		cityHash := c.PostForm("kabko")

		// Get monthly prayer schedule from service
		response, err := prayerService.GetMonthlyPrayerSchedule(
			c.Request.Context(),
			year,
			month,
			provinceHash,
			cityHash,
		)
		if err != nil {
			log.Printf("Error getting monthly prayer schedule: %v", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve monthly prayer schedule"})
			return
		}

		c.JSON(200, response)
	}
}

// getApiimsakiyahHandler handles POST /api/apiv1/getApiimsakiyah - Get fasting/imsakiyah prayer schedule API
func getApiimsakiyahHandler(prayerService services.PrayerService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Parse request parameters (matching PHP POST format)
		year := c.PostForm("thn")
		provinceHash := c.PostForm("prov")
		cityHash := c.PostForm("kabko")

		// Get imsakiyah/fasting prayer schedule from service
		response, err := prayerService.GetImsakiyahSchedule(
			c.Request.Context(),
			year,
			provinceHash,
			cityHash,
		)
		if err != nil {
			log.Printf("Error getting imsakiyah schedule: %v", err)
			c.JSON(500, gin.H{"error": "Failed to retrieve imsakiyah schedule"})
			return
		}

		c.JSON(200, response)
	}
}
