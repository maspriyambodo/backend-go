package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"adminbe/internal/app/models"
)

// MenuRepository interface defines data access methods for menus
type MenuRepository interface {
	GetAll() ([]models.Menu, error)
	GetByID(id uint) (*models.Menu, error)
	Create(req models.Menu) (uint, error)
	Update(id uint, req map[string]interface{}) error
	Delete(id uint, deletedBy *uint64) error
}

// menuRepository implements MenuRepository
type menuRepository struct {
	db *sql.DB
}

// NewMenuRepository creates a new menu repository
func NewMenuRepository(db *sql.DB) MenuRepository {
	return &menuRepository{db: db}
}

// GetAll retrieves all active menus
func (r *menuRepository) GetAll() ([]models.Menu, error) {
	rows, err := r.db.Query(`
		SELECT id, label, url, icon, parent_id, sort_order, created_at, updated_at, deleted_at, deleted_by
		FROM menu
		WHERE deleted_at IS NULL
		ORDER BY sort_order`)
	if err != nil {
		return nil, fmt.Errorf("failed to query menus: %w", err)
	}
	defer rows.Close()

	var menus []models.Menu
	for rows.Next() {
		var m models.Menu
		if err := rows.Scan(&m.ID, &m.Label, &m.Url, &m.Icon, &m.ParentID, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.DeletedBy); err != nil {
			return nil, fmt.Errorf("failed to scan menu: %w", err)
		}
		menus = append(menus, m)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating menus: %w", err)
	}

	return menus, nil
}

// GetByID retrieves a menu by ID
func (r *menuRepository) GetByID(id uint) (*models.Menu, error) {
	var m models.Menu
	row := r.db.QueryRow(`
		SELECT id, label, url, icon, parent_id, sort_order, created_at, updated_at, deleted_at, deleted_by
		FROM menu
		WHERE id = ? AND deleted_at IS NULL`,
		id)

	err := row.Scan(&m.ID, &m.Label, &m.Url, &m.Icon, &m.ParentID, &m.SortOrder, &m.CreatedAt, &m.UpdatedAt, &m.DeletedAt, &m.DeletedBy)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan menu: %w", err)
	}

	return &m, nil
}

// Create inserts a new menu
func (r *menuRepository) Create(req models.Menu) (uint, error) {
	result, err := r.db.Exec(`
		INSERT INTO menu (label, url, icon, parent_id, sort_order, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		req.Label, req.Url, req.Icon, req.ParentID, req.SortOrder, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to insert menu: %w", err)
	}

	menuID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint(menuID), nil
}

// Update modifies an existing menu with dynamic fields
func (r *menuRepository) Update(id uint, req map[string]interface{}) error {
	var setParts []string
	var args []interface{}

	if label, ok := req["label"].(string); ok && label != "" {
		setParts = append(setParts, "label = ?")
		args = append(args, label)
	}
	if url, ok := req["url"]; ok {
		setParts = append(setParts, "url = ?")
		args = append(args, url)
	}
	if icon, ok := req["icon"]; ok {
		setParts = append(setParts, "icon = ?")
		args = append(args, icon)
	}
	if parentID, ok := req["parent_id"]; ok {
		setParts = append(setParts, "parent_id = ?")
		args = append(args, parentID)
	}
	if sortOrder, ok := req["sort_order"].(uint16); ok {
		setParts = append(setParts, "sort_order = ?")
		args = append(args, sortOrder)
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now())

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf("UPDATE menu SET %s WHERE id = ? AND deleted_at IS NULL", setClause)
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}

// Delete performs a soft delete
func (r *menuRepository) Delete(id uint, deletedBy *uint64) error {
	_, err := r.db.Exec(`
		UPDATE menu SET deleted_at = ?, updated_at = ?, deleted_by = ?
		WHERE id = ? AND deleted_at IS NULL`,
		time.Now(), time.Now(), deletedBy, id)
	return err
}
