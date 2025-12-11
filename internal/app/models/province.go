package models

// Province represents the app_province table
type Province struct {
	ID            int    `json:"id" db:"province_id"`
	Title         string `json:"title" db:"province_title"`
	ProvinceIDNew int    `json:"province_id_new" db:"province_id_new"`
}
