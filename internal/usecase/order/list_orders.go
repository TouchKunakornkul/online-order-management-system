package order

import (
	"context"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
	"online-order-management-system/pkg/logger"
)

// ListOrdersUseCase handles the business logic for listing orders
type ListOrdersUseCase struct {
	orderRepo repository.OrderRepository
	logger    *logger.Logger
}

// NewListOrdersUseCase creates a new ListOrdersUseCase
func NewListOrdersUseCase(orderRepo repository.OrderRepository) *ListOrdersUseCase {
	return &ListOrdersUseCase{
		orderRepo: orderRepo,
		logger:    logger.New("list-orders-usecase", "1.0.0"),
	}
}

// ListOrdersResponse represents the response for listing orders
type ListOrdersResponse struct {
	Orders     []*entity.Order            `json:"orders"`
	Pagination *repository.PaginationInfo `json:"pagination"`
}

// Execute retrieves orders with pagination
func (uc *ListOrdersUseCase) Execute(ctx context.Context, page int, limit int) (*ListOrdersResponse, error) {
	uc.logger.WithFields(map[string]interface{}{
		"page":  page,
		"limit": limit,
	}).Debug("Starting orders listing")

	// Validate and normalize pagination parameters
	originalPage, originalLimit := page, limit

	// Set default page if not provided or invalid
	if page <= 0 {
		page = 1
	}

	// Set default limit if not provided or invalid
	if limit <= 0 {
		limit = 10
	}

	// Set maximum limit to prevent abuse
	const maxLimit = 100
	if limit > maxLimit {
		limit = maxLimit
	}

	// Log parameter adjustments if any
	if page != originalPage || limit != originalLimit {
		uc.logger.WithFields(map[string]interface{}{
			"original_page":  originalPage,
			"original_limit": originalLimit,
			"adjusted_page":  page,
			"adjusted_limit": limit,
		}).Debug("Adjusted pagination parameters")
	}

	orders, paginationInfo, err := uc.orderRepo.ListOrders(ctx, page, limit)
	if err != nil {
		uc.logger.WithError(err).WithFields(map[string]interface{}{
			"page":  page,
			"limit": limit,
		}).Error("Failed to list orders")
		return nil, err // Repository errors are already wrapped
	}

	response := &ListOrdersResponse{
		Orders:     orders,
		Pagination: paginationInfo,
	}

	uc.logger.WithFields(map[string]interface{}{
		"page":         page,
		"limit":        limit,
		"orders_count": len(orders),
		"total_count":  paginationInfo.TotalCount,
		"total_pages":  paginationInfo.TotalPages,
	}).Debug("Successfully listed orders")

	return response, nil
}
