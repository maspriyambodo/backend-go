package models

import (
	"time"
)

// Menu represents the menu table
type Menu struct {
	ID        uint       `json:"id" db:"id"`
	Label     string     `json:"label" db:"label"`
	Url       *string    `json:"url" db:"url"`
	Icon      *string    `json:"icon" db:"icon"`
	ParentID  *uint      `json:"parent_id" db:"parent_id"`
	SortOrder uint16     `json:"sort_order" db:"sort_order"`
	CreatedAt *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy *uint64    `json:"deleted_by" db:"deleted_by"`
}

// MenuNavigation represents the menu_navigation view
type MenuNavigation struct {
	ID       uint   `json:"id" db:"id"`
	Label    string `json:"label" db:"label"`
	URL      string `json:"url" db:"url"`
	Icon     string `json:"icon" db:"icon"`
	Children string `json:"children" db:"children"`
}
