package entities

import (
	"errors"
	"time"

	"github.com/shopspring/decimal"
)

// MenuItem represents a single item on the room service menu
type MenuItem struct {
	ID          string
	Description string
	Price       decimal.Decimal
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// NewMenuItem creates a new menu item with validation
func NewMenuItem(id, description string, price decimal.Decimal) (*MenuItem, error) {
	if id == "" {
		return nil, errors.New("menu item ID cannot be empty")
	}
	if description == "" {
		return nil, errors.New("menu item description cannot be empty")
	}
	if price.LessThanOrEqual(decimal.Zero) {
		return nil, errors.New("menu item price must be greater than zero")
	}

	now := time.Now()
	return &MenuItem{
		ID:          id,
		Description: description,
		Price:       price,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// UpdateDescription updates the menu item description
func (m *MenuItem) UpdateDescription(description string) error {
	if description == "" {
		return errors.New("menu item description cannot be empty")
	}
	m.Description = description
	m.UpdatedAt = time.Now()
	return nil
}

// UpdatePrice updates the menu item price
func (m *MenuItem) UpdatePrice(price decimal.Decimal) error {
	if price.LessThanOrEqual(decimal.Zero) {
		return errors.New("menu item price must be greater than zero")
	}
	m.Price = price
	m.UpdatedAt = time.Now()
	return nil
}

// Menu represents a collection of menu items
type Menu struct {
	Items []MenuItem
}

// NewMenu creates a new menu
func NewMenu() *Menu {
	return &Menu{
		Items: make([]MenuItem, 0),
	}
}

// AddItem adds a menu item to the menu
func (m *Menu) AddItem(item MenuItem) {
	m.Items = append(m.Items, item)
}

// FindItemByID finds a menu item by ID
func (m *Menu) FindItemByID(id string) (*MenuItem, error) {
	for i := range m.Items {
		if m.Items[i].ID == id {
			return &m.Items[i], nil
		}
	}
	return nil, errors.New("menu item not found")
}

// RemoveItem removes a menu item from the menu by ID
func (m *Menu) RemoveItem(id string) error {
	for i, item := range m.Items {
		if item.ID == id {
			m.Items = append(m.Items[:i], m.Items[i+1:]...)
			return nil
		}
	}
	return errors.New("menu item not found")
}
