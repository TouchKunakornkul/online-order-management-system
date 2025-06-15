package db

import (
	"context"
	"database/sql"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
	apperrors "online-order-management-system/pkg/errors"
	"online-order-management-system/pkg/logger"
	"online-order-management-system/pkg/retryutil"

	_ "github.com/lib/pq"
)

// PostgresOrderRepository implements the OrderRepository interface using PostgreSQL
type PostgresOrderRepository struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewPostgresOrderRepository creates a new PostgresOrderRepository
func NewPostgresOrderRepository(db *sql.DB) repository.OrderRepository {
	return &PostgresOrderRepository{
		db:     db,
		logger: logger.New("postgres-order-repository", "1.0.0"),
	}
}

// CreateOrderWithItems creates a new order with its items in a single transaction
// This method is designed to handle concurrent requests efficiently with retry logic
func (r *PostgresOrderRepository) CreateOrderWithItems(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	var createdOrder *entity.Order

	config := retryutil.DefaultRetryConfig()
	err := retryutil.RetryWithBackoff(ctx, config, func() error {
		var err error
		createdOrder, err = r.createOrderWithItemsInternal(ctx, order)
		return err
	})

	if err != nil {
		r.logger.WithError(err).WithField("customer_name", order.CustomerName).
			Error("Failed to create order with items after retries")
		return nil, apperrors.NewDatabaseTransactionError("Failed to create order").WithCause(err)
	}

	r.logger.WithFields(map[string]interface{}{
		"order_id":      createdOrder.ID,
		"customer_name": createdOrder.CustomerName,
		"total_amount":  createdOrder.TotalAmount,
		"items_count":   len(createdOrder.Items),
	}).Info("Successfully created order with items")

	return createdOrder, nil
}

// createOrderWithItemsInternal is the internal implementation without retry logic
func (r *PostgresOrderRepository) createOrderWithItemsInternal(ctx context.Context, order *entity.Order) (*entity.Order, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, apperrors.NewDatabaseConnectionError("Failed to begin transaction").WithCause(err)
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
		return nil, apperrors.NewDatabaseQueryError("Failed to insert order").WithCause(err)
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
			return nil, apperrors.NewDatabaseQueryError("Failed to insert order item").WithCause(err)
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
		return nil, apperrors.NewDatabaseTransactionError("Failed to commit transaction").WithCause(err)
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
			r.logger.WithField("order_id", id).Warn("Order not found")
			return nil, apperrors.NewNotFoundError("order")
		}
		r.logger.WithError(err).WithField("order_id", id).Error("Failed to get order")
		return nil, apperrors.NewDatabaseQueryError("Failed to get order").WithCause(err)
	}

	// Get order items
	items, err := r.getOrderItems(ctx, id)
	if err != nil {
		r.logger.WithError(err).WithField("order_id", id).Error("Failed to get order items")
		return nil, err
	}
	order.Items = items

	r.logger.WithFields(map[string]interface{}{
		"order_id":    order.ID,
		"items_count": len(order.Items),
	}).Debug("Successfully retrieved order by ID")

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
		r.logger.WithError(err).Error("Failed to get total count of orders")
		return nil, nil, apperrors.NewDatabaseQueryError("Failed to get total count").WithCause(err)
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
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"page":   page,
			"limit":  limit,
			"offset": offset,
		}).Error("Failed to list orders")
		return nil, nil, apperrors.NewDatabaseQueryError("Failed to list orders").WithCause(err)
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
			r.logger.WithError(err).Error("Failed to scan order")
			return nil, nil, apperrors.NewDatabaseQueryError("Failed to scan order").WithCause(err)
		}

		// Get items for each order
		items, err := r.getOrderItems(ctx, order.ID)
		if err != nil {
			r.logger.WithError(err).WithField("order_id", order.ID).Error("Failed to get order items")
			return nil, nil, err
		}
		order.Items = items

		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		r.logger.WithError(err).Error("Error iterating orders")
		return nil, nil, apperrors.NewDatabaseQueryError("Error iterating orders").WithCause(err)
	}

	r.logger.WithFields(map[string]interface{}{
		"page":         page,
		"limit":        limit,
		"total_count":  totalCount,
		"total_pages":  totalPages,
		"orders_count": len(orders),
	}).Debug("Successfully listed orders")

	return orders, paginationInfo, nil
}

// UpdateOrderStatus updates the status of an existing order
func (r *PostgresOrderRepository) UpdateOrderStatus(ctx context.Context, id int64, status string) error {
	query := `
		UPDATE orders 
		SET status = $1, updated_at = NOW()
		WHERE id = $2`

	result, err := r.db.ExecContext(ctx, query, status, id)
	if err != nil {
		r.logger.WithError(err).WithFields(map[string]interface{}{
			"order_id": id,
			"status":   status,
		}).Error("Failed to update order status")
		return apperrors.NewDatabaseQueryError("Failed to update order status").WithCause(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.WithError(err).WithField("order_id", id).Error("Failed to get rows affected")
		return apperrors.NewDatabaseQueryError("Failed to get rows affected").WithCause(err)
	}

	if rowsAffected == 0 {
		r.logger.WithField("order_id", id).Warn("Order not found for status update")
		return apperrors.NewNotFoundError("order")
	}

	r.logger.WithFields(map[string]interface{}{
		"order_id": id,
		"status":   status,
	}).Info("Successfully updated order status")

	return nil
}

// getOrderItems retrieves order items for a specific order
func (r *PostgresOrderRepository) getOrderItems(ctx context.Context, orderID int64) ([]entity.OrderItem, error) {
	itemsQuery := `
		SELECT id, order_id, product_name, quantity, unit_price, total_price
		FROM order_items
		WHERE order_id = $1
		ORDER BY id`

	rows, err := r.db.QueryContext(ctx, itemsQuery, orderID)
	if err != nil {
		return nil, apperrors.NewDatabaseQueryError("Failed to get order items").WithCause(err)
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
			return nil, apperrors.NewDatabaseQueryError("Failed to scan order item").WithCause(err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return nil, apperrors.NewDatabaseQueryError("Error iterating order items").WithCause(err)
	}

	return items, nil
}
