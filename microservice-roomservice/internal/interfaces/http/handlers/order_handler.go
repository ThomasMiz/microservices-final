package handlers

import (
	"net/http"

	"microservice-roomservice/internal/application/usecases"
	"microservice-roomservice/internal/domain/services"
	"microservice-roomservice/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/trace"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	createOrder       *usecases.CreateOrder
	getOrderStatus    *usecases.GetOrderStatus
	updateOrderStatus *usecases.UpdateOrderStatus
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(
	createOrder *usecases.CreateOrder,
	getOrderStatus *usecases.GetOrderStatus,
	updateOrderStatus *usecases.UpdateOrderStatus,
) *OrderHandler {
	return &OrderHandler{
		createOrder:       createOrder,
		getOrderStatus:    getOrderStatus,
		updateOrderStatus: updateOrderStatus,
	}
}

// CreateOrderRequest represents the request body for creating an order
type CreateOrderRequest struct {
	RoomID int64       `json:"room_id" binding:"required"`
	Items  []OrderItem `json:"items" binding:"required,dive"`
}

// OrderItem represents an item in the order request
type OrderItem struct {
	MenuItemID string `json:"menu_item_id" binding:"required"`
	Quantity   int    `json:"quantity" binding:"required,min=1"`
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	ctx := c.Request.Context()
	span := trace.SpanFromContext(ctx)
	spanCtx := span.SpanContext()
	logger.CtxDebug(ctx).Msgf("OrderHandler.CreateOrder - Incoming span: traceID=%s spanID=%s isValid=%v isRecording=%v isSampled=%v",
		spanCtx.TraceID().String(),
		spanCtx.SpanID().String(),
		spanCtx.IsValid(),
		span.IsRecording(),
		spanCtx.IsSampled())

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Convert to domain request
	items := make([]services.OrderItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = services.OrderItemRequest{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
		}
	}

	order, err := h.createOrder.Execute(c.Request.Context(), services.CreateOrderRequest{
		RoomID: req.RoomID,
		Items:  items,
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, order)
}

// GetOrderByID handles GET /orders/:id
func (h *OrderHandler) GetOrderByID(c *gin.Context) {
	id := c.Param("id")

	order, err := h.getOrderStatus.ByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// GetOrdersByRoom handles GET /orders/room/:roomId
func (h *OrderHandler) GetOrdersByRoom(c *gin.Context) {
	roomID := c.Param("roomId")

	orders, err := h.getOrderStatus.ByRoomID(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, orders)
}
