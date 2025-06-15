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
	Orders     []*entity.Order            `json:"orders"`
	Pagination *repository.PaginationInfo `json:"pagination"`
}

// Execute retrieves orders with pagination
func (uc *ListOrdersUseCase) Execute(ctx context.Context, page int, limit int) (*ListOrdersResponse, error) {
	// Set default page if not provided or invalid
	if page <= 0 {
		page = 1
	}

	// Set default limit if not provided or invalid
	if limit <= 0 {
		limit = 10
	}

	// Set maximum limit to prevent abuse
	if limit > 100 {
		limit = 100
	}

	orders, paginationInfo, err := uc.orderRepo.ListOrders(ctx, page, limit)
	if err != nil {
		return nil, err
	}

	return &ListOrdersResponse{
		Orders:     orders,
		Pagination: paginationInfo,
	}, nil
}
