package services

import (
	"database/sql"
	"fmt"
	"strconv"

	"adminbe/internal/app/models"
	"adminbe/internal/app/repositories"

	"github.com/gin-gonic/gin"
)

// UserRoleService interface defines business logic for user roles
type UserRoleService interface {
	ListUserRoles() ([]models.UserRole, error)
	GetUserRole(userIDStr, roleIDStr string) (*models.UserRole, error)
	CreateUserRole(req models.CreateUserRoleRequest) (*models.UserRole, error)
	DeleteUserRole(userIDStr, roleIDStr string) error
}

// userRoleService implements UserRoleService
type userRoleService struct {
	repo repositories.UserRoleRepository
}

// NewUserRoleService creates a new user role service
func NewUserRoleService(repo repositories.UserRoleRepository) UserRoleService {
	return &userRoleService{repo: repo}
}

// ListUserRoles handles listing all user-role assignments
func (s *userRoleService) ListUserRoles() ([]models.UserRole, error) {
	userRoles, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return userRoles, nil
}

// GetUserRole handles getting a user-role assignment by user and role IDs
func (s *userRoleService) GetUserRole(userIDStr, roleIDStr string) (*models.UserRole, error) {
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	roleID, err := parseUint(roleIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	userRole, err := s.repo.GetByUserAndRole(userID, roleID)
	if err == sql.ErrNoRows {
		return nil, gin.Error{
			Err:  fmt.Errorf("user-role assignment not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %w", err)
	}

	return userRole, nil
}

// CreateUserRole handles creating a new user-role assignment
func (s *userRoleService) CreateUserRole(req models.CreateUserRoleRequest) (*models.UserRole, error) {
	// Check if assignment already exists
	existing, err := s.repo.GetByUserAndRole(req.UserID, req.RoleID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check user role existence: %w", err)
	}
	if existing != nil {
		return nil, gin.Error{
			Err:  fmt.Errorf("user-role assignment already exists"),
			Type: gin.ErrorTypePublic,
		}
	}

	userRole := models.UserRole{
		UserID: req.UserID,
		RoleID: req.RoleID,
		// deleted_at and deleted_by are nil for active
	}

	err = s.repo.Create(userRole)
	if err != nil {
		return nil, fmt.Errorf("failed to create user role: %w", err)
	}

	// Return the created assignment
	createdUserRole, err := s.repo.GetByUserAndRole(req.UserID, req.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created user role: %w", err)
	}

	return createdUserRole, nil
}

// DeleteUserRole handles deleting a user-role assignment
func (s *userRoleService) DeleteUserRole(userIDStr, roleIDStr string) error {
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	roleID, err := parseUint(roleIDStr)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	// Check if assignment exists
	_, err = s.repo.GetByUserAndRole(userID, roleID)
	if err == sql.ErrNoRows {
		return gin.Error{
			Err:  fmt.Errorf("user-role assignment not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return fmt.Errorf("failed to check user role existence: %w", err)
	}

	return s.repo.Delete(userID, roleID, nil) // TODO: get current user ID for audit
}
