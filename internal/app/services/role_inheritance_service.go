package services

import (
	"database/sql"
	"fmt"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/app/repositories"

	"github.com/gin-gonic/gin"
)

// RoleInheritanceService interface defines business logic for role inheritances
type RoleInheritanceService interface {
	ListRoleInheritances() ([]models.RoleInheritance, error)
	GetRoleInheritance(id string) (*models.RoleInheritance, error)
	CreateRoleInheritance(req models.CreateRoleInheritanceRequest) (*models.RoleInheritance, error)
	UpdateRoleInheritance(id string, req models.UpdateRoleInheritanceRequest) (*models.RoleInheritance, error)
	DeleteRoleInheritance(id string) error
}

// roleInheritanceService implements RoleInheritanceService
type roleInheritanceService struct {
	repo repositories.RoleInheritanceRepository
}

// NewRoleInheritanceService creates a new role inheritance service
func NewRoleInheritanceService(repo repositories.RoleInheritanceRepository) RoleInheritanceService {
	return &roleInheritanceService{repo: repo}
}

// ListRoleInheritances handles listing all role inheritances
func (s *roleInheritanceService) ListRoleInheritances() ([]models.RoleInheritance, error) {
	inheritances, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get role inheritances: %w", err)
	}

	return inheritances, nil
}

// GetRoleInheritance handles getting a role inheritance by ID
func (s *roleInheritanceService) GetRoleInheritance(id string) (*models.RoleInheritance, error) {
	inheritanceID, err := parseUint64(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	inheritance, err := s.repo.GetByID(inheritanceID)
	if err == sql.ErrNoRows {
		return nil, gin.Error{
			Err:  fmt.Errorf("role inheritance not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role inheritance: %w", err)
	}

	return inheritance, nil
}

// CreateRoleInheritance handles creating a new role inheritance
func (s *roleInheritanceService) CreateRoleInheritance(req models.CreateRoleInheritanceRequest) (*models.RoleInheritance, error) {
	now := time.Now()
	inheritance := models.RoleInheritance{
		RoleID:       req.RoleID,
		ParentRoleID: req.ParentRoleID,
		CreatedAt:    &now,
	}

	id, err := s.repo.Create(inheritance)
	if err != nil {
		return nil, fmt.Errorf("failed to create role inheritance: %w", err)
	}

	// Return the created inheritance
	createdInheritance, err := s.repo.GetByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created role inheritance: %w", err)
	}

	return createdInheritance, nil
}

// UpdateRoleInheritance handles updating an existing role inheritance
func (s *roleInheritanceService) UpdateRoleInheritance(id string, req models.UpdateRoleInheritanceRequest) (*models.RoleInheritance, error) {
	inheritanceID, err := parseUint64(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	// Check if inheritance exists
	_, err = s.repo.GetByID(inheritanceID)
	if err == sql.ErrNoRows {
		return nil, gin.Error{
			Err:  fmt.Errorf("role inheritance not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to check role inheritance existence: %w", err)
	}

	updateData := make(map[string]interface{})
	if req.RoleID != nil {
		updateData["role_id"] = *req.RoleID
	}
	if req.ParentRoleID != nil {
		updateData["parent_role_id"] = *req.ParentRoleID
	}

	err = s.repo.Update(inheritanceID, updateData)
	if err != nil {
		return nil, fmt.Errorf("failed to update role inheritance: %w", err)
	}

	// Return updated inheritance
	updatedInheritance, err := s.repo.GetByID(inheritanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated role inheritance: %w", err)
	}

	return updatedInheritance, nil
}

// DeleteRoleInheritance handles deleting a role inheritance
func (s *roleInheritanceService) DeleteRoleInheritance(id string) error {
	inheritanceID, err := parseUint64(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	// Check if inheritance exists
	_, err = s.repo.GetByID(inheritanceID)
	if err == sql.ErrNoRows {
		return gin.Error{
			Err:  fmt.Errorf("role inheritance not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return fmt.Errorf("failed to check role inheritance existence: %w", err)
	}

	return s.repo.Delete(inheritanceID)
}

// parseUint64 is a helper function to parse uint64 from string
func parseUint64(s string) (uint64, error) {
	var id uint64
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}
