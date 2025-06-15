package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"online-order-management-system/internal/api/http/handler/dto"
	"online-order-management-system/internal/api/validation"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/usecase/order"
	apperrors "online-order-management-system/pkg/errors"
	"online-order-management-system/pkg/logger"

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
	createOrderUC       *order.CreateOrderUseCase
	getOrderUC          *order.GetOrderUseCase
	listOrdersUC        *order.ListOrdersUseCase
	updateOrderStatusUC *order.UpdateOrderStatusUseCase
	logger              *logger.Logger
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(
	createOrderUC *order.CreateOrderUseCase,
	getOrderUC *order.GetOrderUseCase,
	listOrdersUC *order.ListOrdersUseCase,
	updateOrderStatusUC *order.UpdateOrderStatusUseCase,
) *OrderHandler {
	return &OrderHandler{
		createOrderUC:       createOrderUC,
		getOrderUC:          getOrderUC,
		listOrdersUC:        listOrdersUC,
		updateOrderStatusUC: updateOrderStatusUC,
		logger:              logger.New("order-handler", "1.0.0"),
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

// getTraceID extracts trace ID from gin context
func getTraceID(c *gin.Context) string {
	if traceID, exists := c.Get("trace_id"); exists {
		if str, ok := traceID.(string); ok {
			return str
		}
	}
	return ""
}

// CreateOrder handles POST /orders
// @Summary      Create a new order
// @Description  Create a new order with customer information and items
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        order  body      dto.CreateOrderRequest  true  "Order creation request"
// @Success      201    {object}  dto.OrderResponse       "Order created successfully"
// @Failure      400    {object}  apperrors.ErrorResponse       "Invalid request body"
// @Failure      500    {object}  apperrors.ErrorResponse       "Internal server error"
// @Router       /orders [post]
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	traceID := getTraceID(c)

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).WithField("trace_id", traceID).Warn("Invalid request body")
		friendlyError := validation.GetOrderValidationMessage(err)
		validationErr := apperrors.NewValidationError(friendlyError)
		response := apperrors.ToErrorResponse(validationErr, traceID)
		c.JSON(validationErr.HTTPStatus, response)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	// Convert DTO to usecase request
	useCaseReq := req.ToUseCaseCreateOrderRequest()
	createdOrder, err := h.createOrderUC.Execute(ctx, useCaseReq)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"trace_id":      traceID,
			"customer_name": req.CustomerName,
			"items_count":   len(req.Items),
		}).Error("Failed to create order")

		response := apperrors.ToErrorResponse(err, traceID)
		statusCode := apperrors.GetHTTPStatus(err)
		c.JSON(statusCode, response)
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"trace_id":      traceID,
		"order_id":      createdOrder.ID,
		"customer_name": createdOrder.CustomerName,
		"total_amount":  createdOrder.TotalAmount,
	}).Info("Successfully created order")

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
// @Failure      400  {object}  apperrors.ErrorResponse   "Invalid order ID"
// @Failure      404  {object}  apperrors.ErrorResponse   "Order not found"
// @Failure      500  {object}  apperrors.ErrorResponse   "Internal server error"
// @Router       /orders/{id} [get]
func (h *OrderHandler) GetOrder(c *gin.Context) {
	traceID := getTraceID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"trace_id": traceID,
			"id_param": idStr,
		}).Warn("Invalid order ID parameter")

		validationErr := apperrors.NewValidationError("Invalid order ID. Must be a valid number")
		response := apperrors.ToErrorResponse(validationErr, traceID)
		c.JSON(validationErr.HTTPStatus, response)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	domainOrder, err := h.getOrderUC.Execute(ctx, id)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"trace_id": traceID,
			"order_id": id,
		}).Error("Failed to get order")

		response := apperrors.ToErrorResponse(err, traceID)
		statusCode := apperrors.GetHTTPStatus(err)
		c.JSON(statusCode, response)
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"trace_id": traceID,
		"order_id": domainOrder.ID,
	}).Debug("Successfully retrieved order")

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
// @Failure      500     {object}  apperrors.ErrorResponse       "Internal server error"
// @Router       /orders [get]
func (h *OrderHandler) ListOrders(c *gin.Context) {
	traceID := getTraceID(c)

	// Parse query parameters
	page := 1
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	result, err := h.listOrdersUC.Execute(ctx, page, limit)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"trace_id": traceID,
			"page":     page,
			"limit":    limit,
		}).Error("Failed to list orders")

		response := apperrors.ToErrorResponse(err, traceID)
		statusCode := apperrors.GetHTTPStatus(err)
		c.JSON(statusCode, response)
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"trace_id":     traceID,
		"page":         page,
		"limit":        limit,
		"orders_count": len(result.Orders),
		"total_count":  result.Pagination.TotalCount,
	}).Debug("Successfully listed orders")

	// Convert to DTO response
	response := dto.ListOrdersResponse{
		Orders:     make([]dto.OrderResponse, len(result.Orders)),
		Pagination: dto.FromDomainPaginationInfo(result.Pagination),
	}

	for i, order := range result.Orders {
		response.Orders[i] = dto.FromDomainOrder(order)
	}

	c.JSON(http.StatusOK, response)
}

// UpdateOrderStatus handles PATCH /orders/:id/status
// @Summary      Update order status
// @Description  Update the status of an existing order
// @Tags         orders
// @Accept       json
// @Produce      json
// @Param        id      path      int                            true  "Order ID"
// @Param        status  body      dto.UpdateOrderStatusRequest  true  "Status update request"
// @Success      200     {object}  dto.SuccessResponse            "Order status updated successfully"
// @Failure      400     {object}  apperrors.ErrorResponse              "Invalid request"
// @Failure      404     {object}  apperrors.ErrorResponse              "Order not found"
// @Failure      500     {object}  apperrors.ErrorResponse              "Internal server error"
// @Router       /orders/{id}/status [patch]
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	traceID := getTraceID(c)

	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"trace_id": traceID,
			"id_param": idStr,
		}).Warn("Invalid order ID parameter")

		validationErr := apperrors.NewValidationError("Invalid order ID. Must be a valid number")
		response := apperrors.ToErrorResponse(validationErr, traceID)
		c.JSON(validationErr.HTTPStatus, response)
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"trace_id": traceID,
			"order_id": id,
		}).Warn("Invalid request body for status update")

		friendlyError := validation.GetOrderValidationMessage(err)
		validationErr := apperrors.NewValidationError(friendlyError)
		response := apperrors.ToErrorResponse(validationErr, traceID)
		c.JSON(validationErr.HTTPStatus, response)
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	err = h.updateOrderStatusUC.Execute(ctx, id, req.Status)
	if err != nil {
		h.logger.WithError(err).WithFields(map[string]interface{}{
			"trace_id": traceID,
			"order_id": id,
			"status":   req.Status,
		}).Error("Failed to update order status")

		response := apperrors.ToErrorResponse(err, traceID)
		statusCode := apperrors.GetHTTPStatus(err)
		c.JSON(statusCode, response)
		return
	}

	h.logger.WithFields(map[string]interface{}{
		"trace_id": traceID,
		"order_id": id,
		"status":   req.Status,
	}).Info("Successfully updated order status")

	c.JSON(http.StatusOK, dto.SuccessResponse{Message: "Order status updated successfully"})
}
