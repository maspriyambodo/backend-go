package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"adminbe/internal/app/models"
)

// RoleInheritanceRepository interface defines data access methods for role inheritances
type RoleInheritanceRepository interface {
	GetAll() ([]models.RoleInheritance, error)
	GetByID(id uint64) (*models.RoleInheritance, error)
	Create(req models.RoleInheritance) (uint64, error)
	Update(id uint64, req map[string]interface{}) error
	Delete(id uint64) error
}

// roleInheritanceRepository implements RoleInheritanceRepository
type roleInheritanceRepository struct {
	db *sql.DB
}

// NewRoleInheritanceRepository creates a new role inheritance repository
func NewRoleInheritanceRepository(db *sql.DB) RoleInheritanceRepository {
	return &roleInheritanceRepository{db: db}
}

// GetAll retrieves all role inheritances
func (r *roleInheritanceRepository) GetAll() ([]models.RoleInheritance, error) {
	rows, err := r.db.Query(`
		SELECT id, role_id, parent_role_id, created_at
		FROM role_inheritances
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("failed to query role inheritances: %w", err)
	}
	defer rows.Close()

	var inheritances []models.RoleInheritance
	for rows.Next() {
		var ri models.RoleInheritance
		if err := rows.Scan(&ri.ID, &ri.RoleID, &ri.ParentRoleID, &ri.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan role inheritance: %w", err)
		}
		inheritances = append(inheritances, ri)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating role inheritances: %w", err)
	}

	return inheritances, nil
}

// GetByID retrieves a role inheritance by ID
func (r *roleInheritanceRepository) GetByID(id uint64) (*models.RoleInheritance, error) {
	var ri models.RoleInheritance
	row := r.db.QueryRow(`
		SELECT id, role_id, parent_role_id, created_at
		FROM role_inheritances
		WHERE id = ?`,
		id)

	err := row.Scan(&ri.ID, &ri.RoleID, &ri.ParentRoleID, &ri.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan role inheritance: %w", err)
	}

	return &ri, nil
}

// Create inserts a new role inheritance
func (r *roleInheritanceRepository) Create(req models.RoleInheritance) (uint64, error) {
	result, err := r.db.Exec(`
		INSERT INTO role_inheritances (role_id, parent_role_id, created_at)
		VALUES (?, ?, ?)`,
		req.RoleID, req.ParentRoleID, req.CreatedAt)
	if err != nil {
		return 0, fmt.Errorf("failed to insert role inheritance: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint64(id), nil
}

// Update modifies an existing role inheritance with dynamic fields
func (r *roleInheritanceRepository) Update(id uint64, req map[string]interface{}) error {
	var setParts []string
	var args []interface{}

	if roleID, ok := req["role_id"].(uint); ok && roleID > 0 {
		setParts = append(setParts, "role_id = ?")
		args = append(args, roleID)
	}
	if parentRoleID, ok := req["parent_role_id"].(uint); ok && parentRoleID > 0 {
		setParts = append(setParts, "parent_role_id = ?")
		args = append(args, parentRoleID)
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf("UPDATE role_inheritances SET %s WHERE id = ?", setClause)
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}

// Delete removes a role inheritance (hard delete)
func (r *roleInheritanceRepository) Delete(id uint64) error {
	_, err := r.db.Exec(`DELETE FROM role_inheritances WHERE id = ?`, id)
	return err
}
