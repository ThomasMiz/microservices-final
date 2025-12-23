package services

import (
	"context"
	"errors"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"
)

// MenuService handles menu-related business logic
type MenuService struct {
	menuRepo repositories.MenuRepository
}

// NewMenuService creates a new menu service
func NewMenuService(menuRepo repositories.MenuRepository) *MenuService {
	return &MenuService{
		menuRepo: menuRepo,
	}
}

// ValidateMenuItem validates a menu item exists
func (s *MenuService) ValidateMenuItem(ctx context.Context, itemID string) (*entities.MenuItem, error) {
	item, err := s.menuRepo.FindByID(ctx, itemID)
	if err != nil {
		return nil, err
	}
	if item == nil {
		return nil, errors.New("menu item not found")
	}
	return item, nil
}

// ValidateMenuItems validates multiple menu items exist
func (s *MenuService) ValidateMenuItems(ctx context.Context, itemIDs []string) ([]entities.MenuItem, error) {
	var items []entities.MenuItem

	for _, itemID := range itemIDs {
		item, err := s.ValidateMenuItem(ctx, itemID)
		if err != nil {
			return nil, err
		}
		items = append(items, *item)
	}

	return items, nil
}

// GetAllMenuItems retrieves all menu items
func (s *MenuService) GetAllMenuItems(ctx context.Context) ([]entities.MenuItem, error) {
	return s.menuRepo.FindAll(ctx)
}
