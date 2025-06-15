package order

import (
	"context"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
	apperrors "online-order-management-system/pkg/errors"
	"online-order-management-system/pkg/logger"
)

// UpdateOrderStatusUseCase handles the business logic for updating order status
type UpdateOrderStatusUseCase struct {
	orderRepo repository.OrderRepository
	logger    *logger.Logger
}

// NewUpdateOrderStatusUseCase creates a new UpdateOrderStatusUseCase
func NewUpdateOrderStatusUseCase(orderRepo repository.OrderRepository) *UpdateOrderStatusUseCase {
	return &UpdateOrderStatusUseCase{
		orderRepo: orderRepo,
		logger:    logger.New("update-order-status-usecase", "1.0.0"),
	}
}

// UpdateOrderStatusRequest represents the input for updating order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending processing completed cancelled"`
}

// Execute updates the status of an order
func (uc *UpdateOrderStatusUseCase) Execute(ctx context.Context, id int64, status string) error {
	uc.logger.WithFields(map[string]interface{}{
		"order_id": id,
		"status":   status,
	}).Info("Starting order status update")

	// Validate inputs
	if id <= 0 {
		uc.logger.WithField("order_id", id).Warn("Invalid order ID")
		return apperrors.NewInvalidOperationError("order ID must be greater than 0").WithDetails(map[string]interface{}{
			"provided_id": id,
		})
	}

	if !entity.IsValidStatus(status) {
		uc.logger.WithFields(map[string]interface{}{
			"order_id":       id,
			"invalid_status": status,
			"valid_statuses": entity.ValidStatuses,
		}).Warn("Invalid order status")
		return apperrors.NewBusinessRuleViolationError("invalid order status").WithDetails(map[string]interface{}{
			"provided_status": status,
			"valid_statuses":  entity.ValidStatuses,
		})
	}

	// Update the order status
	err := uc.orderRepo.UpdateOrderStatus(ctx, id, status)
	if err != nil {
		uc.logger.WithError(err).WithFields(map[string]interface{}{
			"order_id": id,
			"status":   status,
		}).Error("Failed to update order status")
		return err // Repository errors are already wrapped
	}

	uc.logger.WithFields(map[string]interface{}{
		"order_id": id,
		"status":   status,
	}).Info("Successfully updated order status")

	return nil
}
