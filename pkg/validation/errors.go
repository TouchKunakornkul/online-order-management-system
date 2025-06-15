package validation

import (
	"fmt"
)

// ValidationError represents a validation error with detailed information
type ValidationError struct {
	Field   string
	Value   interface{}
	Tag     string
	Message string
}

func (e ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("validation failed for field '%s' with value '%v' (tag: %s)",
		e.Field, e.Value, e.Tag)
}

// FieldValidationError represents a field-specific validation error
type FieldValidationError struct {
	Field   string                 `json:"field"`
	Value   interface{}            `json:"value,omitempty"`
	Tag     string                 `json:"tag"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e FieldValidationError) Error() string {
	return fmt.Sprintf("validation failed for field '%s': %s", e.Field, e.Message)
}

// NewFieldValidationError creates a new field validation error
func NewFieldValidationError(field, tag, message string, value interface{}) *FieldValidationError {
	return &FieldValidationError{
		Field:   field,
		Tag:     tag,
		Message: message,
		Value:   value,
	}
}

// WithDetails adds additional details to the validation error
func (e *FieldValidationError) WithDetails(details map[string]interface{}) *FieldValidationError {
	newErr := *e
	if newErr.Details == nil {
		newErr.Details = make(map[string]interface{})
	}
	for k, v := range details {
		newErr.Details[k] = v
	}
	return &newErr
}

// ValidationResult represents the result of validation with multiple potential errors
type ValidationResult struct {
	Valid  bool                    `json:"valid"`
	Errors []*FieldValidationError `json:"errors,omitempty"`
}

// HasErrors returns true if there are validation errors
func (r *ValidationResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// AddError adds a validation error to the result
func (r *ValidationResult) AddError(err *FieldValidationError) {
	r.Errors = append(r.Errors, err)
	r.Valid = false
}

// GetFirstError returns the first validation error if any
func (r *ValidationResult) GetFirstError() *FieldValidationError {
	if len(r.Errors) > 0 {
		return r.Errors[0]
	}
	return nil
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Errors: make([]*FieldValidationError, 0),
	}
}
