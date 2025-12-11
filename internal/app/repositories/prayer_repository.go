package repositories

import (
	"adminbe/internal/app/models"
	"context"
	"database/sql"
	"fmt"
)

// LocationData holds location information for prayer calculations
type LocationData struct {
	ID           int     `db:"id_kota"`
	Latitude     *string `db:"lintang_tempat"`
	Longitude    *string `db:"bujur_tempat"`
	TimeZone     *string `db:"time_zone"`
	Elevation    *int    `db:"h"`
	ProvinceName string  `db:"province_name"`
	CityName     string  `db:"city_name"`
}

// ProvinceData holds province information
type ProvinceData struct {
	ID    int    `db:"province_id"`
	Title string `db:"province_title"`
}

// CityData holds city/regency information
type CityData struct {
	ID         int    `db:"city_id"`
	ProvinceID int    `db:"city_province"`
	Title      string `db:"city_title"`
}

// PrayerRepository interface defines data access methods for prayer calculations
type PrayerRepository interface {
	GetLocationData(ctx context.Context, provinceID, cityID string) (*LocationData, error)
	GetAllProvinces(ctx context.Context) ([]*ProvinceData, error)
	GetCitiesByProvince(ctx context.Context, provinceHash string) ([]*CityData, error)
	GetLocationDataByHashes(ctx context.Context, provinceHash, cityHash string) (*LocationData, error)
	GetFastingData(ctx context.Context, year int) (*models.FastingData, error)
}

// prayerRepository implements PrayerRepository
type prayerRepository struct {
	db *sql.DB
}

// NewPrayerRepository creates a new prayer repository
func NewPrayerRepository(db *sql.DB) PrayerRepository {
	return &prayerRepository{db: db}
}

// GetLocationData retrieves location data required for prayer calculations
func (r *prayerRepository) GetLocationData(ctx context.Context, provinceID, cityID string) (*LocationData, error) {
	query := `
		SELECT dlk.id_kota, dlk.lintang_tempat, dlk.bujur_tempat, dlk.time_zone, dlk.h,
			   UPPER(p.province_title) as province_name, UPPER(c.city_title) as city_name
		FROM data_lintang_kota_cms_new dlk
		JOIN app_province p ON p.province_id = dlk.nama_propinsi
		JOIN app_city c ON c.city_id = dlk.nama_kota
		WHERE 1=1
	`

	var args []interface{}

	if provinceID != "" {
		query += " AND dlk.nama_propinsi = ?"
		args = append(args, provinceID)
	}

	if cityID != "" {
		query += " AND dlk.nama_kota = ?"
		args = append(args, cityID)
	}

	query += " LIMIT 1"

	var locationData LocationData
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&locationData.ID,
		&locationData.Latitude,
		&locationData.Longitude,
		&locationData.TimeZone,
		&locationData.Elevation,
		&locationData.ProvinceName,
		&locationData.CityName,
	)

	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get location data: %w", err)
	}

	return &locationData, nil
}

// GetFastingData retrieves fasting year data by year
func (r *prayerRepository) GetFastingData(ctx context.Context, year int) (*models.FastingData, error) {
	query := `
		SELECT tgl_tahun, tgl_hijriah, tgl_start, tgl_end
		FROM hisab_tgl_puasa
		WHERE tgl_tahun = ?
		LIMIT 1
	`

	var fastingData models.FastingData
	err := r.db.QueryRowContext(ctx, query, year).Scan(
		&fastingData.Tahun,
		&fastingData.TglHijriah,
		&fastingData.TglStart,
		&fastingData.TglEnd,
	)

	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get fasting data: %w", err)
	}

	return &fastingData, nil
}

// GetAllProvinces retrieves all provinces ordered by ID
func (r *prayerRepository) GetAllProvinces(ctx context.Context) ([]*ProvinceData, error) {
	query := `
		SELECT province_id, province_title
		FROM app_province
		ORDER BY province_id ASC
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get provinces: %w", err)
	}
	defer rows.Close()

	var provinces []*ProvinceData
	for rows.Next() {
		var province ProvinceData
		err := rows.Scan(&province.ID, &province.Title)
		if err != nil {
			return nil, fmt.Errorf("failed to scan province: %w", err)
		}
		provinces = append(provinces, &province)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating provinces: %w", err)
	}

	return provinces, nil
}

// GetCitiesByProvince retrieves cities by province hash (matching PHP getApiKabko logic)
func (r *prayerRepository) GetCitiesByProvince(ctx context.Context, provinceHash string) ([]*CityData, error) {
	query := `
		SELECT city_id, city_province, city_title
		FROM app_city
		WHERE MD5(city_province) = ?
		ORDER BY city_id ASC
	`

	rows, err := r.db.QueryContext(ctx, query, provinceHash)
	if err != nil {
		return nil, fmt.Errorf("failed to get cities: %w", err)
	}
	defer rows.Close()

	var cities []*CityData
	for rows.Next() {
		var city CityData
		err := rows.Scan(&city.ID, &city.ProvinceID, &city.Title)
		if err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, &city)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating cities: %w", err)
	}

	return cities, nil
}

// GetLocationDataByHashes retrieves location data using MD5 hashes (matching PHP getApiSholatbln)
func (r *prayerRepository) GetLocationDataByHashes(ctx context.Context, provinceHash, cityHash string) (*LocationData, error) {
	query := `
		SELECT dlk.id_kota, dlk.lintang_tempat, dlk.bujur_tempat, dlk.time_zone, dlk.h,
			   UPPER(p.province_title) as province_title, UPPER(c.city_title) as city_title
		FROM data_lintang_kota_cms_new dlk
		JOIN app_province p ON p.province_id = dlk.nama_propinsi
		JOIN app_city c ON c.city_id = dlk.nama_kota
		WHERE 1=1
	`

	args := []interface{}{}
	if provinceHash != "" {
		query += " AND MD5(p.province_id) = ?"
		args = append(args, provinceHash)
	}
	if cityHash != "" {
		query += " AND MD5(c.city_id) = ?"
		args = append(args, cityHash)
	}
	query += " LIMIT 1"

	var locationData LocationData
	err := r.db.QueryRowContext(ctx, query, args...).Scan(
		&locationData.ID,
		&locationData.Latitude,
		&locationData.Longitude,
		&locationData.TimeZone,
		&locationData.Elevation,
		&locationData.ProvinceName,
		&locationData.CityName,
	)

	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get location data by hashes: %w", err)
	}

	return &locationData, nil
}
