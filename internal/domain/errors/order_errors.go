package errors

import (
	apperrors "online-order-management-system/pkg/errors"
)

// Order domain specific error codes
const (
	ErrCodeInvalidOrderStatus = "INVALID_ORDER_STATUS"
	ErrCodeOrderNotFound      = "ORDER_NOT_FOUND"
	ErrCodeInvalidQuantity    = "INVALID_QUANTITY"
	ErrCodeInvalidPrice       = "INVALID_PRICE"
	ErrCodeEmptyCustomerName  = "EMPTY_CUSTOMER_NAME"
	ErrCodeEmptyItems         = "EMPTY_ITEMS"
	ErrCodeEmptyProductName   = "EMPTY_PRODUCT_NAME"
)

// Order-specific error constructors
func NewOrderNotFoundError(orderID int64) *apperrors.AppError {
	return apperrors.NewNotFoundError("order not found").WithDetails(map[string]interface{}{
		"order_id": orderID,
	})
}

func NewCustomerNameRequiredError() *apperrors.AppError {
	return apperrors.NewInvalidEntityError("customer name is required")
}

func NewEmptyOrderItemsError() *apperrors.AppError {
	return apperrors.NewInvalidEntityError("order must have at least one item")
}

func NewProductNameRequiredError(itemIndex int) *apperrors.AppError {
	return apperrors.NewInvalidEntityError("product name is required").WithDetails(map[string]interface{}{
		"item_index": itemIndex,
	})
}

func NewInvalidQuantityError(itemIndex int, quantity int) *apperrors.AppError {
	return apperrors.NewInvalidEntityError("quantity must be greater than 0").WithDetails(map[string]interface{}{
		"item_index": itemIndex,
		"quantity":   quantity,
	})
}

func NewInvalidUnitPriceError(itemIndex int, unitPrice float64) *apperrors.AppError {
	return apperrors.NewInvalidEntityError("unit price cannot be negative").WithDetails(map[string]interface{}{
		"item_index": itemIndex,
		"unit_price": unitPrice,
	})
}

func NewInvalidOrderStatusError(status string, validStatuses []string) *apperrors.AppError {
	return apperrors.NewBusinessRuleViolationError("invalid order status").WithDetails(map[string]interface{}{
		"provided_status": status,
		"valid_statuses":  validStatuses,
	})
}

func NewInvalidOrderIDError(orderID int64) *apperrors.AppError {
	return apperrors.NewInvalidOperationError("order ID must be greater than 0").WithDetails(map[string]interface{}{
		"provided_id": orderID,
	})
}
