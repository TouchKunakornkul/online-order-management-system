package order

import (
	"context"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
)

// ListOrdersUseCase handles the business logic for listing orders
type ListOrdersUseCase struct {
	orderRepo repository.OrderRepository
}

// NewListOrdersUseCase creates a new ListOrdersUseCase
func NewListOrdersUseCase(orderRepo repository.OrderRepository) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		orderRepo: orderRepo,
	}
}

// ListOrdersResponse represents the response for listing orders
type ListOrdersResponse struct {
	Orders     []*entity.Order `json:"orders"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

// Execute retrieves orders with pagination
func (uc *ListOrdersUseCase) Execute(ctx context.Context, limit int, cursor string) (*ListOrdersResponse, error) {
	// Set default limit if not provided or invalid
	if limit <= 0 {
		limit = 10
	}

	// Set maximum limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	orders, nextCursor, err := uc.orderRepo.ListOrders(ctx, limit, cursor)
	if err != nil {
		return nil, err
	}

	return &ListOrdersResponse{
		Orders:     orders,
		NextCursor: nextCursor,
	}, nil
}
