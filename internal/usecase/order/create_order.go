package order

import (
	"context"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
	apperrors "online-order-management-system/pkg/errors"
	"online-order-management-system/pkg/logger"
)

// CreateOrderUseCase handles the business logic for creating orders
type CreateOrderUseCase struct {
	orderRepo repository.OrderRepository
	logger    *logger.Logger
}

// NewCreateOrderUseCase creates a new CreateOrderUseCase
func NewCreateOrderUseCase(orderRepo repository.OrderRepository) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepo: orderRepo,
		logger:    logger.New("create-order-usecase", "1.0.0"),
	}
}

// CreateOrderRequest represents the input for creating an order
type CreateOrderRequest struct {
	CustomerName string                   `json:"customer_name" binding:"required"`
	Items        []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

// CreateOrderItemRequest represents an order item in the request
type CreateOrderItemRequest struct {
	ProductName string  `json:"product_name" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" binding:"required,min=0"`
}

// Execute creates a new order
func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest) (*entity.Order, error) {
	uc.logger.WithFields(map[string]interface{}{
		"customer_name": req.CustomerName,
		"items_count":   len(req.Items),
	}).Info("Starting order creation")

	// Validate request
	if err := uc.validateCreateOrderRequest(req); err != nil {
		uc.logger.WithError(err).WithField("customer_name", req.CustomerName).Warn("Invalid order creation request")
		return nil, err
	}

	// Convert request items to domain entities
	items := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = entity.OrderItem{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	// Create order domain entity with business rules validation
	order, err := entity.NewOrder(req.CustomerName, items)
	if err != nil {
		uc.logger.WithError(err).WithField("customer_name", req.CustomerName).Error("Failed to create domain order entity")
		// Wrap domain errors
		return nil, apperrors.NewBusinessRuleViolationError(err.Error()).WithCause(err)
	}

	// Persist the order
	createdOrder, err := uc.orderRepo.CreateOrderWithItems(ctx, order)
	if err != nil {
		uc.logger.WithError(err).WithFields(map[string]interface{}{
			"customer_name": req.CustomerName,
			"total_amount":  order.TotalAmount,
		}).Error("Failed to persist order")
		return nil, err // Repository errors are already wrapped
	}

	uc.logger.WithFields(map[string]interface{}{
		"order_id":      createdOrder.ID,
		"customer_name": createdOrder.CustomerName,
		"total_amount":  createdOrder.TotalAmount,
		"items_count":   len(createdOrder.Items),
	}).Info("Successfully created order")

	return createdOrder, nil
}

// validateCreateOrderRequest validates the create order request
func (uc *CreateOrderUseCase) validateCreateOrderRequest(req CreateOrderRequest) error {
	if req.CustomerName == "" {
		return apperrors.NewInvalidEntityError("customer name is required")
	}

	if len(req.Items) == 0 {
		return apperrors.NewInvalidEntityError("at least one item is required")
	}

	for i, item := range req.Items {
		if item.ProductName == "" {
			return apperrors.NewInvalidEntityError("product name is required").WithDetails(map[string]interface{}{
				"item_index": i,
			})
		}
		if item.Quantity <= 0 {
			return apperrors.NewInvalidEntityError("quantity must be greater than 0").WithDetails(map[string]interface{}{
				"item_index": i,
				"quantity":   item.Quantity,
			})
		}
		if item.UnitPrice < 0 {
			return apperrors.NewInvalidEntityError("unit price must be 0 or greater").WithDetails(map[string]interface{}{
				"item_index": i,
				"unit_price": item.UnitPrice,
			})
		}
	}

	return nil
}
