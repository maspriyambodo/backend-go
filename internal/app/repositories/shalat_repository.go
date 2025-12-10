package repositories

import (
	"database/sql"
	"fmt"

	"adminbe/internal/app/models"
)

// ShalatRepository interface defines data access methods for shalat calculations
type ShalatRepository interface {
	GetLocationData(provID, cityID uint64) (*models.DataLintangKotaCmsNew, error)
	GetHisabTglPuasa(year int) (*models.HisabTglPuasa, error)
	GetProvinces() ([]models.AppProvince, error)
	GetCitiesByProvince(provID uint64) ([]models.AppCity, error)
}

// shalatRepository implements ShalatRepository
type shalatRepository struct {
	db *sql.DB
}

// NewShalatRepository creates a new shalat repository
func NewShalatRepository(db *sql.DB) ShalatRepository {
	return &shalatRepository{db: db}
}

// GetLocationData retrieves location data for a given province and city
func (r *shalatRepository) GetLocationData(provID, cityID uint64) (*models.DataLintangKotaCmsNew, error) {
	var data models.DataLintangKotaCmsNew
	query := `
		SELECT dlk.id_kota, dlk.nama_propinsi, dlk.nama_kota, dlk.lintang_tempat, dlk.bujur_tempat, dlk.time_zone, dlk.h, dlk.time_create,
		       UPPER(ap.province_title) as province_title, UPPER(ac.city_title) as city_title
		FROM data_lintang_kota_cms_new dlk
		JOIN app_province ap ON ap.province_id = dlk.nama_propinsi
		JOIN app_city ac ON ac.city_id = dlk.nama_kota
		WHERE ap.province_id = ? AND ac.city_id = ?`

	if cityID == 0 {
		query = `
		SELECT dlk.id_kota, dlk.nama_propinsi, dlk.nama_kota, dlk.lintang_tempat, dlk.bujur_tempat, dlk.time_zone, dlk.h, dlk.time_create,
		       UPPER(ap.province_title) as province_title, UPPER(ac.city_title) as city_title
		FROM data_lintang_kota_cms_new dlk
		JOIN app_province ap ON ap.province_id = dlk.nama_propinsi
		JOIN app_city ac ON ac.city_id = dlk.nama_kota
		WHERE ap.province_id = ?`
		row := r.db.QueryRow(query, provID)
		err := row.Scan(&data.IDKota, &data.NamaPropinsi, &data.NamaKota, &data.LintangTempat, &data.BujurTempat, &data.TimeZone, &data.H, &data.TimeCreate, &data.ProvinceTitle, &data.CityTitle)
		if err == sql.ErrNoRows {
			return nil, err
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan location data: %w", err)
		}
		return &data, nil
	} else {
		row := r.db.QueryRow(query, provID, cityID)
		err := row.Scan(&data.IDKota, &data.NamaPropinsi, &data.NamaKota, &data.LintangTempat, &data.BujurTempat, &data.TimeZone, &data.H, &data.TimeCreate, &data.ProvinceTitle, &data.CityTitle)
		if err == sql.ErrNoRows {
			return nil, err
		}
		if err != nil {
			return nil, fmt.Errorf("failed to scan location data: %w", err)
		}
		return &data, nil
	}
}

// GetHisabTglPuasa retrieves hisab data for a given year
func (r *shalatRepository) GetHisabTglPuasa(year int) (*models.HisabTglPuasa, error) {
	var hisab models.HisabTglPuasa
	row := r.db.QueryRow("SELECT tgl_tahun, tgl_hijriah, tgl_start, tgl_end FROM hisab_tgl_puasa WHERE tgl_tahun = ?", year)
	err := row.Scan(&hisab.TglTahun, &hisab.TglHijriah, &hisab.TglStart, &hisab.TglEnd)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan hisab data: %w", err)
	}
	return &hisab, nil
}

// GetProvinces retrieves all provinces
func (r *shalatRepository) GetProvinces() ([]models.AppProvince, error) {
	rows, err := r.db.Query("SELECT province_id, province_title FROM app_province ORDER BY province_id ASC")
	if err != nil {
		return nil, fmt.Errorf("failed to query provinces: %w", err)
	}
	defer rows.Close()

	var provinces []models.AppProvince
	for rows.Next() {
		var p models.AppProvince
		if err := rows.Scan(&p.ProvinceID, &p.ProvinceTitle); err != nil {
			return nil, fmt.Errorf("failed to scan province: %w", err)
		}
		provinces = append(provinces, p)
	}
	return provinces, nil
}

// GetCitiesByProvince retrieves cities for a given province
func (r *shalatRepository) GetCitiesByProvince(provID uint64) ([]models.AppCity, error) {
	rows, err := r.db.Query("SELECT city_id, city_title, city_province FROM app_city WHERE city_province = ? ORDER BY city_id ASC", provID)
	if err != nil {
		return nil, fmt.Errorf("failed to query cities: %w", err)
	}
	defer rows.Close()

	var cities []models.AppCity
	for rows.Next() {
		var c models.AppCity
		if err := rows.Scan(&c.CityID, &c.CityTitle, &c.CityProvince); err != nil {
			return nil, fmt.Errorf("failed to scan city: %w", err)
		}
		cities = append(cities, c)
	}
	return cities, nil
}
