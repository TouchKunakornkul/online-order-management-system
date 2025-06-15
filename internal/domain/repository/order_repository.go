package repository

import (
	"context"
	"online-order-management-system/internal/domain/entity"
)

// OrderRepository defines the contract for order data access operations
type OrderRepository interface {
	// CreateOrderWithItems creates a new order with its items in a single transaction
	CreateOrderWithItems(ctx context.Context, order *entity.Order) (*entity.Order, error)

	// GetOrderByID retrieves an order by its ID including its items
	GetOrderByID(ctx context.Context, id int64) (*entity.Order, error)

	// ListOrders retrieves orders with pagination using cursor-based pagination
	ListOrders(ctx context.Context, limit int, cursor string) ([]*entity.Order, string, error)

	// UpdateOrderStatus updates the status of an existing order
	UpdateOrderStatus(ctx context.Context, id int64, status string) error
}
