package models

import "time"

// DataLintangKota represents the data_lintang_kota_cms_new table
type DataLintangKota struct {
	ID            int        `json:"id_kota" db:"id_kota"`
	NamaPropinsi  *string    `json:"nama_propinsi" db:"nama_propinsi"`
	NamaKota      *string    `json:"nama_kota" db:"nama_kota"`
	BujurTempat   *string    `json:"bujur_tempat" db:"bujur_tempat"`
	LintangTempat *string    `json:"lintang_tempat" db:"lintang_tempat"`
	TimeZone      *string    `json:"time_zone" db:"time_zone"`
	H             *int       `json:"h" db:"h"`
	TimeCreate    *time.Time `json:"time_create" db:"time_create"`
}

// PrayerSchedule represents the prayer schedule response
type PrayerSchedule struct {
	Tanggal string `json:"tanggal"`
	Imsak   string `json:"imsak"`
	Subuh   string `json:"subuh"`
	Terbit  string `json:"terbit"`
	Dhuha   string `json:"dhuha"`
	Dzuhur  string `json:"dzuhur"`
	Ashar   string `json:"ashar"`
	Maghrib string `json:"maghrib"`
	Isya    string `json:"isya"`
}

// ShalatRequest represents the request parameters
type ShalatRequest struct {
	Prov  string `form:"prov" json:"prov" binding:"required"`
	Kabko string `form:"kabko" json:"kabko"`
	Tgl   string `form:"tgl" json:"tgl" binding:"required"`
}

// ShalatResponse represents the complete response
type ShalatResponse struct {
	*PrayerSchedule
	Prov string `json:"prov"`
	Kota string `json:"kota"`
	Time string `json:"time"`
	Msg  string `json:"msg"`
}

// MonthlyShalatRequest represents request for monthly prayer schedule
type MonthlyShalatRequest struct {
	Thn   string `form:"thn" json:"thn" binding:"required"`
	Bln   string `form:"bln" json:"bln" binding:"required"`
	Prov  string `form:"prov" json:"prov" binding:"required"`
	Kabko string `form:"kabko" json:"kabko"`
}

// MonthlyScheduleItem represents daily prayer schedule in monthly data
type MonthlyScheduleItem struct {
	Date    string `json:"date"`
	Imsak   string `json:"imsak"`
	Subuh   string `json:"subuh"`
	Terbit  string `json:"terbit"`
	Dhuha   string `json:"dhuha"`
	Dzuhur  string `json:"dzuhur"`
	Ashar   string `json:"ashar"`
	Maghrib string `json:"maghrib"`
	Isya    string `json:"isya"`
}

// MonthlyShalatResponse represents the monthly prayer schedule API response
type MonthlyShalatResponse struct {
	Status  int                   `json:"status"`
	Message string                `json:"message"`
	Prov    string                `json:"prov"`
	Kabko   string                `json:"kabko"`
	Data    []MonthlyScheduleItem `json:"data"`
}

// FastingData represents fasting year data from hisab_tgl_puasa table
type FastingData struct {
	Tahun      int    `db:"tgl_tahun"`
	TglHijriah string `db:"tgl_hijriah"`
	TglStart   string `db:"tgl_start"`
	TglEnd     string `db:"tgl_end"`
}

// ImsakiyahRequest represents request for imsakiyah/fasting prayer schedule
type ImsakiyahRequest struct {
	Thn   string `form:"thn" json:"thn" binding:"required"`
	Prov  string `form:"prov" json:"prov" binding:"required"`
	Kabko string `form:"kabko" json:"kabko"`
}

// ImsakiyahScheduleItem represents daily fasting schedule with prayer times
type ImsakiyahScheduleItem struct {
	Date    string `json:"date"`
	Imsak   string `json:"imsak"`
	Subuh   string `json:"subuh"`
	Terbit  string `json:"terbit"`
	Dhuha   string `json:"dhuha"`
	Dzuhur  string `json:"dzuhur"`
	Ashar   string `json:"ashar"`
	Maghrib string `json:"maghrib"`
	Isya    string `json:"isya"`
}

// ImsakiyahResponse represents the imsakiyah/fasting prayer schedule API response
type ImsakiyahResponse struct {
	Status  int                     `json:"status"`
	Message string                  `json:"message"`
	Prov    string                  `json:"prov"`
	Kabko   string                  `json:"kabko"`
	Lintang string                  `json:"lintang,omitempty"`
	Bujur   string                  `json:"bujur,omitempty"`
	Hijriah string                  `json:"hijriah"`
	Tahun   string                  `json:"tahun"`
	Data    []ImsakiyahScheduleItem `json:"data"`
}
