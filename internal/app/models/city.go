package models

// City represents the app_city table
type City struct {
	ID        int     `json:"id" db:"city_id"`
	Title     *string `json:"title" db:"city_title"`
	Province  int     `json:"province" db:"city_province"`
	CityIDNew int     `json:"city_id_new" db:"city_id_new"`
}
