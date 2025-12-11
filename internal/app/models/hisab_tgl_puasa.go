package models

import "time"

// HisabTglPuasa represents the hisab_tgl_puasa table
type HisabTglPuasa struct {
	ID         int        `json:"tgl_id" db:"tgl_id"`
	TglTahun   *int       `json:"tgl_tahun" db:"tgl_tahun"`
	TglStart   *time.Time `json:"tgl_start" db:"tgl_start"`
	TglEnd     *time.Time `json:"tgl_end" db:"tgl_end"`
	TglStatus  *int       `json:"tgl_status" db:"tgl_status"`
	TglHijriah *int       `json:"tgl_hijriah" db:"tgl_hijriah"`
	TimeAdd    *time.Time `json:"time_add" db:"time_add"`
	TimeUpdate *time.Time `json:"time_update" db:"time_update"`
	UserAdd    *int       `json:"user_add" db:"user_add"`
	UserUpdate *int       `json:"user_update" db:"user_update"`
}
