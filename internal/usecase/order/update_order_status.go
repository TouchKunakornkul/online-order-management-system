package order

import (
	"context"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
)

// UpdateOrderStatusUseCase handles the business logic for updating order status
type UpdateOrderStatusUseCase struct {
	orderRepo repository.OrderRepository
}

// NewUpdateOrderStatusUseCase creates a new UpdateOrderStatusUseCase
func NewUpdateOrderStatusUseCase(orderRepo repository.OrderRepository) *UpdateOrderStatusUseCase {
	return &UpdateOrderStatusUseCase{
		orderRepo: orderRepo,
	}
}

// UpdateOrderStatusRequest represents the input for updating order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending processing completed cancelled"`
}

// Execute updates the order status
func (uc *UpdateOrderStatusUseCase) Execute(ctx context.Context, id int64, status string) error {
	// Validate status using domain rules
	tempOrder := &entity.Order{}
	if err := tempOrder.UpdateStatus(status); err != nil {
		return err
	}

	// Update in repository
	return uc.orderRepo.UpdateOrderStatus(ctx, id, status)
}
