package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// ErrorType represents the type of error
type ErrorType string

const (
	// Domain layer errors
	ErrorTypeDomain ErrorType = "DOMAIN"

	// Use case layer errors
	ErrorTypeUseCase ErrorType = "USECASE"

	// Infrastructure layer errors
	ErrorTypeInfrastructure ErrorType = "INFRASTRUCTURE"

	// API layer errors
	ErrorTypeAPI ErrorType = "API"
)

// ErrorCode represents specific error codes
type ErrorCode string

const (
	// Generic domain errors
	ErrCodeInvalidEntity         ErrorCode = "INVALID_ENTITY"
	ErrCodeBusinessRuleViolation ErrorCode = "BUSINESS_RULE_VIOLATION"

	// Generic use case errors
	ErrCodeNotFound         ErrorCode = "NOT_FOUND"
	ErrCodeAlreadyExists    ErrorCode = "ALREADY_EXISTS"
	ErrCodeInvalidOperation ErrorCode = "INVALID_OPERATION"
	ErrCodePermissionDenied ErrorCode = "PERMISSION_DENIED"

	// Generic infrastructure errors
	ErrCodeDatabaseConnection  ErrorCode = "DATABASE_CONNECTION"
	ErrCodeDatabaseQuery       ErrorCode = "DATABASE_QUERY"
	ErrCodeDatabaseTransaction ErrorCode = "DATABASE_TRANSACTION"
	ErrCodeExternalService     ErrorCode = "EXTERNAL_SERVICE"
	ErrCodeTimeout             ErrorCode = "TIMEOUT"
	ErrCodeNetworkError        ErrorCode = "NETWORK_ERROR"

	// Generic API errors
	ErrCodeValidation     ErrorCode = "VALIDATION"
	ErrCodeAuthentication ErrorCode = "AUTHENTICATION"
	ErrCodeAuthorization  ErrorCode = "AUTHORIZATION"
	ErrCodeRateLimit      ErrorCode = "RATE_LIMIT"
	ErrCodeBadRequest     ErrorCode = "BAD_REQUEST"
	ErrCodeInternalError  ErrorCode = "INTERNAL_ERROR"
)

// AppError represents a structured application error
type AppError struct {
	Type       ErrorType              `json:"type"`
	Code       ErrorCode              `json:"code"`
	Message    string                 `json:"message"`
	Details    map[string]interface{} `json:"details,omitempty"`
	Cause      error                  `json:"-"`
	HTTPStatus int                    `json:"-"`
}

func (e *AppError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (caused by: %v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Cause
}

// Is implements error matching
func (e *AppError) Is(target error) bool {
	if target == nil {
		return false
	}

	if appErr, ok := target.(*AppError); ok {
		return e.Code == appErr.Code && e.Type == appErr.Type
	}

	return false
}

// WithDetails adds details to the error
func (e *AppError) WithDetails(details map[string]interface{}) *AppError {
	newErr := *e
	if newErr.Details == nil {
		newErr.Details = make(map[string]interface{})
	}
	for k, v := range details {
		newErr.Details[k] = v
	}
	return &newErr
}

// WithCause adds a cause to the error
func (e *AppError) WithCause(cause error) *AppError {
	newErr := *e
	newErr.Cause = cause
	return &newErr
}

// NewAppError creates a new application error
func NewAppError(errorType ErrorType, code ErrorCode, message string) *AppError {
	httpStatus := getHTTPStatusFromCode(code)
	return &AppError{
		Type:       errorType,
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

// getHTTPStatusFromCode maps error codes to HTTP status codes
func getHTTPStatusFromCode(code ErrorCode) int {
	switch code {
	case ErrCodeNotFound:
		return http.StatusNotFound
	case ErrCodeAlreadyExists:
		return http.StatusConflict
	case ErrCodeValidation, ErrCodeInvalidEntity, ErrCodeBusinessRuleViolation, ErrCodeBadRequest:
		return http.StatusBadRequest
	case ErrCodeAuthentication:
		return http.StatusUnauthorized
	case ErrCodeAuthorization, ErrCodePermissionDenied:
		return http.StatusForbidden
	case ErrCodeRateLimit:
		return http.StatusTooManyRequests
	case ErrCodeTimeout:
		return http.StatusRequestTimeout
	case ErrCodeDatabaseConnection, ErrCodeDatabaseQuery, ErrCodeDatabaseTransaction,
		ErrCodeExternalService, ErrCodeNetworkError, ErrCodeInternalError:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// Generic error constructors for each layer
func NewDomainError(code ErrorCode, message string) *AppError {
	return NewAppError(ErrorTypeDomain, code, message)
}

func NewUseCaseError(code ErrorCode, message string) *AppError {
	return NewAppError(ErrorTypeUseCase, code, message)
}

func NewInfrastructureError(code ErrorCode, message string) *AppError {
	return NewAppError(ErrorTypeInfrastructure, code, message)
}

func NewAPIError(code ErrorCode, message string) *AppError {
	return NewAppError(ErrorTypeAPI, code, message)
}

// Common error patterns (but still generic)
func NewInvalidEntityError(message string) *AppError {
	return NewDomainError(ErrCodeInvalidEntity, message)
}

func NewBusinessRuleViolationError(message string) *AppError {
	return NewDomainError(ErrCodeBusinessRuleViolation, message)
}

func NewNotFoundError(message string) *AppError {
	return NewUseCaseError(ErrCodeNotFound, message)
}

func NewAlreadyExistsError(message string) *AppError {
	return NewUseCaseError(ErrCodeAlreadyExists, message)
}

func NewInvalidOperationError(message string) *AppError {
	return NewUseCaseError(ErrCodeInvalidOperation, message)
}

func NewPermissionDeniedError(message string) *AppError {
	return NewUseCaseError(ErrCodePermissionDenied, message)
}

func NewDatabaseConnectionError(message string) *AppError {
	return NewInfrastructureError(ErrCodeDatabaseConnection, message)
}

func NewDatabaseQueryError(message string) *AppError {
	return NewInfrastructureError(ErrCodeDatabaseQuery, message)
}

func NewDatabaseTransactionError(message string) *AppError {
	return NewInfrastructureError(ErrCodeDatabaseTransaction, message)
}

func NewExternalServiceError(message string) *AppError {
	return NewInfrastructureError(ErrCodeExternalService, message)
}

func NewTimeoutError(message string) *AppError {
	return NewInfrastructureError(ErrCodeTimeout, message)
}

func NewNetworkError(message string) *AppError {
	return NewInfrastructureError(ErrCodeNetworkError, message)
}

func NewValidationError(message string) *AppError {
	return NewAPIError(ErrCodeValidation, message)
}

func NewAuthenticationError(message string) *AppError {
	return NewAPIError(ErrCodeAuthentication, message)
}

func NewAuthorizationError(message string) *AppError {
	return NewAPIError(ErrCodeAuthorization, message)
}

func NewRateLimitError(message string) *AppError {
	return NewAPIError(ErrCodeRateLimit, message)
}

func NewBadRequestError(message string) *AppError {
	return NewAPIError(ErrCodeBadRequest, message)
}

func NewInternalError(message string) *AppError {
	return NewAPIError(ErrCodeInternalError, message)
}

// Error handling utilities
func IsAppError(err error) bool {
	var appErr *AppError
	return errors.As(err, &appErr)
}

func GetAppError(err error) *AppError {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr
	}
	return nil
}

func GetHTTPStatus(err error) int {
	if appErr := GetAppError(err); appErr != nil {
		return appErr.HTTPStatus
	}
	return http.StatusInternalServerError
}

// Error response for API
type ErrorResponse struct {
	Error   ErrorInfo `json:"error"`
	TraceID string    `json:"trace_id,omitempty"`
}

type ErrorInfo struct {
	Code    ErrorCode              `json:"code"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

// ToErrorResponse converts an error to API error response
func ToErrorResponse(err error, traceID string) ErrorResponse {
	appErr := GetAppError(err)
	if appErr == nil {
		// Handle non-app errors
		return ErrorResponse{
			Error: ErrorInfo{
				Code:    ErrCodeInternalError,
				Message: "An internal error occurred",
			},
			TraceID: traceID,
		}
	}

	return ErrorResponse{
		Error: ErrorInfo{
			Code:    appErr.Code,
			Message: appErr.Message,
			Details: appErr.Details,
		},
		TraceID: traceID,
	}
}
