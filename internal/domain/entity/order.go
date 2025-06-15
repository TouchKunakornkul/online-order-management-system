package entity

import (
	"errors"
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

// NewOrder creates a new order with validation
func NewOrder(customerName string, items []OrderItem) (*Order, error) {
	if customerName == "" {
		return nil, errors.New("customer name is required")
	}
	if len(items) == 0 {
		return nil, errors.New("order must have at least one item")
	}

	// Calculate total amount
	var totalAmount float64
	for i := range items {
		if items[i].Quantity <= 0 {
			return nil, errors.New("item quantity must be greater than 0")
		}
		if items[i].UnitPrice < 0 {
			return nil, errors.New("item unit price cannot be negative")
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
		return errors.New("invalid order status")
	}
	o.Status = status
	o.UpdatedAt = time.Now()
	return nil
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
