package repositories

import (
	"database/sql"
	"fmt"

	"adminbe/internal/app/models"
)

// UserMenuRepository interface defines data access methods for user menus
type UserMenuRepository interface {
	GetAll() ([]models.UserMenu, error)
	GetByUserAndMenu(userID uint64, menuID uint) (*models.UserMenu, error)
	Create(req models.UserMenu) error
	Delete(userID uint64, menuID uint, deletedBy *uint64) error
}

// userMenuRepository implements UserMenuRepository
type userMenuRepository struct {
	db *sql.DB
}

// NewUserMenuRepository creates a new user menu repository
func NewUserMenuRepository(db *sql.DB) UserMenuRepository {
	return &userMenuRepository{db: db}
}

// GetAll retrieves all active user-menu assignments
func (r *userMenuRepository) GetAll() ([]models.UserMenu, error) {
	rows, err := r.db.Query(`
		SELECT user_id, menu_id, deleted_at, deleted_by
		FROM user_menu
		WHERE deleted_at IS NULL`)
	if err != nil {
		return nil, fmt.Errorf("failed to query user menus: %w", err)
	}
	defer rows.Close()

	var userMenus []models.UserMenu
	for rows.Next() {
		var um models.UserMenu
		if err := rows.Scan(&um.UserID, &um.MenuID, &um.DeletedAt, &um.DeletedBy); err != nil {
			return nil, fmt.Errorf("failed to scan user menu: %w", err)
		}
		userMenus = append(userMenus, um)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user menus: %w", err)
	}

	return userMenus, nil
}

// GetByUserAndMenu retrieves a user-menu assignment by user and menu IDs
func (r *userMenuRepository) GetByUserAndMenu(userID uint64, menuID uint) (*models.UserMenu, error) {
	var um models.UserMenu
	row := r.db.QueryRow(`
		SELECT user_id, menu_id, deleted_at, deleted_by
		FROM user_menu
		WHERE user_id = ? AND menu_id = ? AND deleted_at IS NULL`,
		userID, menuID)

	err := row.Scan(&um.UserID, &um.MenuID, &um.DeletedAt, &um.DeletedBy)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan user menu: %w", err)
	}

	return &um, nil
}

// Create inserts a new user-menu assignment
func (r *userMenuRepository) Create(req models.UserMenu) error {
	_, err := r.db.Exec(`
		INSERT INTO user_menu (user_id, menu_id, deleted_at, deleted_by)
		VALUES (?, ?, ?, ?)`,
		req.UserID, req.MenuID, req.DeletedAt, req.DeletedBy)
	return err
}

// Delete performs a soft delete
func (r *userMenuRepository) Delete(userID uint64, menuID uint, deletedBy *uint64) error {
	_, err := r.db.Exec(`
		UPDATE user_menu SET deleted_at = NOW(), deleted_by = ?
		WHERE user_id = ? AND menu_id = ? AND deleted_at IS NULL`,
		deletedBy, userID, menuID)
	return err
}
