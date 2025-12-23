package handlers

import (
	"net/http"

	"microservice-roomservice/internal/application/usecases"

	"github.com/gin-gonic/gin"
	"github.com/shopspring/decimal"
)

// MenuHandler handles menu-related HTTP requests
type MenuHandler struct {
	manageMenu *usecases.ManageMenu
}

// NewMenuHandler creates a new menu handler
func NewMenuHandler(manageMenu *usecases.ManageMenu) *MenuHandler {
	return &MenuHandler{
		manageMenu: manageMenu,
	}
}

// AddMenuItemRequest represents the request body for adding a menu item
type AddMenuItemRequest struct {
	Description string `json:"description" binding:"required"`
	Price       string `json:"price" binding:"required"`
}

// UpdateMenuItemRequest represents the request body for updating a menu item
type UpdateMenuItemRequest struct {
	Description string `json:"description"`
	Price       string `json:"price"`
}

// GetAllMenuItems handles GET /menu
func (h *MenuHandler) GetAllMenuItems(c *gin.Context) {
	items, err := h.manageMenu.GetAllMenuItems(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

// AddMenuItem handles POST /menu/items
func (h *MenuHandler) AddMenuItem(c *gin.Context) {
	var req AddMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	price, err := decimal.NewFromString(req.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price format"})
		return
	}

	item, err := h.manageMenu.AddMenuItem(c.Request.Context(), usecases.AddMenuItemRequest{
		Description: req.Description,
		Price:       price,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// UpdateMenuItem handles PUT /menu/items/:id
func (h *MenuHandler) UpdateMenuItem(c *gin.Context) {
	id := c.Param("id")

	var req UpdateMenuItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var price decimal.Decimal
	if req.Price != "" {
		var err error
		price, err = decimal.NewFromString(req.Price)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid price format"})
			return
		}
	}

	item, err := h.manageMenu.UpdateMenuItem(c.Request.Context(), usecases.UpdateMenuItemRequest{
		ID:          id,
		Description: req.Description,
		Price:       price,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteMenuItem handles DELETE /menu/items/:id
func (h *MenuHandler) DeleteMenuItem(c *gin.Context) {
	id := c.Param("id")

	if err := h.manageMenu.RemoveMenuItem(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
