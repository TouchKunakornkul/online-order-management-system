package dto

import (
	"time"
)

// CreateOrderRequest represents the API request for creating an order
type CreateOrderRequest struct {
	CustomerName  string                   `json:"customer_name" binding:"required"`
	CustomerEmail string                   `json:"customer_email" binding:"required,email"`
	Items         []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

// CreateOrderItemRequest represents an order item in the create request
type CreateOrderItemRequest struct {
	ProductName string  `json:"product_name" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" binding:"required,min=0"`
}

// UpdateOrderStatusRequest represents the API request for updating order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending processing completed cancelled"`
}

// OrderResponse represents the API response for a single order
type OrderResponse struct {
	ID            int64               `json:"id"`
	CustomerName  string              `json:"customer_name"`
	CustomerEmail string              `json:"customer_email"`
	Status        string              `json:"status"`
	TotalAmount   float64             `json:"total_amount"`
	Items         []OrderItemResponse `json:"items"`
	CreatedAt     time.Time           `json:"created_at"`
	UpdatedAt     time.Time           `json:"updated_at"`
}

// OrderItemResponse represents an order item in the API response
type OrderItemResponse struct {
	ID          int64   `json:"id"`
	OrderID     int64   `json:"order_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

// ListOrdersResponse represents the API response for listing orders
type ListOrdersResponse struct {
	Orders     []OrderResponse `json:"orders"`
	NextCursor string          `json:"next_cursor,omitempty"`
}

// ErrorResponse represents the API error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string `json:"message"`
}
