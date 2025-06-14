package handler

import (
	"context"
	"net/http"
	"online-order-management-system/internal/usecase/order"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	createOrderUC       *order.CreateOrderUseCase
	getOrderUC          *order.GetOrderUseCase
	listOrdersUC        *order.ListOrdersUseCase
	updateOrderStatusUC *order.UpdateOrderStatusUseCase
	bulkCreateOrdersUC  *order.BulkCreateOrdersUseCase
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(
	createOrderUC *order.CreateOrderUseCase,
	getOrderUC *order.GetOrderUseCase,
	listOrdersUC *order.ListOrdersUseCase,
	updateOrderStatusUC *order.UpdateOrderStatusUseCase,
	bulkCreateOrdersUC *order.BulkCreateOrdersUseCase,
) *OrderHandler {
	return &OrderHandler{
		createOrderUC:       createOrderUC,
		getOrderUC:          getOrderUC,
		listOrdersUC:        listOrdersUC,
		updateOrderStatusUC: updateOrderStatusUC,
		bulkCreateOrdersUC:  bulkCreateOrdersUC,
	}
}

// RegisterRoutes registers all order routes to the Gin router
func (h *OrderHandler) RegisterRoutes(router *gin.Engine) {
	orders := router.Group("/orders")
	{
		orders.POST("", h.CreateOrder)
		orders.POST("/bulk", h.BulkCreateOrders)
		orders.GET("", h.ListOrders)
		orders.GET("/:id", h.GetOrder)
		orders.PUT("/:id/status", h.UpdateOrderStatus)
	}
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req order.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	createdOrder, err := h.createOrderUC.Execute(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, createdOrder)
}

// GetOrder handles GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	order, err := h.getOrderUC.Execute(ctx, id)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, order)
}

// ListOrders handles GET /orders
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// Parse query parameters
	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	cursor := c.Query("cursor")

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	response, err := h.listOrdersUC.Execute(ctx, limit, cursor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}

// UpdateOrderStatus handles PUT /orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req order.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	err = h.updateOrderStatusUC.Execute(ctx, id, req.Status)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "order status updated successfully"})
}

// BulkCreateOrders handles POST /orders/bulk
func (h *OrderHandler) BulkCreateOrders(c *gin.Context) {
	var req order.BulkCreateOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Minute) // Longer timeout for bulk operations
	defer cancel()

	response, err := h.bulkCreateOrdersUC.Execute(ctx, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, response)
}
