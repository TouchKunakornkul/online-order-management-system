package handler

import (
	"context"
	"net/http"
	"online-order-management-system/internal/api/http/handler/dto"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/usecase/order"
	"strconv"
	"strings"
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
	Execute(ctx context.Context, page int, limit int) (*order.ListOrdersResponse, error)
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

// getValidationErrorMessage returns user-friendly error messages for validation failures
func getValidationErrorMessage(err error) string {
	errStr := err.Error()

	// Handle status validation errors
	if strings.Contains(errStr, "oneof") && strings.Contains(errStr, "Status") {
		return "Invalid status. Must be one of: pending, processing, completed, cancelled"
	}

	// Handle required field errors
	if strings.Contains(errStr, "required") {
		if strings.Contains(errStr, "CustomerName") {
			return "Customer name is required"
		}
		if strings.Contains(errStr, "Items") {
			return "At least one item is required"
		}
		if strings.Contains(errStr, "ProductName") {
			return "Product name is required for all items"
		}
		if strings.Contains(errStr, "Status") {
			return "Status is required"
		}
	}

	// Handle items array validation errors
	if strings.Contains(errStr, "min") && strings.Contains(errStr, "Items") {
		return "At least one item is required"
	}

	// Handle quantity validation errors
	if strings.Contains(errStr, "min") && strings.Contains(errStr, "Quantity") {
		return "Quantity must be at least 1"
	}

	// Handle price validation errors
	if strings.Contains(errStr, "min") && strings.Contains(errStr, "UnitPrice") {
		return "Unit price must be 0 or greater"
	}

	// Handle JSON parsing errors
	if strings.Contains(errStr, "invalid character") || strings.Contains(errStr, "unexpected end of JSON") {
		return "Invalid JSON format in request body"
	}

	// Default to original error if no specific handling
	return errStr
}

// CreateOrder handles POST /orders
// @Summary      Create a new order
// @Description  Create a new order with customer information and items
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        order  body      dto.CreateOrderRequest  true  "Order creation request"
// @Success      201    {object}  dto.OrderResponse       "Order created successfully"
// @Failure      400    {object}  dto.ErrorResponse       "Invalid request body"
// @Failure      500    {object}  dto.ErrorResponse       "Internal server error"
// @Router       /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		friendlyError := getValidationErrorMessage(err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: friendlyError})
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
// @Summary      Get an order by ID
// @Description  Retrieve a specific order by its ID
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id   path      int                 true  "Order ID"
// @Success      200  {object}  dto.OrderResponse   "Order retrieved successfully"
// @Failure      400  {object}  dto.ErrorResponse   "Invalid order ID"
// @Failure      404  {object}  dto.ErrorResponse   "Order not found"
// @Failure      500  {object}  dto.ErrorResponse   "Internal server error"
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid order ID. Must be a valid number"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	domainOrder, err := h.getOrderUC.Execute(ctx, id)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Order not found"})
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
// @Summary      List orders with pagination
// @Description  Retrieve a paginated list of orders using page number and limit
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        page    query     int     false  "Page number (default: 1, min: 1)"
// @Param        limit   query     int     false  "Number of orders to return (default: 10, max: 100)"
// @Success      200     {object}  dto.ListOrdersResponse  "Orders retrieved successfully"
// @Failure      500     {object}  dto.ErrorResponse       "Internal server error"
// @Router       /orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	// Parse query parameters
	pageStr := c.DefaultQuery("page", "1")
	page, err := strconv.Atoi(pageStr)
	if err != nil || page <= 0 {
		page = 1
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 10
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	useCaseResponse, err := h.listOrdersUC.Execute(ctx, page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	// Convert usecase response to DTO response
	response := dto.FromUseCaseListOrdersResponse(useCaseResponse)
	c.JSON(http.StatusOK, response)
}

// UpdateOrderStatus handles PUT /orders/:id/status
// @Summary      Update order status
// @Description  Update the status of an existing order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id      path      int                           true  "Order ID"
// @Param        status  body      dto.UpdateOrderStatusRequest  true  "Status update request"
// @Success      200     {object}  dto.SuccessResponse           "Order status updated successfully"
// @Failure      400     {object}  dto.ErrorResponse             "Invalid request"
// @Failure      404     {object}  dto.ErrorResponse             "Order not found"
// @Failure      500     {object}  dto.ErrorResponse             "Internal server error"
// @Router       /orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "Invalid order ID. Must be a valid number"})
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		friendlyError := getValidationErrorMessage(err)
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: friendlyError})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	err = h.updateOrderStatusUC.Execute(ctx, id, req.Status)
	if err != nil {
		if err.Error() == "order not found" {
			c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "Order not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "Order status updated successfully"})
}
