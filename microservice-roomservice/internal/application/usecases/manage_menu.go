package usecases

import (
	"context"
	"microservice-roomservice/internal/domain/entities"
	"microservice-roomservice/internal/domain/repositories"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ManageMenu handles menu management operations
type ManageMenu struct {
	menuRepo repositories.MenuRepository
}

// NewManageMenu creates a new manage menu use case
func NewManageMenu(menuRepo repositories.MenuRepository) *ManageMenu {
	return &ManageMenu{
		menuRepo: menuRepo,
	}
}

// AddMenuItemRequest represents a request to add a menu item
type AddMenuItemRequest struct {
	Description string
	Price       decimal.Decimal
}

// AddMenuItem adds a new menu item
func (uc *ManageMenu) AddMenuItem(ctx context.Context, req AddMenuItemRequest) (*entities.MenuItem, error) {
	item, err := entities.NewMenuItem(
		uuid.New().String(),
		req.Description,
		req.Price,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.menuRepo.Save(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// UpdateMenuItemRequest represents a request to update a menu item
type UpdateMenuItemRequest struct {
	ID          string
	Description string
	Price       decimal.Decimal
}

// UpdateMenuItem updates an existing menu item
func (uc *ManageMenu) UpdateMenuItem(ctx context.Context, req UpdateMenuItemRequest) (*entities.MenuItem, error) {
	item, err := uc.menuRepo.FindByID(ctx, req.ID)
	if err != nil {
		return nil, err
	}

	if req.Description != "" {
		if err := item.UpdateDescription(req.Description); err != nil {
			return nil, err
		}
	}

	if !req.Price.IsZero() {
		if err := item.UpdatePrice(req.Price); err != nil {
			return nil, err
		}
	}

	if err := uc.menuRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// RemoveMenuItem removes a menu item
func (uc *ManageMenu) RemoveMenuItem(ctx context.Context, id string) error {
	return uc.menuRepo.Delete(ctx, id)
}

// GetMenuItem retrieves a menu item by ID
func (uc *ManageMenu) GetMenuItem(ctx context.Context, id string) (*entities.MenuItem, error) {
	return uc.menuRepo.FindByID(ctx, id)
}

// GetAllMenuItems retrieves all menu items
func (uc *ManageMenu) GetAllMenuItems(ctx context.Context) ([]entities.MenuItem, error) {
	return uc.menuRepo.FindAll(ctx)
}
