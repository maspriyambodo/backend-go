package services

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/app/repositories"
)

type ShalatSchedule struct {
	Subuh   string `json:"subuh,omitempty"`
	Dzuhur  string `json:"dzuhur,omitempty"`
	Ashar   string `json:"ashar,omitempty"`
	Maghrib string `json:"maghrib,omitempty"`
	Isya    string `json:"isya,omitempty"`
	Imshak  string `json:"imshak,omitempty"`
	Tanggal string `json:"tanggal,omitempty"`
	Hari    string `json:"hari,omitempty"`
}

type ShalatResponse struct {
	Status  int              `json:"status"`
	Message string           `json:"message,omitempty"`
	Prov    string           `json:"prov,omitempty"`
	Kabko   string           `json:"kabko,omitempty"`
	Tahun   int              `json:"tahun,omitempty"`
	Data    []ShalatSchedule `json:"data,omitempty"`
}

type MonthShalatResponse struct {
	Status  int              `json:"status"`
	Message string           `json:"message"`
	Prov    string           `json:"prov,omitempty"`
	Kabko   string           `json:"kabko,omitempty"`
	Data    []ShalatSchedule `json:"data,omitempty"`
}

type ShalatService interface {
	GetShalat(provID, cityID uint64, date string) (map[string]interface{}, error)
	GetShalatMonth(provID, cityID uint64, year, month int) (*MonthShalatResponse, error)
	GetShalat30Days(provID, cityID uint64, startDate string) (map[string]interface{}, error)
	GetImsakiyah(provID, cityID uint64, year int) (*ShalatResponse, error)
	GetTahunImsakiyah() (map[string]interface{}, error)
	GetProvinsi() ([]map[string]interface{}, error)
	GetKota(provID uint64) ([]map[string]interface{}, error)
}

type shalatService struct {
	repo          repositories.ShalatRepository
	provinceCache map[uint64][]map[string]interface{}
	cityCache     map[uint64][]map[string]interface{}
	imsakCache    map[int]map[string]interface{}
}

// Predefined error messages for consistency
const (
	ErrInvalidParameter = "Invalid parameter"
	ErrDataNotFound     = "Data not found"
	ErrServiceError     = "Service error"
)

func NewShalatService(repo repositories.ShalatRepository) ShalatService {
	return &shalatService{repo: repo}
}

func (s *shalatService) GetShalat(provID, cityID uint64, date string) (map[string]interface{}, error) {
	location, err := s.repo.GetLocationData(provID, cityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location data: %w", err)
	}

	// Parse date
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return nil, fmt.Errorf("invalid date format: %w", err)
	}

	// Calculate prayer times
	schedule := s.calculateShalatPerHari(location, parsedDate)

	// Format date info
	days := []string{"AHAD", "SENIN", "SELASA", "RABU", "KAMIS", "JUMAT", "SABTU"}
	hari := days[int(parsedDate.Weekday())]
	tgl := parsedDate.Format("02")
	blnNames := []string{"JANUARI", "FEBRUARI", "MARET", "APRIL", "MEI", "JUNI", "JULI", "AGUSTUS", "SEPTEMBER", "OKTOBER", "NOVEMBER", "DESEMBER"}
	bln := blnNames[parsedDate.Month()-1]
	thn := parsedDate.Format("2006")

	// Use struct for consistent response
	result := map[string]interface{}{
		"msg":     "sukses",
		"prov":    location.ProvinceTitle,
		"kota":    location.CityTitle,
		"time":    hari + ", " + tgl + " " + bln + " " + thn,
		"subuh":   schedule["subuh"],
		"dzuhur":  schedule["dzuhur"],
		"ashar":   schedule["ashar"],
		"maghrib": schedule["maghrib"],
		"isya":    schedule["isya"],
		"imshak":  schedule["imshak"],
	}
	return result, nil
}

func (s *shalatService) GetShalatMonth(provID, cityID uint64, year, month int) (*MonthShalatResponse, error) {
	location, err := s.repo.GetLocationData(provID, cityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location data: %w", err)
	}

	daysInMonth := time.Date(year, time.Month(month+1), 1, 0, 0, 0, 0, time.UTC).AddDate(0, 0, -1).Day()
	dates := make([]time.Time, daysInMonth)
	schedules := make([]ShalatSchedule, daysInMonth)

	for i := 0; i < daysInMonth; i++ {
		date := time.Date(year, time.Month(month), i+1, 0, 0, 0, 0, time.UTC)
		dates[i] = date
		schedule := s.calculateShalatPerHari(location, date)
		schedules[i] = ShalatSchedule{
			Subuh:   schedule["subuh"],
			Dzuhur:  schedule["dzuhur"],
			Ashar:   schedule["ashar"],
			Maghrib: schedule["maghrib"],
			Isya:    schedule["isya"],
			Imshak:  schedule["imshak"],
			Tanggal: date.Format("2006-01-02"),
			Hari:    getDayNameIndo(date.Weekday()),
		}
	}

	return &MonthShalatResponse{
		Status:  1,
		Message: "Success",
		Prov:    location.ProvinceTitle,
		Kabko:   location.CityTitle,
		Data:    schedules,
	}, nil
}

func (s *shalatService) GetShalat30Days(provID, cityID uint64, startDate string) (map[string]interface{}, error) {
	location, err := s.repo.GetLocationData(provID, cityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location data: %w", err)
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	shalatArray := []map[string]interface{}{}
	for i := 0; i < 30; i++ {
		date := start.AddDate(0, 0, i)
		scheduleMap := s.calculateShalatPerHari(location, date)
		schedule := map[string]interface{}{
			"subuh":   scheduleMap["subuh"],
			"dzuhur":  scheduleMap["dzuhur"],
			"ashar":   scheduleMap["ashar"],
			"maghrib": scheduleMap["maghrib"],
			"isya":    scheduleMap["isya"],
			"imshak":  scheduleMap["imshak"],
			"tanggal": date.Format("2006-01-02"),
			"hari":    getDayNameIndo(date.Weekday()),
		}
		shalatArray = append(shalatArray, schedule)
	}

	return map[string]interface{}{
		"rows": shalatArray,
		"msg":  "sukses",
	}, nil
}

func (s *shalatService) GetImsakiyah(provID, cityID uint64, year int) (*ShalatResponse, error) {
	location, err := s.repo.GetLocationData(provID, cityID)
	if err != nil {
		return nil, fmt.Errorf("failed to get location data: %w", err)
	}

	hisab, err := s.repo.GetHisabTglPuasa(year)
	if err != nil {
		return nil, fmt.Errorf("failed to get hisab data: %w", err)
	}

	if hisab.TglStart == nil || hisab.TglEnd == nil || *hisab.TglHijriah == 0 {
		return &ShalatResponse{
			Status:  0,
			Message: fmt.Sprintf("Jadwal Imsakiyah tahun %d belum ditetapkan", year),
			Data:    []ShalatSchedule{},
		}, nil
	}

	// Note: This would need proper date parsing and imsakiyah calculation, placeholder

	return &ShalatResponse{
		Status:  1,
		Message: "Success",
		Prov:    location.ProvinceTitle,
		Kabko:   location.CityTitle,
		Tahun:   year,
		Data:    []ShalatSchedule{}, // placeholder
	}, nil
}

func (s *shalatService) GetTahunImsakiyah() (map[string]interface{}, error) {
	// Placeholder implementation
	options := []map[string]interface{}{
		{"value": 2024, "text": "1445 H/2024 M"},
		{"value": 2025, "text": "1446 H/2025 M"},
	}
	return map[string]interface{}{"options": options}, nil
}

func (s *shalatService) GetProvinsi() ([]map[string]interface{}, error) {
	provinces, err := s.repo.GetProvinces()
	if err != nil {
		return nil, fmt.Errorf("failed to get provinces: %w", err)
	}

	result := make([]map[string]interface{}, len(provinces))
	for i, p := range provinces {
		result[i] = map[string]interface{}{
			"value": strconv.FormatUint(p.ProvinceID, 10),
			"text":  strings.ToUpper(p.ProvinceTitle),
		}
	}
	return result, nil
}

func (s *shalatService) GetKota(provID uint64) ([]map[string]interface{}, error) {
	cities, err := s.repo.GetCitiesByProvince(provID)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities: %w", err)
	}

	result := make([]map[string]interface{}, len(cities))
	for i, c := range cities {
		result[i] = map[string]interface{}{
			"value": strconv.FormatUint(c.CityID, 10),
			"text":  strings.ToUpper(c.CityTitle),
		}
	}
	return result, nil
}

// Helper functions for prayer time calculations

func DEGtoDECFull(coord string) float64 {
	// Convert degree format (e.g., "6°30'18.43\"S") to decimal degrees
	// Placeholder implementation
	parts := strings.FieldsFunc(coord, func(r rune) bool {
		return r == '°' || r == '\'' || r == '"' || r == 'S' || r == 'N' || r == 'E' || r == 'W'
	})
	if len(parts) < 3 {
		return 0
	}
	deg, _ := strconv.ParseFloat(parts[0], 64)
	min, _ := strconv.ParseFloat(parts[1], 64)
	sec, _ := strconv.ParseFloat(parts[2], 64)
	decimal := deg + min/60 + sec/3600
	if strings.Contains(coord, "S") || strings.Contains(coord, "W") {
		decimal = -decimal
	}
	return decimal
}

func (s *shalatService) calculateShalatPerHari(location *models.DataLintangKotaCmsNew, date time.Time) map[string]string {
	// Basic prayer time calculation placeholders
	// In real implementation, use astronomical calculations
	return map[string]string{
		"subuh":   "04:00",
		"dzuhur":  "12:00",
		"ashar":   "15:00",
		"maghrib": "18:00",
		"isya":    "19:30",
		"imshak":  "03:45",
	}
}

func getDayNameIndo(day time.Weekday) string {
	days := map[time.Weekday]string{
		time.Sunday:    "AHAD",
		time.Monday:    "SENIN",
		time.Tuesday:   "SELASA",
		time.Wednesday: "RABU",
		time.Thursday:  "KAMIS",
		time.Friday:    "JUMAT",
		time.Saturday:  "SABTU",
	}
	return days[day]
}

// More calculation functions needed but placeholders for brevity
