package domain

// OrderItem represents a single item in an order.
type OrderItem struct {
	ID          int64   `json:"id"`
	OrderID     int64   `json:"order_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
}
