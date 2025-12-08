package repositories

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"adminbe/internal/app/models"
)

// RoleRepository interface defines data access methods for roles
type RoleRepository interface {
	GetAll() ([]models.Role, error)
	GetByID(id uint) (*models.Role, error)
	GetByName(name string) (*models.Role, error)
	Create(req models.Role) (uint, error)
	Update(id uint, req map[string]interface{}) error
	Delete(id uint, deletedBy *uint64) error
}

// roleRepository implements RoleRepository
type roleRepository struct {
	db *sql.DB
}

// NewRoleRepository creates a new role repository
func NewRoleRepository(db *sql.DB) RoleRepository {
	return &roleRepository{db: db}
}

// GetAll retrieves all active roles
func (r *roleRepository) GetAll() ([]models.Role, error) {
	rows, err := r.db.Query(`
		SELECT id, name, description, created_at, updated_at, deleted_at, deleted_by
		FROM roles
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query roles: %w", err)
	}
	defer rows.Close()

	var roles []models.Role
	for rows.Next() {
		var role models.Role
		if err := rows.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt, &role.DeletedBy); err != nil {
			return nil, fmt.Errorf("failed to scan role: %w", err)
		}
		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating roles: %w", err)
	}

	return roles, nil
}

// GetByID retrieves a role by ID
func (r *roleRepository) GetByID(id uint) (*models.Role, error) {
	var role models.Role
	row := r.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at, deleted_at, deleted_by
		FROM roles
		WHERE id = ? AND deleted_at IS NULL`,
		id)

	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt, &role.DeletedBy)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan role: %w", err)
	}

	return &role, nil
}

// GetByName retrieves a role by name
func (r *roleRepository) GetByName(name string) (*models.Role, error) {
	var role models.Role
	row := r.db.QueryRow(`
		SELECT id, name, description, created_at, updated_at, deleted_at, deleted_by
		FROM roles
		WHERE name = ? AND deleted_at IS NULL`,
		name)

	err := row.Scan(&role.ID, &role.Name, &role.Description, &role.CreatedAt, &role.UpdatedAt, &role.DeletedAt, &role.DeletedBy)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan role: %w", err)
	}

	return &role, nil
}

// Create inserts a new role
func (r *roleRepository) Create(req models.Role) (uint, error) {
	result, err := r.db.Exec(`
		INSERT INTO roles (name, description, created_at, updated_at)
		VALUES (?, ?, ?, ?)`,
		req.Name, req.Description, req.CreatedAt, req.UpdatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to insert role: %w", err)
	}

	roleID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint(roleID), nil
}

// Update modifies an existing role with dynamic fields
func (r *roleRepository) Update(id uint, req map[string]interface{}) error {
	var setParts []string
	var args []interface{}

	if name, ok := req["name"].(string); ok && name != "" {
		setParts = append(setParts, "name = ?")
		args = append(args, name)
	}
	if description, ok := req["description"]; ok {
		setParts = append(setParts, "description = ?")
		args = append(args, description)
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setParts = append(setParts, "updated_at = ?")
	args = append(args, time.Now())

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf("UPDATE roles SET %s WHERE id = ? AND deleted_at IS NULL", setClause)
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}

// Delete performs a soft delete
func (r *roleRepository) Delete(id uint, deletedBy *uint64) error {
	_, err := r.db.Exec(`
		UPDATE roles SET deleted_at = ?, updated_at = ?, deleted_by = ?
		WHERE id = ? AND deleted_at IS NULL`,
		time.Now(), time.Now(), deletedBy, id)
	return err
}
