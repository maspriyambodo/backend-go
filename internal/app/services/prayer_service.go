package services

import (
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/app/repositories"
)

// PrayerTimes holds calculated prayer times
type PrayerTimes struct {
	Imsak, Subuh, Terbit, Dhuha, Dzuhur, Ashar, Maghrib, Isya string
}

// ProvinceAPIResponse represents a province with hashed ID for API
type ProvinceAPIResponse struct {
	ProvKode string `json:"provKode"`
	ProvNama string `json:"provNama"`
}

// CityAPIResponse represents a city/regency with hashed ID for API
type CityAPIResponse struct {
	KabkoKode string `json:"kabkoKode"`
	KabkoNama string `json:"kabkoNama"`
}

// PrayerService interface defines business logic for prayer calculations
type PrayerService interface {
	GetPrayerSchedule(ctx context.Context, provinceID, cityID, dateStr string) (*models.ShalatResponse, error)
	GetAllProvinces(ctx context.Context) ([]*ProvinceAPIResponse, error)
	GetCitiesByProvince(ctx context.Context, provinceHash string) ([]*CityAPIResponse, error)
	GetMonthlyPrayerSchedule(ctx context.Context, year, month, provinceHash, cityHash string) (*models.MonthlyShalatResponse, error)
	GetImsakiyahSchedule(ctx context.Context, year string, provinceHash, cityHash string) (*models.ImsakiyahResponse, error)
}

// prayerService implements PrayerService
type prayerService struct {
	repo repositories.PrayerRepository
}

// NewPrayerService creates a new prayer service
func NewPrayerService(repo repositories.PrayerRepository) PrayerService {
	return &prayerService{repo: repo}
}

// Indonesian day and month names - initialized once
var idDays = []string{"AHAD", "SENIN", "SELASA", "RABU", "KAMIS", "JUMAT", "SABTU"}
var idMonths = []string{"JANUARI", "FEBRUARI", "MARET", "APRIL", "MEI", "JUNI", "JULI", "AGUSTUS", "SEPTEMBER", "OKTOBER", "NOVEMBER", "DESEMBER"}

// formatIndonesianDate formats date in Indonesian locale
func formatIndonesianDate(dateParsed time.Time) string {
	dayName := idDays[int(dateParsed.Weekday())]
	formattedDate := dayName + ", " + dateParsed.Format("02") + " " + idMonths[dateParsed.Month()-1] + " " + dateParsed.Format("2006")
	return formattedDate
}

// calculatePrayerTimes returns placeholder prayer times (to be implemented with actual astronomical calculations)
func (s *prayerService) calculatePrayerTimes(locationData *repositories.LocationData, dateParsed time.Time) *PrayerTimes {
	// TODO: Implement actual prayer time calculation using jadwal_sholat_perhari logic
	// For now, returning placeholder times based on Indonesian standard times
	return &PrayerTimes{
		Imsak:   "04:30",
		Subuh:   "04:45",
		Terbit:  "06:00",
		Dhuha:   "07:00",
		Dzuhur:  "12:00",
		Ashar:   "15:00",
		Maghrib: "18:00",
		Isya:    "19:30",
	}
}

// GetPrayerSchedule retrieves prayer schedule for given location and date
func (s *prayerService) GetPrayerSchedule(ctx context.Context, provinceID, cityID, dateStr string) (*models.ShalatResponse, error) {
	// Parse and validate date
	dateParsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid date format. Use YYYY-MM-DD: %w", err)
	}

	// Retrieve location data
	locationData, err := s.repo.GetLocationData(ctx, provinceID, cityID)
	if err == sql.ErrNoRows {
		// Return the same error format as PHP
		return &models.ShalatResponse{Msg: "error"}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve location data: %w", err)
	}

	// Apply Jakarta special case
	cityName := locationData.CityName
	if cityID == "192" {
		cityName = "KOTA JAKARTA"
	}

	// Format date in Indonesian locale
	formattedDate := formatIndonesianDate(dateParsed)

	// Calculate prayer times
	prayerTimes := s.calculatePrayerTimes(locationData, dateParsed)

	// Build response
	response := &models.ShalatResponse{
		PrayerSchedule: &models.PrayerSchedule{
			Tanggal: dateStr,
			Imsak:   prayerTimes.Imsak,
			Subuh:   prayerTimes.Subuh,
			Terbit:  prayerTimes.Terbit,
			Dhuha:   prayerTimes.Dhuha,
			Dzuhur:  prayerTimes.Dzuhur,
			Ashar:   prayerTimes.Ashar,
			Maghrib: prayerTimes.Maghrib,
			Isya:    prayerTimes.Isya,
		},
		Prov: locationData.ProvinceName,
		Kota: cityName,
		Time: formattedDate,
		Msg:  "sukses",
	}

	return response, nil
}

// GetImsakiyahSchedule retrieves fasting/imsakiyah prayer schedule (matching PHP getApiimsakiyah)
func (s *prayerService) GetImsakiyahSchedule(ctx context.Context, year string, provinceHash, cityHash string) (*models.ImsakiyahResponse, error) {
	// Convert year string to int for repository
	yearInt := 0
	if year != "" {
		fmt.Sscanf(year, "%d", &yearInt)
	}

	// Get fasting data first
	fastingData, err := s.repo.GetFastingData(ctx, yearInt)
	if err == sql.ErrNoRows {
		return &models.ImsakiyahResponse{
			Status:  0,
			Message: fmt.Sprintf("Jadwal Imsakiyah tahun %s belum ditetapkan", year),
			Data:    []models.ImsakiyahScheduleItem{},
		}, nil
	}
	if err != nil {
		return &models.ImsakiyahResponse{
			Status:  0,
			Message: "Database error",
			Data:    []models.ImsakiyahScheduleItem{},
		}, nil
	}

	// Validate fasting data
	if fastingData.TglHijriah == "" || fastingData.TglStart == "" || fastingData.TglEnd == "" {
		return &models.ImsakiyahResponse{
			Status:  0,
			Message: fmt.Sprintf("Jadwal Imsakiyah tahun %s belum ditetapkan", year),
			Data:    []models.ImsakiyahScheduleItem{},
		}, nil
	}

	// Get location data
	locationData, err := s.repo.GetLocationDataByHashes(ctx, provinceHash, cityHash)
	if err == sql.ErrNoRows {
		return &models.ImsakiyahResponse{
			Status:  0,
			Message: "Error Parameter",
			Data:    []models.ImsakiyahScheduleItem{},
		}, nil
	}
	if err != nil {
		return &models.ImsakiyahResponse{
			Status:  0,
			Message: "Database error",
			Data:    []models.ImsakiyahScheduleItem{},
		}, nil
	}

	// Parameter validation (matching PHP logic)
	if locationData.Latitude == nil || *locationData.Latitude == "" ||
		locationData.Longitude == nil || *locationData.Longitude == "" ||
		locationData.TimeZone == nil || *locationData.TimeZone == "" ||
		year == "" {
		return &models.ImsakiyahResponse{
			Status:  0,
			Message: "Error Parameter",
			Data:    []models.ImsakiyahScheduleItem{},
		}, nil
	}

	// Handle Jakarta special case
	cityName := locationData.CityName
	if cityHash == fmt.Sprintf("%x", md5.Sum([]byte("192"))) {
		cityName = "KOTA JAKARTA"
	}

	// TODO: Implement actual jadwal_imsak_by_date logic
	// For now, generate placeholder fasting schedule for the period
	fastingSchedule := []models.ImsakiyahScheduleItem{}

	// Parse date range
	startDate, err := time.Parse("2006-01-02", fastingData.TglStart)
	if err != nil {
		startDate = time.Now() // fallback
	}
	endDate, err := time.Parse("2006-01-02", fastingData.TglEnd)
	if err != nil {
		endDate = startDate.AddDate(0, 0, 30) // 30 day fallback
	}

	// Generate dates for fasting period
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		dateStr := d.Format("2006-01-02")
		fastingSchedule = append(fastingSchedule, models.ImsakiyahScheduleItem{
			Date:    dateStr,
			Imsak:   "04:30",
			Subuh:   "04:45",
			Terbit:  "06:00",
			Dhuha:   "07:00",
			Dzuhur:  "12:00",
			Ashar:   "15:00",
			Maghrib: "18:00",
			Isya:    "19:30",
		})
	}

	return &models.ImsakiyahResponse{
		Status:  1,
		Message: "Success",
		Prov:    locationData.ProvinceName,
		Kabko:   cityName,
		Hijriah: fastingData.TglHijriah,
		Tahun:   year,
		Data:    fastingSchedule,
	}, nil
}

// GetMonthlyPrayerSchedule retrieves prayer schedule for entire month (matching PHP getApiSholatbln)
func (s *prayerService) GetMonthlyPrayerSchedule(ctx context.Context, year, month, provinceHash, cityHash string) (*models.MonthlyShalatResponse, error) {
	// Retrieve location data using repository
	locationData, err := s.repo.GetLocationDataByHashes(ctx, provinceHash, cityHash)
	if err == sql.ErrNoRows {
		return &models.MonthlyShalatResponse{
			Status:  0,
			Message: "Error Parameter",
			Data:    []models.MonthlyScheduleItem{},
		}, nil
	}
	if err != nil {
		return &models.MonthlyShalatResponse{
			Status:  0,
			Message: "Database error",
			Data:    []models.MonthlyScheduleItem{},
		}, nil
	}

	// Parameter validation (matching PHP logic)
	if locationData.Latitude == nil || *locationData.Latitude == "" ||
		locationData.Longitude == nil || *locationData.Longitude == "" ||
		locationData.TimeZone == nil || *locationData.TimeZone == "" ||
		year == "" || month == "" {
		return &models.MonthlyShalatResponse{
			Status:  0,
			Message: "Error Parameter",
			Data:    []models.MonthlyScheduleItem{},
		}, nil
	}

	// Handle Jakarta special case
	cityName := locationData.CityName
	if cityHash == fmt.Sprintf("%x", md5.Sum([]byte("192"))) {
		cityName = "KOTA JAKARTA"
	}

	// TODO: Implement actual jadwal_sholat_perbulan logic
	// For now, generate placeholder monthly schedule
	monthlyData := []models.MonthlyScheduleItem{}

	// Generate dates for the month (proper date validation)
	for day := 1; day <= 31; day++ {
		dateStr := fmt.Sprintf("%s-%s-%02d", year, month, day)
		if _, err := time.Parse("2006-01-02", dateStr); err != nil {
			break // Stop if invalid date (e.g., Feb 30)
		}

		monthlyData = append(monthlyData, models.MonthlyScheduleItem{
			Date:    dateStr,
			Imsak:   "04:30",
			Subuh:   "04:45",
			Terbit:  "06:00",
			Dhuha:   "07:00",
			Dzuhur:  "12:00",
			Ashar:   "15:00",
			Maghrib: "18:00",
			Isya:    "19:30",
		})
	}

	return &models.MonthlyShalatResponse{
		Status:  1,
		Message: "Success",
		Prov:    locationData.ProvinceName,
		Kabko:   cityName,
		Data:    monthlyData,
	}, nil
}

// GetCitiesByProvince retrieves cities/regencies by province hash (matching PHP getApiKabko special logic)
func (s *prayerService) GetCitiesByProvince(ctx context.Context, provinceHash string) ([]*CityAPIResponse, error) {
	// Special handling for Jakarta (province hash for ID=13)
	jakartaHash := fmt.Sprintf("%x", md5.Sum([]byte("13")))

	if provinceHash == jakartaHash {
		// Return hardcoded Jakarta cities (matching PHP logic)
		return []*CityAPIResponse{
			{
				KabkoKode: fmt.Sprintf("%x", md5.Sum([]byte("192"))),
				KabkoNama: "KOTA JAKARTA",
			},
			{
				KabkoKode: fmt.Sprintf("%x", md5.Sum([]byte("190"))),
				KabkoNama: "KAB. KEPULAUAN SERIBU",
			},
		}, nil
	}

	// For non-Jakarta provinces, get cities from database
	cities, err := s.repo.GetCitiesByProvince(ctx, provinceHash)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cities: %w", err)
	}

	var response []*CityAPIResponse
	for _, city := range cities {
		// Create MD5 hash of city ID (matching PHP md5() function)
		hashedID := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", city.ID))))

		response = append(response, &CityAPIResponse{
			KabkoKode: hashedID,
			KabkoNama: strings.ToUpper(city.Title),
		})
	}

	return response, nil
}

// GetAllProvinces retrieves all provinces with MD5 hashed IDs (matching PHP getApiProv)
func (s *prayerService) GetAllProvinces(ctx context.Context) ([]*ProvinceAPIResponse, error) {
	provinces, err := s.repo.GetAllProvinces(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve provinces: %w", err)
	}

	var response []*ProvinceAPIResponse
	for _, province := range provinces {
		// Create MD5 hash of province ID (matching PHP md5() function)
		hashedID := fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%d", province.ID))))

		response = append(response, &ProvinceAPIResponse{
			ProvKode: hashedID,
			ProvNama: strings.ToUpper(province.Title),
		})
	}

	return response, nil
}
