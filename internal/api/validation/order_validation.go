package validation

import (
	"online-order-management-system/pkg/validation"
	"strings"
)

// GetOrderValidationMessage returns order-specific user-friendly error messages
func GetOrderValidationMessage(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Handle order status validation errors
	if strings.Contains(errStr, "oneof") && strings.Contains(errStr, "Status") {
		return "Invalid status. Must be one of: pending, processing, completed, cancelled"
	}

	// Handle order-specific required fields
	if strings.Contains(errStr, "required") {
		if strings.Contains(errStr, "CustomerName") {
			return "Customer name is required"
		}
		if strings.Contains(errStr, "Items") {
			return "At least one item is required"
		}
		if strings.Contains(errStr, "ProductName") {
			return "Product name is required for all items"
		}
		if strings.Contains(errStr, "Status") {
			return "Status is required"
		}
		// Generic required field handling
		return "This field is required"
	}

	// Handle order items validation
	if strings.Contains(errStr, "min") {
		if strings.Contains(errStr, "Items") {
			return "At least one item is required"
		}
		if strings.Contains(errStr, "Quantity") {
			return "Quantity must be at least 1"
		}
		if strings.Contains(errStr, "UnitPrice") {
			return "Unit price must be 0 or greater"
		}
		// Generic min validation
		if strings.Contains(errStr, "array") || strings.Contains(errStr, "slice") {
			return "At least one item is required"
		}
		return "Value is too small"
	}

	// Handle max validation
	if strings.Contains(errStr, "max") {
		return "Value is too large"
	}

	// Handle email validation
	if strings.Contains(errStr, "email") {
		return "Invalid email format"
	}

	// Handle URL validation
	if strings.Contains(errStr, "url") {
		return "Invalid URL format"
	}

	// Handle oneof validation (generic case)
	if strings.Contains(errStr, "oneof") {
		return "Invalid value. Please check allowed values"
	}

	// Handle JSON parsing errors
	if strings.Contains(errStr, "invalid character") || strings.Contains(errStr, "unexpected end of JSON") {
		return "Invalid JSON format in request body"
	}

	// Default to original error if no specific handling
	return errStr
}

// Order field validation constants
const (
	MinQuantity  = 1
	MinUnitPrice = 0.0
	MinItems     = 1
)

// ValidateOrderFields performs order-specific field validation
func ValidateOrderFields(customerName string, items []interface{}) *validation.ValidationResult {
	result := validation.NewValidationResult()

	// Validate customer name
	if strings.TrimSpace(customerName) == "" {
		result.AddError(validation.NewFieldValidationError(
			"customer_name",
			"required",
			"Customer name is required",
			customerName,
		))
	}

	// Validate items
	if len(items) == 0 {
		result.AddError(validation.NewFieldValidationError(
			"items",
			"min",
			"At least one item is required",
			len(items),
		))
	}

	return result
}

// ValidateOrderItemFields performs order item specific validation
func ValidateOrderItemFields(itemIndex int, productName string, quantity int, unitPrice float64) *validation.ValidationResult {
	result := validation.NewValidationResult()

	// Validate product name
	if strings.TrimSpace(productName) == "" {
		result.AddError(validation.NewFieldValidationError(
			"product_name",
			"required",
			"Product name is required",
			productName,
		).WithDetails(map[string]interface{}{
			"item_index": itemIndex,
		}))
	}

	// Validate quantity
	if quantity < MinQuantity {
		result.AddError(validation.NewFieldValidationError(
			"quantity",
			"min",
			"Quantity must be at least 1",
			quantity,
		).WithDetails(map[string]interface{}{
			"item_index": itemIndex,
			"min_value":  MinQuantity,
		}))
	}

	// Validate unit price
	if unitPrice < MinUnitPrice {
		result.AddError(validation.NewFieldValidationError(
			"unit_price",
			"min",
			"Unit price must be 0 or greater",
			unitPrice,
		).WithDetails(map[string]interface{}{
			"item_index": itemIndex,
			"min_value":  MinUnitPrice,
		}))
	}

	return result
}
