package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/app/repositories"

	"github.com/gin-gonic/gin"
)

// RoleService interface defines business logic for roles
type RoleService interface {
	ListRoles() ([]models.Role, error)
	GetRole(id string) (*models.Role, error)
	CreateRole(req models.CreateRoleRequest) (*models.Role, error)
	UpdateRole(id string, req models.UpdateRoleRequest) (*models.Role, error)
	DeleteRole(id string) error
}

// roleService implements RoleService
type roleService struct {
	repo repositories.RoleRepository
}

// NewRoleService creates a new role service
func NewRoleService(repo repositories.RoleRepository) RoleService {
	return &roleService{repo: repo}
}

// ListRoles handles listing all roles
func (s *roleService) ListRoles() ([]models.Role, error) {
	roles, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %w", err)
	}

	return roles, nil
}

// GetRole handles getting a role by ID
func (s *roleService) GetRole(id string) (*models.Role, error) {
	roleID, err := parseUint(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	role, err := s.repo.GetByID(roleID)
	if err == sql.ErrNoRows {
		return nil, gin.Error{
			Err:  fmt.Errorf("role not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}

	return role, nil
}

// CreateRole handles creating a new role
func (s *roleService) CreateRole(req models.CreateRoleRequest) (*models.Role, error) {
	if err := s.validateRoleNameUniqueness(req.Name, 0); err != nil {
		return nil, err
	}

	now := time.Now()
	role := models.Role{
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   &now,
		UpdatedAt:   &now,
	}

	roleID, err := s.repo.Create(role)
	if err != nil {
		// Check for duplicate key error
		if strings.Contains(err.Error(), "1062") {
			return nil, gin.Error{
				Err:  fmt.Errorf("role name already exists"),
				Type: gin.ErrorTypePublic,
			}
		}
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	// Return the created role
	return s.retrieveRoleByID(roleID)
}

// UpdateRole handles updating an existing role
func (s *roleService) UpdateRole(id string, req models.UpdateRoleRequest) (*models.Role, error) {
	roleID, err := parseUint(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	if err := s.ensureRoleExists(roleID); err != nil {
		return nil, err
	}

	// Check name uniqueness if name is being updated
	if req.Name != nil {
		if err := s.validateRoleNameUniqueness(*req.Name, roleID); err != nil {
			return nil, err
		}
	}

	updateData := make(map[string]interface{})
	if req.Name != nil {
		updateData["name"] = *req.Name
	}
	if req.Description != nil {
		updateData["description"] = req.Description
	}

	if err := s.repo.Update(roleID, updateData); err != nil {
		// Check for duplicate key error
		if strings.Contains(err.Error(), "1062") {
			return nil, gin.Error{
				Err:  fmt.Errorf("role name already exists"),
				Type: gin.ErrorTypePublic,
			}
		}
		return nil, fmt.Errorf("failed to update role: %w", err)
	}

	// Return updated role
	return s.retrieveRoleByID(roleID)
}

// DeleteRole handles deleting a role
func (s *roleService) DeleteRole(id string) error {
	roleID, err := parseUint(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	if err := s.ensureRoleExists(roleID); err != nil {
		return err
	}

	return s.repo.Delete(roleID, nil) // TODO: get current user ID for audit
}

// validateRoleNameUniqueness checks if a role name is unique, excluding a specific ID
func (s *roleService) validateRoleNameUniqueness(name string, excludeID uint) error {
	existing, err := s.repo.GetByName(name)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check role name uniqueness: %w", err)
	}
	if existing != nil && existing.ID != excludeID {
		return gin.Error{
			Err:  fmt.Errorf("role name already exists"),
			Type: gin.ErrorTypePublic,
		}
	}
	return nil
}

// ensureRoleExists checks if a role exists by ID
func (s *roleService) ensureRoleExists(roleID uint) error {
	_, err := s.repo.GetByID(roleID)
	if err == sql.ErrNoRows {
		return gin.Error{
			Err:  fmt.Errorf("role not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return fmt.Errorf("failed to check role existence: %w", err)
	}
	return nil
}

// retrieveRoleByID fetches a role by ID with error formatting
func (s *roleService) retrieveRoleByID(roleID uint) (*models.Role, error) {
	role, err := s.repo.GetByID(roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve role: %w", err)
	}
	return role, nil
}
