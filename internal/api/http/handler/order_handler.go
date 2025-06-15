package handler

import (
	"context"
	"net/http"
	"online-order-management-system/internal/api/http/handler/dto"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/usecase/order"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Use case interfaces for better testability
type CreateOrderUseCase interface {
	Execute(ctx context.Context, req order.CreateOrderRequest) (*entity.Order, error)
}

type GetOrderUseCase interface {
	Execute(ctx context.Context, id int64) (*entity.Order, error)
}

type ListOrdersUseCase interface {
	Execute(ctx context.Context, limit int, cursor string) (*order.ListOrdersResponse, error)
}

type UpdateOrderStatusUseCase interface {
	Execute(ctx context.Context, id int64, status string) error
}

// OrderHandler handles HTTP requests for order operations
type OrderHandler struct {
	createOrderUC       CreateOrderUseCase
	getOrderUC          GetOrderUseCase
	listOrdersUC        ListOrdersUseCase
	updateOrderStatusUC UpdateOrderStatusUseCase
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(
	createOrderUC CreateOrderUseCase,
	getOrderUC GetOrderUseCase,
	listOrdersUC ListOrdersUseCase,
	updateOrderStatusUC UpdateOrderStatusUseCase,
) *OrderHandler {
	return &OrderHandler{
		createOrderUC:       createOrderUC,
		getOrderUC:          getOrderUC,
		listOrdersUC:        listOrdersUC,
		updateOrderStatusUC: updateOrderStatusUC,
	}
}

// RegisterRoutes registers all order routes to the Gin router
func (h *OrderHandler) RegisterRoutes(router gin.IRouter) {
	orders := router.Group("/orders")
	{
		orders.POST("", h.CreateOrder)
		orders.GET("", h.ListOrders)
		orders.GET("/:id", h.GetOrder)
		orders.PUT("/:id/status", h.UpdateOrderStatus)
	}
}

// CreateOrder handles POST /orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Convert DTO to usecase request
	useCaseReq := req.ToUseCaseCreateOrderRequest()
	createdOrder, err := h.createOrderUC.Execute(ctx, useCaseReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Convert domain entity to DTO response
	response := dto.FromDomainOrder(createdOrder)
	c.JSON(http.StatusCreated, response)
}

// GetOrder handles GET /orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid order ID"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	domainOrder, err := h.getOrderUC.Execute(ctx, id)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Convert domain entity to DTO response
	response := dto.FromDomainOrder(domainOrder)
	c.JSON(http.StatusOK, response)
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

	useCaseResponse, err := h.listOrdersUC.Execute(ctx, limit, cursor)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Convert usecase response to DTO response
	response := dto.FromUseCaseListOrdersResponse(useCaseResponse)
	c.JSON(http.StatusOK, response)
}

// UpdateOrderStatus handles PUT /orders/:id/status
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid order ID"})
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	err = h.updateOrderStatusUC.Execute(ctx, id, req.Status)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "order status updated successfully"})
}
