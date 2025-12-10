package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"adminbe/internal/app/services"

	"github.com/gin-gonic/gin"
)

func getShalatHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		prov := c.PostForm("prov")
		kabko := c.PostForm("kabko")
		date := c.PostForm("tgl")

		if date == "" || prov == "" {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "error", "error": "prov and tgl parameters are required"})
			return
		}

		provID, err := parseProvID(prov)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "error", "error": "invalid prov parameter"})
			return
		}

		cityID := uint64(0) // Default to 0 if not provided
		if kabko != "" {
			if cityID, err = parseCityID(kabko); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"msg": "error", "error": "invalid kabko parameter"})
				return
			}
		}

		result, err := shalatService.GetShalat(provID, cityID, date)
		if err != nil {
			handleServiceError(c, err, "get shalat")
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func getShalatBlnHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		thn := c.PostForm("thn")
		bln := c.PostForm("bln")
		prov := c.PostForm("prov")
		kabko := c.PostForm("kabko")

		year, err := strconv.Atoi(thn)
		if err != nil || year == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
			return
		}

		month, err := strconv.Atoi(bln)
		if err != nil || month == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
			return
		}

		provID, err := parseProvID(prov)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
			return
		}

		var cityID uint64
		if kabko != "" {
			cityID, err = parseCityID(kabko)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
				return
			}
		}

		result, err := shalatService.GetShalatMonth(provID, cityID, year, month) // months are 0-based in Go
		if err != nil {
			handleServiceError(c, err, "get shalat month")
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func getShalat30Handler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		prov := c.PostForm("prov")
		kabko := c.PostForm("kabko")
		date := c.PostForm("tgl")

		if date == "" || prov == "" {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "error"})
			return
		}

		provID, err := parseProvID(prov)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"msg": "error"})
			return
		}

		var cityID uint64
		if kabko != "" {
			cityID, err = parseCityID(kabko)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"msg": "error"})
				return
			}
		}

		result, err := shalatService.GetShalat30Days(provID, cityID, date)
		if err != nil {
			handleServiceError(c, err, "get shalat 30 days")
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func getImsakiyahHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		thn := c.PostForm("thn")
		prov := c.PostForm("prov")
		kabko := c.PostForm("kabko")

		year, err := strconv.Atoi(thn)
		if err != nil || year == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
			return
		}

		provID, err := parseProvID(prov)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
			return
		}

		var cityID uint64
		if kabko != "" {
			cityID, err = parseCityID(kabko)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
				return
			}
		}

		result, err := shalatService.GetImsakiyah(provID, cityID, year)
		if err != nil {
			handleServiceError(c, err, "get imsakiyah")
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func getTahunImsakHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := shalatService.GetTahunImsakiyah()
		if err != nil {
			handleServiceError(c, err, "get tahun imsakiyah")
			return
		}

		c.String(http.StatusOK, buildOptionString(result["options"].([]map[string]interface{})))
	}
}

func getProvHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		result, err := shalatService.GetProvinsi()
		if err != nil {
			handleServiceError(c, err, "get provinces")
			return
		}

		c.String(http.StatusOK, buildOptionString(result))
	}
}

func getKabkoHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		x := c.PostForm("x")

		provID, err := parseProvID(x)
		if err != nil {
			provID = 13 // default
		}

		result, err := shalatService.GetKota(provID)
		if err != nil {
			handleServiceError(c, err, "get cities")
			return
		}

		// Add special handling for DKI Jakarta
		if provID == 13 {
			// Insert Jakarta options
		}

		c.String(http.StatusOK, buildOptionString(result))
	}
}

func getApiProvHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		provinces, err := shalatService.GetProvinsi()
		if err != nil {
			handleServiceError(c, err, "get provinces")
			return
		}

		result := make([]map[string]interface{}, len(provinces))
		for i, p := range provinces {
			result[i] = map[string]interface{}{
				"provKode": p["value"].(string),
				"provNama": p["text"].(string),
			}
		}

		c.JSON(http.StatusOK, result)
	}
}

func getApiKabkoHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		x := c.PostForm("x")

		provID, err := strconv.ParseUint(x, 10, 64)
		if err != nil {
			provID = 13 // default
		}

		cities, err := shalatService.GetKota(provID)
		if err != nil {
			handleServiceError(c, err, "get cities")
			return
		}

		result := make([]map[string]interface{}, len(cities))
		for i, city := range cities {
			result[i] = map[string]interface{}{
				"kabkoKode": city["value"].(string),
				"kabkoNama": city["text"].(string),
			}
		}

		// Special handling for Jakarta
		if provID == 13 {
			result = append([]map[string]interface{}{
				{"kabkoKode": "192", "kabkoNama": "KOTA JAKARTA"},
				{"kabkoKode": "190", "kabkoNama": "KAB. KEPULAUAN SERIBU"},
			}, result...)
		}

		c.JSON(http.StatusOK, result)
	}
}

func getApiSholatblnHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		thn := c.PostForm("thn")
		bln := c.PostForm("bln")
		prov := c.PostForm("prov")
		kabko := c.PostForm("kabko")

		year, err := strconv.Atoi(thn)
		if err != nil || year == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
			return
		}

		month, err := strconv.Atoi(bln)
		if err != nil || month == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
			return
		}

		provID, err := strconv.ParseUint(prov, 10, 64)
		if err != nil {
			provID = 0
		}
		var cityID uint64
		if kabko != "" {
			cityID, err = strconv.ParseUint(kabko, 10, 64)
			if err != nil {
				cityID = 0
			}
		}

		result, err := shalatService.GetShalatMonth(provID, cityID, year, month)
		if err != nil {
			handleServiceError(c, err, "get shalat month")
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

func getApiImsakiyahHandler(shalatService services.ShalatService) gin.HandlerFunc {
	return func(c *gin.Context) {
		thn := c.PostForm("thn")
		prov := c.PostForm("prov")
		kabko := c.PostForm("kabko")

		year, err := strconv.Atoi(thn)
		if err != nil || year == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"status": 0, "message": "Error Parameter", "data": []interface{}{}})
			return
		}

		provID, err := strconv.ParseUint(prov, 10, 64)
		if err != nil {
			provID = 0
		}
		var cityID uint64
		if kabko != "" {
			cityID, err = strconv.ParseUint(kabko, 10, 64)
			if err != nil {
				cityID = 0
			}
		}

		result, err := shalatService.GetImsakiyah(provID, cityID, year)
		if err != nil {
			handleServiceError(c, err, "get imsakiyah")
			return
		}

		c.JSON(http.StatusOK, result)
	}
}

// Helper functions

func parseProvID(param string) (uint64, error) {
	return strconv.ParseUint(param, 10, 64)
}

func parseCityID(param string) (uint64, error) {
	return strconv.ParseUint(param, 10, 64)
}

func buildOptionString(options []map[string]interface{}) string {
	if len(options) == 0 {
		return ""
	}

	var result strings.Builder
	result.Grow(len(options) * 50) // Pre-allocate approximate size
	for _, opt := range options {
		result.WriteString("<option value='")
		// Handle both int and string values
		switch value := opt["value"].(type) {
		case int:
			result.WriteString(strconv.Itoa(value))
		case string:
			result.WriteString(value)
		default:
			result.WriteString(fmt.Sprintf("%v", value))
		}
		result.WriteString("'>")
		result.WriteString(opt["text"].(string))
		result.WriteString("</option>")
	}
	return result.String()
}
