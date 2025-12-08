package services

import (
	"database/sql"
	"fmt"
	"time"

	"adminbe/internal/app/models"
	"adminbe/internal/app/repositories"

	"github.com/gin-gonic/gin"
)

// MenuService interface defines business logic for menus
type MenuService interface {
	ListMenus() ([]models.Menu, error)
	GetMenu(id string) (*models.Menu, error)
	CreateMenu(req models.Menu) (*models.Menu, error)
	UpdateMenu(id string, req map[string]interface{}) (*models.Menu, error)
	DeleteMenu(id string) error
}

// menuService implements MenuService
type menuService struct {
	repo repositories.MenuRepository
}

// NewMenuService creates a new menu service
func NewMenuService(repo repositories.MenuRepository) MenuService {
	return &menuService{repo: repo}
}

// ListMenus handles listing all menus
func (s *menuService) ListMenus() ([]models.Menu, error) {
	menus, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get menus: %w", err)
	}

	return menus, nil
}

// GetMenu handles getting a menu by ID
func (s *menuService) GetMenu(id string) (*models.Menu, error) {
	menuID, err := parseUint(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	menu, err := s.repo.GetByID(menuID)
	if err == sql.ErrNoRows {
		return nil, gin.Error{
			Err:  fmt.Errorf("menu not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}

	return menu, nil
}

// CreateMenu handles creating a new menu
func (s *menuService) CreateMenu(req models.Menu) (*models.Menu, error) {
	s.setTimestamps(&req)

	menuID, err := s.repo.Create(req)
	if err != nil {
		return nil, fmt.Errorf("failed to create menu: %w", err)
	}

	return s.retrieveMenuByID(menuID)
}

// UpdateMenu handles updating an existing menu
func (s *menuService) UpdateMenu(id string, req map[string]interface{}) (*models.Menu, error) {
	menuID, err := parseUint(id)
	if err != nil {
		return nil, fmt.Errorf("invalid ID: %w", err)
	}

	if err := s.ensureMenuExists(menuID); err != nil {
		return nil, err
	}

	if err := s.repo.Update(menuID, req); err != nil {
		return nil, fmt.Errorf("failed to update menu: %w", err)
	}

	return s.retrieveMenuByID(menuID)
}

// DeleteMenu handles deleting a menu
func (s *menuService) DeleteMenu(id string) error {
	menuID, err := parseUint(id)
	if err != nil {
		return fmt.Errorf("invalid ID: %w", err)
	}

	if err := s.ensureMenuExists(menuID); err != nil {
		return err
	}

	return s.repo.Delete(menuID, nil) // TODO: get current user ID for audit
}

// parseUint is a helper function to parse uint from string
func parseUint(s string) (uint, error) {
	var id uint
	_, err := fmt.Sscanf(s, "%d", &id)
	return id, err
}

// setTimestamps sets created_at and updated_at on menu
func (s *menuService) setTimestamps(menu *models.Menu) {
	now := time.Now()
	menu.CreatedAt = &now
	menu.UpdatedAt = &now
}

// ensureMenuExists checks if a menu exists by ID
func (s *menuService) ensureMenuExists(menuID uint) error {
	_, err := s.repo.GetByID(menuID)
	if err == sql.ErrNoRows {
		return gin.Error{
			Err:  fmt.Errorf("menu not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return fmt.Errorf("failed to check menu existence: %w", err)
	}
	return nil
}

// retrieveMenuByID fetches a menu by ID with error formatting
func (s *menuService) retrieveMenuByID(menuID uint) (*models.Menu, error) {
	menu, err := s.repo.GetByID(menuID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve menu: %w", err)
	}
	return menu, nil
}
