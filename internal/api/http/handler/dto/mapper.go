package dto

import (
	"online-order-management-system/internal/domain/entity"
	"online-order-management-system/internal/usecase/order"
)

// ToUseCaseCreateOrderRequest converts API DTO to usecase request
func (req *CreateOrderRequest) ToUseCaseCreateOrderRequest() order.CreateOrderRequest {
	items := make([]order.CreateOrderItemRequest, len(req.Items))
	for i, item := range req.Items {
		items[i] = order.CreateOrderItemRequest{
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
		}
	}

	return order.CreateOrderRequest{
		CustomerName: req.CustomerName,
		Items:        items,
	}
}

// ToUseCaseUpdateOrderStatusRequest converts API DTO to usecase request
func (req *UpdateOrderStatusRequest) ToUseCaseUpdateOrderStatusRequest() order.UpdateOrderStatusRequest {
	return order.UpdateOrderStatusRequest{
		Status: req.Status,
	}
}

// FromDomainOrder converts domain entity to API DTO
func FromDomainOrder(domainOrder *entity.Order) OrderResponse {
	items := make([]OrderItemResponse, len(domainOrder.Items))
	for i, item := range domainOrder.Items {
		items[i] = OrderItemResponse{
			ID:          item.ID,
			OrderID:     item.OrderID,
			ProductName: item.ProductName,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		}
	}

	return OrderResponse{
		ID:           domainOrder.ID,
		CustomerName: domainOrder.CustomerName,
		Status:       domainOrder.Status,
		TotalAmount:  domainOrder.TotalAmount,
		Items:        items,
		CreatedAt:    domainOrder.CreatedAt,
		UpdatedAt:    domainOrder.UpdatedAt,
	}
}

// FromDomainOrders converts multiple domain entities to API DTOs
func FromDomainOrders(domainOrders []*entity.Order) []OrderResponse {
	orders := make([]OrderResponse, len(domainOrders))
	for i, domainOrder := range domainOrders {
		orders[i] = FromDomainOrder(domainOrder)
	}
	return orders
}

// FromUseCaseListOrdersResponse converts usecase response to API DTO
func FromUseCaseListOrdersResponse(useCaseResponse *order.ListOrdersResponse) ListOrdersResponse {
	return ListOrdersResponse{
		Orders:     FromDomainOrders(useCaseResponse.Orders),
		Pagination: FromDomainPaginationInfo(useCaseResponse.Pagination),
	}
}
