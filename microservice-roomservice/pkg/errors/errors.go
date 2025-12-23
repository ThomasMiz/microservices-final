package errors

import "fmt"

// DomainError represents a business logic error
type DomainError struct {
	Code    string
	Message string
}

func (e *DomainError) Error() string {
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

// NewDomainError creates a new domain error
func NewDomainError(code, message string) *DomainError {
	return &DomainError{
		Code:    code,
		Message: message,
	}
}

// Common error codes
const (
	ErrorCodeValidation     = "VALIDATION_ERROR"
	ErrorCodeNotFound       = "NOT_FOUND"
	ErrorCodeInternalServer = "INTERNAL_SERVER_ERROR"
	ErrorCodeInvalidStatus  = "INVALID_STATUS"
	ErrorCodeUnauthorized   = "UNAUTHORIZED"
)

// Common errors
var (
	ErrMenuItemNotFound    = NewDomainError(ErrorCodeNotFound, "menu item not found")
	ErrOrderNotFound       = NewDomainError(ErrorCodeNotFound, "order not found")
	ErrReservationNotFound = NewDomainError(ErrorCodeNotFound, "reservation not found")
	ErrInvalidPrice        = NewDomainError(ErrorCodeValidation, "price must be greater than zero")
	ErrInvalidQuantity     = NewDomainError(ErrorCodeValidation, "quantity must be greater than zero")
	ErrEmptyOrder          = NewDomainError(ErrorCodeValidation, "order must have at least one item")
	ErrInvalidOrderStatus  = NewDomainError(ErrorCodeInvalidStatus, "invalid order status")
)
