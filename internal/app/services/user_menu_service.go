package services

import (
	"database/sql"
	"fmt"
	"strconv"

	"adminbe/internal/app/models"
	"adminbe/internal/app/repositories"

	"github.com/gin-gonic/gin"
)

// UserMenuService interface defines business logic for user menus
type UserMenuService interface {
	ListUserMenus() ([]models.UserMenu, error)
	GetUserMenu(userIDStr, menuIDStr string) (*models.UserMenu, error)
	CreateUserMenu(req models.CreateUserMenuRequest) (*models.UserMenu, error)
	DeleteUserMenu(userIDStr, menuIDStr string) error
}

// userMenuService implements UserMenuService
type userMenuService struct {
	repo repositories.UserMenuRepository
}

// NewUserMenuService creates a new user menu service
func NewUserMenuService(repo repositories.UserMenuRepository) UserMenuService {
	return &userMenuService{repo: repo}
}

// ListUserMenus handles listing all user-menu assignments
func (s *userMenuService) ListUserMenus() ([]models.UserMenu, error) {
	userMenus, err := s.repo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to get user menus: %w", err)
	}

	return userMenus, nil
}

// GetUserMenu handles getting a user-menu assignment by user and menu IDs
func (s *userMenuService) GetUserMenu(userIDStr, menuIDStr string) (*models.UserMenu, error) {
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID: %w", err)
	}

	menuID, err := parseUint(menuIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid menu ID: %w", err)
	}

	userMenu, err := s.repo.GetByUserAndMenu(userID, menuID)
	if err == sql.ErrNoRows {
		return nil, gin.Error{
			Err:  fmt.Errorf("user-menu assignment not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user menu: %w", err)
	}

	return userMenu, nil
}

// CreateUserMenu handles creating a new user-menu assignment
func (s *userMenuService) CreateUserMenu(req models.CreateUserMenuRequest) (*models.UserMenu, error) {
	// Check if assignment already exists
	existing, err := s.repo.GetByUserAndMenu(req.UserID, req.MenuID)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to check user menu existence: %w", err)
	}
	if existing != nil {
		return nil, gin.Error{
			Err:  fmt.Errorf("user-menu assignment already exists"),
			Type: gin.ErrorTypePublic,
		}
	}

	userMenu := models.UserMenu{
		UserID: req.UserID,
		MenuID: req.MenuID,
		// deleted_at and deleted_by are nil for active
	}

	err = s.repo.Create(userMenu)
	if err != nil {
		return nil, fmt.Errorf("failed to create user menu: %w", err)
	}

	// Return the created assignment
	createdUserMenu, err := s.repo.GetByUserAndMenu(req.UserID, req.MenuID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve created user menu: %w", err)
	}

	return createdUserMenu, nil
}

// DeleteUserMenu handles deleting a user-menu assignment
func (s *userMenuService) DeleteUserMenu(userIDStr, menuIDStr string) error {
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid user ID: %w", err)
	}

	menuID, err := parseUint(menuIDStr)
	if err != nil {
		return fmt.Errorf("invalid menu ID: %w", err)
	}

	// Check if assignment exists
	_, err = s.repo.GetByUserAndMenu(userID, menuID)
	if err == sql.ErrNoRows {
		return gin.Error{
			Err:  fmt.Errorf("user-menu assignment not found"),
			Type: gin.ErrorTypePublic,
		}
	}
	if err != nil {
		return fmt.Errorf("failed to check user menu existence: %w", err)
	}

	return s.repo.Delete(userID, menuID, nil) // TODO: get current user ID for audit
}
