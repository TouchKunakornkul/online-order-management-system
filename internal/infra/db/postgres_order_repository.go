package db

import (
	"context"
	"database/sql"
	"fmt"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// PostgresOrderRepository implements the OrderRepository interface using PostgreSQL
type PostgresOrderRepository struct {
	db *sql.DB
}

// NewPostgresOrderRepository creates a new PostgresOrderRepository
func NewPostgresOrderRepository(db *sql.DB) repository.OrderRepository {
	return &PostgresOrderRepository{
		db: db,
	}
}

// isConnectionError checks if the error is related to database connection limits
func isConnectionError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "too many clients already") ||
		strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset")
}

// retryWithBackoff executes a function with exponential backoff retry logic
func retryWithBackoff(ctx context.Context, maxRetries int, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter
			backoff := time.Duration(attempt*attempt) * 10 * time.Millisecond
			if backoff > 500*time.Millisecond {
				backoff = 500 * time.Millisecond
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Only retry on connection errors
		if !isConnectionError(err) {
			return err
		}
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// CreateOrderWithItems creates a new order with its items in a single transaction
// This method is designed to handle concurrent requests efficiently with retry logic
func (r *PostgresOrderRepository) CreateOrderWithItems(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	var createdOrder *entity.Order

	err := retryWithBackoff(ctx, 3, func() error {
		var err error
		createdOrder, err = r.createOrderWithItemsInternal(ctx, order)
		return err
	})

	return createdOrder, err
}

// createOrderWithItemsInternal is the internal implementation without retry logic
func (r *PostgresOrderRepository) createOrderWithItemsInternal(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert order
	orderQuery := `
		INSERT INTO orders (customer_name, total_amount, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var orderID int64
	err = tx.QueryRowContext(ctx, orderQuery,
		order.CustomerName,
		order.TotalAmount,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
	).Scan(&orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to insert order: %w", err)
	}

	// Insert order items
	itemQuery := `
		INSERT INTO order_items (order_id, product_name, quantity, unit_price, total_price)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	items := make([]entity.OrderItem, len(order.Items))
	for i, item := range order.Items {
		var itemID int64
		err = tx.QueryRowContext(ctx, itemQuery,
			orderID,
			item.ProductName,
			item.Quantity,
			item.UnitPrice,
			item.TotalPrice,
		).Scan(&itemID)
		if err != nil {
			return nil, fmt.Errorf("failed to insert order item: %w", err)
		}

		items[i] = entity.OrderItem{
			ID:          itemID,
			OrderID:     orderID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the created order with IDs
	createdOrder := &entity.Order{
		ID:           orderID,
		CustomerName: order.CustomerName,
		TotalAmount:  order.TotalAmount,
		Status:       order.Status,
		Items:        items,
		CreatedAt:    order.CreatedAt,
		UpdatedAt:    order.UpdatedAt,
	}

	return createdOrder, nil
}

// GetOrderByID retrieves an order by its ID including its items
func (r *PostgresOrderRepository) GetOrderByID(ctx context.Context, id int64) (*entity.Order, error) {
	// Get order
	orderQuery := `
		SELECT id, customer_name, total_amount, status, created_at, updated_at
		FROM orders
		WHERE id = $1`

	var order entity.Order
	err := r.db.QueryRowContext(ctx, orderQuery, id).Scan(
		&order.ID,
		&order.CustomerName,
		&order.TotalAmount,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("order not found")
		}
		return nil, fmt.Errorf("failed to get order: %w", err)
	}

	// Get order items
	itemsQuery := `
		SELECT id, order_id, product_name, quantity, unit_price, total_price
		FROM order_items
		WHERE order_id = $1
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, itemsQuery, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []entity.OrderItem
	for rows.Next() {
		var item entity.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	order.Items = items
	return &order, nil
}

// ListOrders retrieves orders with pagination using page number and limit
func (r *PostgresOrderRepository) ListOrders(ctx context.Context, page int, limit int) ([]*entity.Order, *repository.PaginationInfo, error) {
	// Validate page number (must be >= 1)
	if page < 1 {
		page = 1
	}

	// Calculate offset
	offset := (page - 1) * limit

	// Get total count first
	countQuery := `SELECT COUNT(*) FROM orders`
	var totalCount int64
	err := r.db.QueryRowContext(ctx, countQuery).Scan(&totalCount)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get total count: %w", err)
	}

	// Calculate pagination info
	totalPages := int((totalCount + int64(limit) - 1) / int64(limit)) // Ceiling division
	if totalPages == 0 {
		totalPages = 1
	}

	paginationInfo := &repository.PaginationInfo{
		CurrentPage:  page,
		TotalPages:   totalPages,
		TotalCount:   totalCount,
		ItemsPerPage: limit,
	}

	// Get orders with pagination
	query := `
		SELECT id, customer_name, total_amount, status, created_at, updated_at
		FROM orders
		ORDER BY created_at DESC, id DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to list orders: %w", err)
	}
	defer rows.Close()

	var orders []*entity.Order
	for rows.Next() {
		order := &entity.Order{}
		err := rows.Scan(
			&order.ID,
			&order.CustomerName,
			&order.TotalAmount,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to scan order: %w", err)
		}

		// Get items for each order
		items, err := r.getOrderItems(ctx, order.ID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to get order items: %w", err)
		}
		order.Items = items

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("error iterating orders: %w", err)
	}

	return orders, paginationInfo, nil
}

// UpdateOrderStatus updates the status of an existing order
func (r *PostgresOrderRepository) UpdateOrderStatus(ctx context.Context, id int64, status string) error {
	query := `
		UPDATE orders
		SET status = $1, updated_at = $2
		WHERE id = $3`

	result, err := r.db.ExecContext(ctx, query, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("order not found")
	}

	return nil
}

// getOrderItems retrieves order items for a given order ID
func (r *PostgresOrderRepository) getOrderItems(ctx context.Context, orderID int64) ([]entity.OrderItem, error) {
	query := `
		SELECT id, order_id, product_name, quantity, unit_price, total_price
		FROM order_items
		WHERE order_id = $1
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, query, orderID)
	if err != nil {
		return nil, fmt.Errorf("failed to query order items: %w", err)
	}
	defer rows.Close()

	var items []entity.OrderItem
	for rows.Next() {
		var item entity.OrderItem
		err := rows.Scan(
			&item.ID,
			&item.OrderID,
			&item.ProductName,
			&item.Quantity,
			&item.UnitPrice,
			&item.TotalPrice,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating order items: %w", err)
	}

	return items, nil
}
