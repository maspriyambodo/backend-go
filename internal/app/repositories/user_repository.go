package repositories

import (
	"database/sql"
	"fmt"
	"strings"

	"adminbe/internal/app/models"
)

// UserRepository interface defines data access methods for users
type UserRepository interface {
	GetAll(limit, offset int) ([]models.User, error)
	GetByID(id uint64) (*models.User, error)
	Create(req models.CreateUserRequest, hashedPassword string) (uint64, error)
	Update(id uint64, req models.UpdateUserRequest, hashedPassword string) error
	Delete(id uint64) error
	CountActive() (int, error)
}

// userRepository implements UserRepository
type userRepository struct {
	db *sql.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

// GetAll retrieves all active users with pagination
func (r *userRepository) GetAll(limit, offset int) ([]models.User, error) {
	rows, err := r.db.Query(`
		SELECT id, username, email, status, created_at, updated_at, deleted_at, deleted_by
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?`,
		limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var u models.User
		if err := rows.Scan(&u.ID, &u.Username, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.DeletedBy); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		u.PasswordHash = "" // Remove sensitive data
		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating users: %w", err)
	}

	return users, nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(id uint64) (*models.User, error) {
	var u models.User
	row := r.db.QueryRow(`
		SELECT id, username, email, status, created_at, updated_at, deleted_at, deleted_by
		FROM users
		WHERE id = ? AND deleted_at IS NULL`,
		id)

	err := row.Scan(&u.ID, &u.Username, &u.Email, &u.Status, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt, &u.DeletedBy)
	if err == sql.ErrNoRows {
		return nil, err
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan user: %w", err)
	}

	return &u, nil
}

// Create inserts a new user
func (r *userRepository) Create(req models.CreateUserRequest, hashedPassword string) (uint64, error) {
	status := uint8(1) // default active
	if req.Status != nil {
		status = *req.Status
	}

	result, err := r.db.Exec(`
		INSERT INTO users (username, email, password_hash, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())`,
		req.Username, req.Email, hashedPassword, status)
	if err != nil {
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get last insert id: %w", err)
	}

	return uint64(userID), nil
}

// Update modifies an existing user
func (r *userRepository) Update(id uint64, req models.UpdateUserRequest, hashedPassword string) error {
	var setParts []string
	var args []interface{}

	if req.Username != "" {
		setParts = append(setParts, "username = ?")
		args = append(args, req.Username)
	}
	if req.Email != "" {
		setParts = append(setParts, "email = ?")
		args = append(args, req.Email)
	}
	if req.Password != "" {
		setParts = append(setParts, "password_hash = ?")
		args = append(args, hashedPassword)
	}
	if req.Status != nil {
		setParts = append(setParts, "status = ?")
		args = append(args, *req.Status)
	}

	if len(setParts) == 0 {
		return fmt.Errorf("no fields to update")
	}

	setClause := strings.Join(setParts, ", ")
	query := fmt.Sprintf("UPDATE users SET %s, updated_at = NOW() WHERE id = ? AND deleted_at IS NULL", setClause)
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}

// Delete performs a soft delete
func (r *userRepository) Delete(id uint64) error {
	_, err := r.db.Exec(`
		UPDATE users SET deleted_at = NOW(), updated_at = NOW()
		WHERE id = ? AND deleted_at IS NULL`,
		id)
	return err
}

// CountActive counts active users
func (r *userRepository) CountActive() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM users WHERE deleted_at IS NULL").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}
	return count, nil
}
