package entity

import (
	"errors"
	apperrors "online-order-management-system/pkg/errors"
	"time"
)

// Order represents the order domain entity
type Order struct {
	ID           int64       `json:"id"`
	CustomerName string      `json:"customer_name"`
	Status       string      `json:"status"`
	TotalAmount  float64     `json:"total_amount"`
	Items        []OrderItem `json:"items"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
}

// OrderItem represents an order item domain entity
type OrderItem struct {
	ID          int64   `json:"id"`
	OrderID     int64   `json:"order_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

// ValidStatuses defines the valid order statuses
var ValidStatuses = []string{"pending", "processing", "completed", "cancelled"}

// Domain errors
var (
	ErrInvalidCustomerName = errors.New("customer name is required")
	ErrEmptyItems          = errors.New("order must have at least one item")
	ErrInvalidQuantity     = errors.New("item quantity must be greater than 0")
	ErrInvalidUnitPrice    = errors.New("item unit price cannot be negative")
	ErrInvalidStatus       = errors.New("invalid order status")
)

// NewOrder creates a new order with validation
func NewOrder(customerName string, items []OrderItem) (*Order, error) {
	if customerName == "" {
		return nil, apperrors.NewInvalidEntityError("customer name is required").WithCause(ErrInvalidCustomerName)
	}
	if len(items) == 0 {
		return nil, apperrors.NewInvalidEntityError("order must have at least one item").WithCause(ErrEmptyItems)
	}

	// Calculate total amount
	var totalAmount float64
	for i := range items {
		if items[i].ProductName == "" {
			return nil, apperrors.NewInvalidEntityError("product name is required").WithDetails(map[string]interface{}{
				"item_index": i,
			})
		}
		if items[i].Quantity <= 0 {
			return nil, apperrors.NewInvalidEntityError("item quantity must be greater than 0").WithDetails(map[string]interface{}{
				"item_index": i,
				"quantity":   items[i].Quantity,
			}).WithCause(ErrInvalidQuantity)
		}
		if items[i].UnitPrice < 0 {
			return nil, apperrors.NewInvalidEntityError("item unit price cannot be negative").WithDetails(map[string]interface{}{
				"item_index": i,
				"unit_price": items[i].UnitPrice,
			}).WithCause(ErrInvalidUnitPrice)
		}
		items[i].TotalPrice = float64(items[i].Quantity) * items[i].UnitPrice
		totalAmount += items[i].TotalPrice
	}

	return &Order{
		CustomerName: customerName,
		Status:       "pending",
		TotalAmount:  totalAmount,
		Items:        items,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

// UpdateStatus updates the order status with validation
func (o *Order) UpdateStatus(status string) error {
	if !isValidStatus(status) {
		return apperrors.NewBusinessRuleViolationError("invalid order status").WithDetails(map[string]interface{}{
			"provided_status": status,
			"valid_statuses":  ValidStatuses,
		}).WithCause(ErrInvalidStatus)
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
}

// IsValidStatus checks if the status is valid (public for external validation)
func IsValidStatus(status string) bool {
	return isValidStatus(status)
}

// isValidStatus checks if the status is valid
func isValidStatus(status string) bool {
	for _, validStatus := range ValidStatuses {
		if status == validStatus {
			return true
		}
	}
	return false
}

// CalculateTotalAmount recalculates the total amount based on items
func (o *Order) CalculateTotalAmount() {
	var total float64
	for _, item := range o.Items {
		total += item.TotalPrice
	}
	o.TotalAmount = total
	o.UpdatedAt = time.Now()
}

// Validate performs comprehensive validation of the order entity
func (o *Order) Validate() error {
	if o.CustomerName == "" {
		return apperrors.NewInvalidEntityError("customer name is required").WithCause(ErrInvalidCustomerName)
	}

	if len(o.Items) == 0 {
		return apperrors.NewInvalidEntityError("order must have at least one item").WithCause(ErrEmptyItems)
	}

	if !isValidStatus(o.Status) {
		return apperrors.NewBusinessRuleViolationError("invalid order status").WithDetails(map[string]interface{}{
			"current_status": o.Status,
			"valid_statuses": ValidStatuses,
		}).WithCause(ErrInvalidStatus)
	}

	for i, item := range o.Items {
		if item.ProductName == "" {
			return apperrors.NewInvalidEntityError("product name is required").WithDetails(map[string]interface{}{
				"item_index": i,
			})
		}
		if item.Quantity <= 0 {
			return apperrors.NewInvalidEntityError("item quantity must be greater than 0").WithDetails(map[string]interface{}{
				"item_index": i,
				"quantity":   item.Quantity,
			}).WithCause(ErrInvalidQuantity)
		}
		if item.UnitPrice < 0 {
			return apperrors.NewInvalidEntityError("item unit price cannot be negative").WithDetails(map[string]interface{}{
				"item_index": i,
				"unit_price": item.UnitPrice,
			}).WithCause(ErrInvalidUnitPrice)
		}
	}

	return nil
}
