package order

import (
	"context"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
)

// GetOrderUseCase handles the business logic for retrieving orders
type GetOrderUseCase struct {
	orderRepo repository.OrderRepository
}

// NewGetOrderUseCase creates a new GetOrderUseCase
func NewGetOrderUseCase(orderRepo repository.OrderRepository) *GetOrderUseCase {
	return &GetOrderUseCase{
		orderRepo: orderRepo,
	}
}

// Execute retrieves an order by ID
func (uc *GetOrderUseCase) Execute(ctx context.Context, id int64) (*entity.Order, error) {
	order, err := uc.orderRepo.GetOrderByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return order, nil
}
