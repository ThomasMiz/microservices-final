package mocks

import (
	"context"
	"microservice-roomservice/internal/domain/entities"

	"github.com/stretchr/testify/mock"
)

// MockMenuRepository is a mock implementation of MenuRepository
type MockMenuRepository struct {
	mock.Mock
}

// Save mocks the Save method
func (m *MockMenuRepository) Save(ctx context.Context, item *entities.MenuItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

// Update mocks the Update method
func (m *MockMenuRepository) Update(ctx context.Context, item *entities.MenuItem) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

// Delete mocks the Delete method
func (m *MockMenuRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// FindByID mocks the FindByID method
func (m *MockMenuRepository) FindByID(ctx context.Context, id string) (*entities.MenuItem, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.MenuItem), args.Error(1)
}

// FindAll mocks the FindAll method
func (m *MockMenuRepository) FindAll(ctx context.Context) ([]entities.MenuItem, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.MenuItem), args.Error(1)
}

// ExistsByID mocks the ExistsByID method
func (m *MockMenuRepository) ExistsByID(ctx context.Context, id string) (bool, error) {
	args := m.Called(ctx, id)
	return args.Bool(0), args.Error(1)
}
