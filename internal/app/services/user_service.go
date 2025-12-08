package services

import (
	"database/sql"
	"fmt"
	"strconv"

	"adminbe/internal/app/models"
	"adminbe/internal/app/repositories"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UserService interface defines business logic for users
type UserService interface {
	ListUsers(page, limit int) (map[string]interface{}, error)
	GetUser(id string) (*models.User, error)
	CreateUser(req models.CreateUserRequest) (*models.User, error)
	UpdateUser(id string, req models.UpdateUserRequest) (*models.User, error)
	DeleteUser(id string) error
}

// userService implements UserService
type userService struct {
	repo repositories.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo repositories.UserRepository) UserService {
	return &userService{repo: repo}
}

// ListUsers handles listing users with pagination
func (s *userService) ListUsers(page, limit int) (map[string]interface{}, error) {
	offset := (page - 1) * limit

	users, err := s.repo.GetAll(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}

	total, err := s.repo.CountActive()
	if err != nil {
		return nil, fmt.Errorf("failed to count users: %w", err)
	}

	totalPages := (total + limit - 1) / limit

	return map[string]interface{}{
		"data": users,
		"pagination": map[string]interface{}{
			"page":        page,
			"limit":       limit,
			"total":       total,
			"total_pages": totalPages,
			"has_next":    page < totalPages,
			"has_prev":    page > 1,
		},
	}, nil
}

// GetUser handles getting a user by ID
func (s *userService) GetUser(id string) (*models.User, error) {
	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	user, err := s.repo.GetByID(userID)
	if err == sql.ErrNoRows {
		return nil, gin.Error{
			Err:  fmt.Errorf("user not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// CreateUser handles creating a new user
func (s *userService) CreateUser(req models.CreateUserRequest) (*models.User, error) {
	// Hash password
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("password hash failed: %w", err)
	}

	userID, err := s.repo.Create(req, string(hashed))
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Return the created user (without password)
	user, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created user: %w", err)
	}

	return user, nil
}

// UpdateUser handles updating an existing user
func (s *userService) UpdateUser(id string, req models.UpdateUserRequest) (*models.User, error) {
	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	// Check if user exists
	_, err = s.repo.GetByID(userID)
	if err == sql.ErrNoRows {
		return nil, gin.Error{
			Err:  fmt.Errorf("user not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	// Hash password if provided
	var hashedPassword string
	if req.Password != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, fmt.Errorf("password hash failed: %w", err)
		}
		hashedPassword = string(hashed)
	}

	err = s.repo.Update(userID, req, hashedPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Return updated user
	updatedUser, err := s.repo.GetByID(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated user: %w", err)
	}

	return updatedUser, nil
}

// DeleteUser handles soft deleting a user
func (s *userService) DeleteUser(id string) error {
	userID, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	// Check if user exists
	_, err = s.repo.GetByID(userID)
	if err == sql.ErrNoRows {
		return gin.Error{
			Err:  fmt.Errorf("user not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return fmt.Errorf("failed to check user existence: %w", err)
	}

	return s.repo.Delete(userID)
}
