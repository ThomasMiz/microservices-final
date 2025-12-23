package repositories

import (
	"context"
	"microservice-roomservice/internal/domain/entities"
)

// MenuRepository defines the interface for menu persistence
type MenuRepository interface {
	// Save saves a menu item
	Save(ctx context.Context, item *entities.MenuItem) error

	// FindByID finds a menu item by ID
	FindByID(ctx context.Context, id string) (*entities.MenuItem, error)

	// FindAll retrieves all menu items
	FindAll(ctx context.Context) ([]entities.MenuItem, error)

	// Update updates a menu item
	Update(ctx context.Context, item *entities.MenuItem) error

	// Delete deletes a menu item by ID
	Delete(ctx context.Context, id string) error

	// ExistsByID checks if a menu item exists by ID
	ExistsByID(ctx context.Context, id string) (bool, error)
}
