package domain

import (
	"errors"
	"time"
)

// Domain entities and business rules for orders and order items.

var ErrInvalidOrder = errors.New("invalid order or order item")

// Order represents a customer order.
type Order struct {
	ID           int64       `json:"id"`
	CustomerName string      `json:"customer_name"`
	TotalAmount  float64     `json:"total_amount"`
	Status       string      `json:"status"`
	CreatedAt    time.Time   `json:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at"`
	Items        []OrderItem `json:"items"`
}

// ValidStatuses defines allowed order statuses.
var ValidStatuses = map[string]struct{}{
	"pending":   {},
	"paid":      {},
	"shipped":   {},
	"completed": {},
	"cancelled": {},
}

// IsValidStatus checks if a status is valid.
func IsValidStatus(status string) bool {
	_, ok := ValidStatuses[status]
	return ok
}
