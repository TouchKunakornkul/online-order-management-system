package repository

import (
	"context"
	"online-order-management-system/internal/domain/entity"
)

// PaginationInfo contains pagination metadata
type PaginationInfo struct {
	CurrentPage  int   `json:"current_page"`
	TotalPages   int   `json:"total_pages"`
	TotalCount   int64 `json:"total_count"`
	ItemsPerPage int   `json:"items_per_page"`
}

// OrderRepository defines the contract for order data access operations
type OrderRepository interface {
	// CreateOrderWithItems creates a new order with its items in a single transaction
	CreateOrderWithItems(ctx context.Context, order *entity.Order) (*entity.Order, error)

	// GetOrderByID retrieves an order by its ID including its items
	GetOrderByID(ctx context.Context, id int64) (*entity.Order, error)

	// ListOrders retrieves orders with pagination using page number and limit
	ListOrders(ctx context.Context, page int, limit int) ([]*entity.Order, *PaginationInfo, error)

	// UpdateOrderStatus updates the status of an existing order
	UpdateOrderStatus(ctx context.Context, id int64, status string) error
}
