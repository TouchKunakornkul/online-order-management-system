package dto

import (
	"online-order-management-system/internal/domain/repository"
	"time"
)

// CreateOrderRequest represents the API request for creating an order
type CreateOrderRequest struct {
	CustomerName string                   `json:"customer_name" binding:"required" example:"John Doe" validate:"required"`
	Items        []CreateOrderItemRequest `json:"items" binding:"required,min=1" validate:"required,min=1"`
}

// CreateOrderItemRequest represents an order item in the create request
type CreateOrderItemRequest struct {
	ProductName string  `json:"product_name" binding:"required" example:"Laptop Computer" validate:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1" example:"2" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" binding:"required,min=0" example:"999.99" validate:"required,min=0"`
}

// UpdateOrderStatusRequest represents the API request for updating order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending processing completed cancelled" example:"processing" validate:"required,oneof=pending processing completed cancelled"`
}

// OrderResponse represents the API response for a single order
type OrderResponse struct {
	ID           int64               `json:"id" example:"12345"`
	CustomerName string              `json:"customer_name" example:"John Doe"`
	Status       string              `json:"status" example:"pending" enums:"pending,processing,completed,cancelled"`
	TotalAmount  float64             `json:"total_amount" example:"1999.98"`
	Items        []OrderItemResponse `json:"items"`
	CreatedAt    time.Time           `json:"created_at" example:"2023-06-15T10:30:00Z"`
	UpdatedAt    time.Time           `json:"updated_at" example:"2023-06-15T10:30:00Z"`
}

// OrderItemResponse represents an order item in the API response
type OrderItemResponse struct {
	ID          int64   `json:"id" example:"67890"`
	OrderID     int64   `json:"order_id" example:"12345"`
	ProductName string  `json:"product_name" example:"Laptop Computer"`
	Quantity    int     `json:"quantity" example:"2"`
	UnitPrice   float64 `json:"unit_price" example:"999.99"`
	TotalPrice  float64 `json:"total_price" example:"1999.98"`
}

// PaginationResponse represents pagination metadata in API responses
type PaginationResponse struct {
	CurrentPage  int   `json:"current_page" example:"1"`
	TotalPages   int   `json:"total_pages" example:"10"`
	TotalCount   int64 `json:"total_count" example:"95"`
	ItemsPerPage int   `json:"items_per_page" example:"10"`
}

// ListOrdersResponse represents the API response for listing orders
type ListOrdersResponse struct {
	Orders     []OrderResponse    `json:"orders"`
	Pagination PaginationResponse `json:"pagination"`
}

// ErrorResponse represents the API error response
type ErrorResponse struct {
	Error string `json:"error" example:"Invalid request parameters"`
}

// SuccessResponse represents a generic success response
type SuccessResponse struct {
	Message string `json:"message" example:"Operation completed successfully"`
}

// FromDomainPaginationInfo converts repository.PaginationInfo to PaginationResponse
func FromDomainPaginationInfo(info *repository.PaginationInfo) PaginationResponse {
	return PaginationResponse{
		CurrentPage:  info.CurrentPage,
		TotalPages:   info.TotalPages,
		TotalCount:   info.TotalCount,
		ItemsPerPage: info.ItemsPerPage,
	}
}
