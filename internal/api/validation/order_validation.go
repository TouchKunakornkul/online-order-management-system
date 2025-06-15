package validation

import (
	"fmt"
	"reflect"
	"strings"

	"online-order-management-system/pkg/validation"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

// RegisterCustomValidations registers custom validation functions with Gin
func RegisterCustomValidations() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// Register custom validation for string max length
		v.RegisterValidation("maxlen", func(fl validator.FieldLevel) bool {
			param := fl.Param()
			if param == "" {
				return true
			}

			// Parse the max length parameter
			var maxLen int
			if _, err := fmt.Sscanf(param, "%d", &maxLen); err != nil {
				return true // If we can't parse, let it pass
			}

			// Check if field is a string and validate length
			if fl.Field().Kind() == reflect.String {
				return len(fl.Field().String()) <= maxLen
			}

			return true
		})
	}
}

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
			return "Product name is required"
		}
		if strings.Contains(errStr, "Quantity") {
			return "Quantity is required"
		}
		if strings.Contains(errStr, "UnitPrice") {
			return "Unit price is required"
		}
		return "This field is required"
	}

	// Handle length validation errors
	if strings.Contains(errStr, "max") || strings.Contains(errStr, "maxlen") {
		if strings.Contains(errStr, "CustomerName") {
			return "Customer name must not exceed 100 characters"
		}
		if strings.Contains(errStr, "ProductName") {
			return "Product name must not exceed 100 characters"
		}
		return "Field exceeds maximum allowed length"
	}

	// Handle minimum value validation errors
	if strings.Contains(errStr, "min") {
		if strings.Contains(errStr, "Quantity") {
			return "Quantity must be at least 1"
		}
		if strings.Contains(errStr, "UnitPrice") {
			return "Unit price must be greater than 0"
		}
		if strings.Contains(errStr, "Items") {
			return "At least one item is required"
		}
		return "Field does not meet minimum requirements"
	}

	// Handle numeric validation errors
	if strings.Contains(errStr, "number") {
		return "Field must be a valid number"
	}

	// Return the original error if no specific handling is found
	return err.Error()
}

// Order field validation constants
const (
	MinQuantity     = 1
	MinUnitPrice    = 0.0
	MinItems        = 1
	MaxCustomerName = 100
	MaxProductName  = 100
)

// ValidateOrderFields performs order-specific field validation
func ValidateOrderFields(customerName string, items []interface{}) *validation.ValidationResult {
	result := validation.NewValidationResult()

	// Validate customer name
	trimmedCustomerName := strings.TrimSpace(customerName)
	if trimmedCustomerName == "" {
		result.AddError(validation.NewFieldValidationError(
			"customer_name",
			"required",
			"Customer name is required",
			customerName,
		))
	} else if len(trimmedCustomerName) > MaxCustomerName {
		result.AddError(validation.NewFieldValidationError(
			"customer_name",
			"max",
			"Customer name cannot exceed 100 characters",
			customerName,
		).WithDetails(map[string]interface{}{
			"max_length":     MaxCustomerName,
			"current_length": len(trimmedCustomerName),
		}))
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
	trimmedProductName := strings.TrimSpace(productName)
	if trimmedProductName == "" {
		result.AddError(validation.NewFieldValidationError(
			"product_name",
			"required",
			"Product name is required",
			productName,
		).WithDetails(map[string]interface{}{
			"item_index": itemIndex,
		}))
	} else if len(trimmedProductName) > MaxProductName {
		result.AddError(validation.NewFieldValidationError(
			"product_name",
			"max",
			"Product name cannot exceed 100 characters",
			productName,
		).WithDetails(map[string]interface{}{
			"item_index":     itemIndex,
			"max_length":     MaxProductName,
			"current_length": len(trimmedProductName),
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
