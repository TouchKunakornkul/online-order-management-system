package order

import (
	"context"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
)

// CreateOrderUseCase handles the business logic for creating orders
type CreateOrderUseCase struct {
	orderRepo repository.OrderRepository
}

// NewCreateOrderUseCase creates a new CreateOrderUseCase
func NewCreateOrderUseCase(orderRepo repository.OrderRepository) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderRepo: orderRepo,
	}
}

// CreateOrderRequest represents the input for creating an order
type CreateOrderRequest struct {
	CustomerName string                   `json:"customer_name" binding:"required"`
	Items        []CreateOrderItemRequest `json:"items" binding:"required,min=1"`
}

// CreateOrderItemRequest represents an order item in the request
type CreateOrderItemRequest struct {
	ProductName string  `json:"product_name" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" binding:"required,min=0"`
}

// Execute creates a new order
func (uc *CreateOrderUseCase) Execute(ctx context.Context, req CreateOrderRequest) (*entity.Order, error) {
	// Convert request items to domain entities
	items := make([]entity.OrderItem, len(req.Items))
	for i, item := range req.Items {
		items[i] = entity.OrderItem{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	// Create order domain entity with business rules validation
	order, err := entity.NewOrder(req.CustomerName, items)
	if err != nil {
		return nil, err
	}

	// Persist the order
	createdOrder, err := uc.orderRepo.CreateOrderWithItems(ctx, order)
	if err != nil {
		return nil, err
	}

	return createdOrder, nil
}
