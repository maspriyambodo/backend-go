package models

import (
	"time"
)

// User represents the users table
type User struct {
	ID           uint64     `json:"id" db:"id"`
	Username     string     `json:"username" db:"username"`
	Email        string     `json:"email" db:"email"`
	PasswordHash string     `json:"-" db:"password_hash"`
	Status       uint8      `json:"status" db:"status"`
	CreatedAt    *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy    *uint64    `json:"deleted_by" db:"deleted_by"`
}

// AuditLog represents the audit_logs table
type AuditLog struct {
	ID        uint64      `json:"id" db:"id"`
	UserID    uint64      `json:"user_id" db:"user_id"`
	EventType string      `json:"event_type" db:"event_type"`
	TableName string      `json:"table_name" db:"table_name"`
	RecordID  uint64      `json:"record_id" db:"record_id"`
	OldValues interface{} `json:"old_values" db:"old_values"`
	NewValues interface{} `json:"new_values" db:"new_values"`
	IPAddress []byte      `json:"ip_address" db:"ip_address"`
	UserAgent *string     `json:"user_agent" db:"user_agent"`
	CreatedAt *time.Time  `json:"created_at" db:"created_at"`
}

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

// CreateUserRequest for creating a new user
type CreateUserRequest struct {
	Username string `json:"username" binding:"required,min=3,max=100"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Status   *uint8 `json:"status,omitempty"`
}

// UpdateUserRequest for updating an existing user
type UpdateUserRequest struct {
	Username string `json:"username,omitempty" binding:"min=3,max=100"`
	Email    string `json:"email,omitempty" binding:"email"`
	Password string `json:"password,omitempty" binding:"min=6"`
	Status   *uint8 `json:"status,omitempty"`
}

// Role represents the roles table
type Role struct {
	ID          uint       `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Description *string    `json:"description" db:"description"`
	CreatedAt   *time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at" db:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy   *uint64    `json:"deleted_by" db:"deleted_by"`
}

// CreateRoleRequest for creating a new role
type CreateRoleRequest struct {
	Name        string  `json:"name" binding:"required,min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

// UpdateRoleRequest for updating an existing role
type UpdateRoleRequest struct {
	Name        *string `json:"name,omitempty" binding:"min=1,max=100"`
	Description *string `json:"description,omitempty"`
}

// RoleInheritance represents the role_inheritances table
type RoleInheritance struct {
	ID           uint64     `json:"id" db:"id"`
	RoleID       uint       `json:"role_id" db:"role_id"`
	ParentRoleID uint       `json:"parent_role_id" db:"parent_role_id"`
	CreatedAt    *time.Time `json:"created_at" db:"created_at"`
}

// CreateRoleInheritanceRequest for creating a new role inheritance
type CreateRoleInheritanceRequest struct {
	RoleID       uint `json:"role_id" binding:"required"`
	ParentRoleID uint `json:"parent_role_id" binding:"required"`
}

// UpdateRoleInheritanceRequest for updating an existing role inheritance
type UpdateRoleInheritanceRequest struct {
	RoleID       *uint `json:"role_id,omitempty"`
	ParentRoleID *uint `json:"parent_role_id,omitempty"`
}

// VRole represents the v_roles view
type VRole struct {
	RoleID    uint   `json:"role_id" db:"role_id"`
	RoleName  string `json:"role_name" db:"role_name"`
	ChildID   uint   `json:"child_id" db:"child_id"`
	ChildName string `json:"child_name" db:"child_name"`
	Level     uint   `json:"level" db:"level"`
}

// RoleMenu represents the role_menu table
type RoleMenu struct {
	RoleID    uint       `json:"role_id" db:"role_id"`
	MenuID    uint       `json:"menu_id" db:"menu_id"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy *uint64    `json:"deleted_by" db:"deleted_by"`
}

// CreateRoleMenuRequest for creating a new role-menu assignment
type CreateRoleMenuRequest struct {
	RoleID uint `json:"role_id" binding:"required"`
	MenuID uint `json:"menu_id" binding:"required"`
}

// UpdateRoleMenuRequest for updating an existing role-menu assignment
type UpdateRoleMenuRequest struct {
	RoleID *uint `json:"role_id,omitempty"`
	MenuID *uint `json:"menu_id,omitempty"`
}

// MenuNavigation represents the menu_navigation view
type MenuNavigation struct {
	ID       uint   `json:"id" db:"id"`
	Label    string `json:"label" db:"label"`
	URL      string `json:"url" db:"url"`
	Icon     string `json:"icon" db:"icon"`
	Children string `json:"children" db:"children"`
}

// UserMenu represents the user_menu table
type UserMenu struct {
	UserID    uint64     `json:"user_id" db:"user_id"`
	MenuID    uint       `json:"menu_id" db:"menu_id"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy *uint64    `json:"deleted_by" db:"deleted_by"`
}

// CreateUserMenuRequest for creating a new user-menu assignment
type CreateUserMenuRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	MenuID uint   `json:"menu_id" binding:"required"`
}

// UpdateUserMenuRequest for updating an existing user-menu assignment
type UpdateUserMenuRequest struct {
	UserID *uint64 `json:"user_id,omitempty"`
	MenuID *uint   `json:"menu_id,omitempty"`
}

// UserRole represents the user_roles table
type UserRole struct {
	UserID    uint64     `json:"user_id" db:"user_id"`
	RoleID    uint       `json:"role_id" db:"role_id"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
	DeletedBy *uint64    `json:"deleted_by" db:"deleted_by"`
}

// CreateUserRoleRequest for creating a new user-role assignment
type CreateUserRoleRequest struct {
	UserID uint64 `json:"user_id" binding:"required"`
	RoleID uint   `json:"role_id" binding:"required"`
}

// UpdateUserRoleRequest for updating an existing user-role assignment
type UpdateUserRoleRequest struct {
	UserID *uint64 `json:"user_id,omitempty"`
	RoleID *uint   `json:"role_id,omitempty"`
}

// JasperServerConfig holds JasperServer configuration
type JasperServerConfig struct {
	BaseURL      string `yaml:"base_url" json:"base_url"`
	Username     string `yaml:"username" json:"username"`
	Password     string `yaml:"password" json:"password"`
	Organization string `yaml:"organization" json:"organization"`
}

// JasperReportRequest represents a request to run a report
type JasperReportRequest struct {
	ReportPath   string                 `json:"report_path" binding:"required"`
	OutputFormat string                 `json:"output_format" binding:"oneof=pdf html excel pptx rtf docx xlsx xls png"`
	Parameters   map[string]interface{} `json:"parameters,omitempty"`
	Interactive  bool                   `json:"interactive,omitempty"`
	Page         uint                   `json:"page,omitempty"`
	Pages        string                 `json:"pages,omitempty"`
}

// JasperReportResponse represents the response from running a report
type JasperReportResponse struct {
	ID           string `json:"id"`
	Output       string `json:"output,omitempty"`
	Status       string `json:"status"`
	ReportURI    string `json:"reportURI,omitempty"`
	ErrorMessage string `json:"errorMessage,omitempty"`
	Permissions  string `json:"permissions,omitempty"`
}

// AppProvince represents the app_province table
type AppProvince struct {
	ProvinceID    uint64 `json:"province_id" db:"province_id"`
	ProvinceTitle string `json:"province_title" db:"province_title"`
	ProvinceIDNew uint64 `json:"province_id_new" db:"province_id_new"`
}

// AppCity represents the app_city table
type AppCity struct {
	CityID       uint64 `json:"city_id" db:"city_id"`
	CityTitle    string `json:"city_title" db:"city_title"`
	CityProvince uint64 `json:"city_province" db:"city_province"`
	CityIDNew    uint64 `json:"city_id_new" db:"city_id_new"`
}

// DataLintangKotaCmsNew represents the data_lintang_kota_cms_new table
type DataLintangKotaCmsNew struct {
	IDKota        uint64     `json:"id_kota" db:"id_kota"`
	NamaPropinsi  uint64     `json:"nama_propinsi" db:"nama_propinsi"`
	NamaKota      uint64     `json:"nama_kota" db:"nama_kota"`
	BujurTempat   string     `json:"bujur_tempat" db:"bujur_tempat"`
	LintangTempat string     `json:"lintang_tempat" db:"lintang_tempat"`
	TimeZone      string     `json:"time_zone" db:"time_zone"`
	H             *int       `json:"h" db:"h"`
	TimeCreate    *time.Time `json:"time_create" db:"time_create"`
	// Joined fields for query results
	ProvinceTitle string `json:"province_title"`
	CityTitle     string `json:"city_title"`
}

// HisabTglPuasa represents the hisab_tgl_puasa table
type HisabTglPuasa struct {
	TglID      uint64     `json:"tgl_id" db:"tgl_id"`
	TglTahun   *int       `json:"tgl_tahun" db:"tgl_tahun"`
	TglStart   *string    `json:"tgl_start" db:"tgl_start"`
	TglEnd     *string    `json:"tgl_end" db:"tgl_end"`
	TglStatus  *int       `json:"tgl_status" db:"tgl_status"`
	TglHijriah *int       `json:"tgl_hijriah" db:"tgl_hijriah"`
	TimeAdd    *time.Time `json:"time_add" db:"time_add"`
	TimeUpdate *time.Time `json:"time_update" db:"time_update"`
	UserAdd    *uint64    `json:"user_add" db:"user_add"`
	UserUpdate *uint64    `json:"user_update" db:"user_update"`
}
