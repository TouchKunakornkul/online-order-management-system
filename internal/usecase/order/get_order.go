package order

import (
	"context"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
	apperrors "online-order-management-system/pkg/errors"
	"online-order-management-system/pkg/logger"
)

// GetOrderUseCase handles the business logic for retrieving orders
type GetOrderUseCase struct {
	orderRepo repository.OrderRepository
	logger    *logger.Logger
}

// NewGetOrderUseCase creates a new GetOrderUseCase
func NewGetOrderUseCase(orderRepo repository.OrderRepository) *GetOrderUseCase {
	return &GetOrderUseCase{
		orderRepo: orderRepo,
		logger:    logger.New("get-order-usecase", "1.0.0"),
	}
}

// Execute retrieves an order by its ID
func (uc *GetOrderUseCase) Execute(ctx context.Context, id int64) (*entity.Order, error) {
	uc.logger.WithField("order_id", id).Debug("Starting order retrieval")

	if id <= 0 {
		uc.logger.WithField("order_id", id).Warn("Invalid order ID")
		return nil, apperrors.NewInvalidOperationError("order ID must be greater than 0").WithDetails(map[string]interface{}{
			"provided_id": id,
		})
	}

	order, err := uc.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		uc.logger.WithError(err).WithField("order_id", id).Error("Failed to retrieve order")
		return nil, err // Repository errors are already wrapped
	}

	uc.logger.WithFields(map[string]interface{}{
		"order_id":      order.ID,
		"customer_name": order.CustomerName,
		"status":        order.Status,
		"items_count":   len(order.Items),
	}).Debug("Successfully retrieved order")

	return order, nil
}
