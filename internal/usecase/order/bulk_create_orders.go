package order

import (
	"context"
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/domain/repository"
	"sync"
)

// BulkCreateOrdersUseCase handles the business logic for bulk creating orders
type BulkCreateOrdersUseCase struct {
	orderRepo repository.OrderRepository
}

// NewBulkCreateOrdersUseCase creates a new BulkCreateOrdersUseCase
func NewBulkCreateOrdersUseCase(orderRepo repository.OrderRepository) *BulkCreateOrdersUseCase {
	return &BulkCreateOrdersUseCase{
		orderRepo: orderRepo,
	}
}

// BulkCreateOrdersRequest represents the input for bulk creating orders
type BulkCreateOrdersRequest struct {
	Orders []CreateOrderRequest `json:"orders" binding:"required,min=1"`
}

// BulkCreateOrdersResponse represents the response for bulk order creation
type BulkCreateOrdersResponse struct {
	CreatedOrders []*entity.Order `json:"created_orders"`
	TotalCreated  int             `json:"total_created"`
	Errors        []string        `json:"errors,omitempty"`
}

// Execute creates multiple orders concurrently
func (uc *BulkCreateOrdersUseCase) Execute(ctx context.Context, req BulkCreateOrdersRequest) (*BulkCreateOrdersResponse, error) {
	// Convert requests to domain entities
	orders := make([]*entity.Order, 0, len(req.Orders))
	var errors []string
	var mu sync.Mutex

	// Process each order request
	for _, orderReq := range req.Orders {
		// Convert request items to domain entities
		items := make([]entity.OrderItem, len(orderReq.Items))
		for i, item := range orderReq.Items {
			items[i] = entity.OrderItem{
				ProductName: item.ProductName,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
			}
		}

		// Create order domain entity with business rules validation
		order, err := entity.NewOrder(orderReq.CustomerName, orderReq.CustomerEmail, items)
		if err != nil {
			mu.Lock()
			errors = append(errors, err.Error())
			mu.Unlock()
			continue
		}

		orders = append(orders, order)
	}

	// If no valid orders, return error response
	if len(orders) == 0 {
		return &BulkCreateOrdersResponse{
			CreatedOrders: []*entity.Order{},
			TotalCreated:  0,
			Errors:        errors,
		}, nil
	}

	// Bulk create orders in repository
	createdOrders, err := uc.orderRepo.BulkCreateOrders(ctx, orders)
	if err != nil {
		return nil, err
	}

	return &BulkCreateOrdersResponse{
		CreatedOrders: createdOrders,
		TotalCreated:  len(createdOrders),
		Errors:        errors,
	}, nil
}
