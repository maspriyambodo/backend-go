package services

import (
	"database/sql"
	"fmt"

	"adminbe/internal/app/models"
	"adminbe/internal/app/repositories"

	"github.com/gin-gonic/gin"
)

// RoleMenuService interface defines business logic for role menus
type RoleMenuService interface {
	ListRoleMenus() ([]models.RoleMenu, error)
	GetRoleMenu(roleIDStr, menuIDStr string) (*models.RoleMenu, error)
	CreateRoleMenu(req models.CreateRoleMenuRequest) (*models.RoleMenu, error)
	DeleteRoleMenu(roleIDStr, menuIDStr string) error
}

// roleMenuService implements RoleMenuService
type roleMenuService struct {
	repo repositories.RoleMenuRepository
}

// NewRoleMenuService creates a new role menu service
func NewRoleMenuService(repo repositories.RoleMenuRepository) RoleMenuService {
	return &roleMenuService{repo: repo}
}

// ListRoleMenus handles listing all role-menu assignments
func (s *roleMenuService) ListRoleMenus() ([]models.RoleMenu, error) {
	roleMenus, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get role menus: %w", err)
	}

	return roleMenus, nil
}

// GetRoleMenu handles getting a role-menu assignment by role and menu IDs
func (s *roleMenuService) GetRoleMenu(roleIDStr, menuIDStr string) (*models.RoleMenu, error) {
	roleID, err := parseUint(roleIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid role ID: %w", err)
	}

	menuID, err := parseUint(menuIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid menu ID: %w", err)
	}

	roleMenu, err := s.repo.GetByRoleAndMenu(roleID, menuID)
	if err == sql.ErrNoRows {
		return nil, gin.Error{
			Err:  fmt.Errorf("role-menu assignment not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get role menu: %w", err)
	}

	return roleMenu, nil
}

// CreateRoleMenu handles creating a new role-menu assignment
func (s *roleMenuService) CreateRoleMenu(req models.CreateRoleMenuRequest) (*models.RoleMenu, error) {
	// Check if assignment already exists
	existing, err := s.repo.GetByRoleAndMenu(req.RoleID, req.MenuID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check role menu existence: %w", err)
	}
	if existing != nil {
		return nil, gin.Error{
			Err:  fmt.Errorf("role-menu assignment already exists"),
			Type: gin.ErrorTypePublic,
		}
	}

	roleMenu := models.RoleMenu{
		RoleID: req.RoleID,
		MenuID: req.MenuID,
		// deleted_at and deleted_by are nil for active
	}

	err = s.repo.Create(roleMenu)
	if err != nil {
		return nil, fmt.Errorf("failed to create role menu: %w", err)
	}

	// Return the created assignment
	createdRoleMenu, err := s.repo.GetByRoleAndMenu(req.RoleID, req.MenuID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created role menu: %w", err)
	}

	return createdRoleMenu, nil
}

// DeleteRoleMenu handles deleting a role-menu assignment
func (s *roleMenuService) DeleteRoleMenu(roleIDStr, menuIDStr string) error {
	roleID, err := parseUint(roleIDStr)
	if err != nil {
		return fmt.Errorf("invalid role ID: %w", err)
	}

	menuID, err := parseUint(menuIDStr)
	if err != nil {
		return fmt.Errorf("invalid menu ID: %w", err)
	}

	// Check if assignment exists
	_, err = s.repo.GetByRoleAndMenu(roleID, menuID)
	if err == sql.ErrNoRows {
		return gin.Error{
			Err:  fmt.Errorf("role-menu assignment not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return fmt.Errorf("failed to check role menu existence: %w", err)
	}

	return s.repo.Delete(roleID, menuID, nil) // TODO: get current user ID for audit
}
