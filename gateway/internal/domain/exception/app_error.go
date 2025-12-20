package exception

import (
	"fmt"
)

type AppError struct {
	Code       string
	Message    string
	HTTPStatus int
	Err        error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewAppError(code, message string, httpStatus int, err error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Err:        err,
	}
}

// Convenience constructors for common statuses
func NewBadRequestError(code, message string, err error) *AppError {
	return NewAppError(code, message, 400, err)
}

func NewInternalError(code, message string, err error) *AppError {
	return NewAppError(code, message, 500, err)
}

// Predefined common errors used across the application
var (
	ErrEmptyPhoto        = NewBadRequestError("ERR_EMPTY_PHOTO", "photo cannot be empty", nil)
	ErrEmptyVideo        = NewBadRequestError("ERR_EMPTY_VIDEO", "video cannot be empty", nil)
	ErrEmptyImage        = NewBadRequestError("ERR_EMPTY_IMAGE", "image cannot be empty", nil)
	ErrInsufficientFunds = NewBadRequestError("ERR_INSUFFICIENT_FUNDS", "insufficient wallet balance", nil)
	ErrServiceNotFound   = NewBadRequestError("ERR_SERVICE_NOT_FOUND", "service not found", nil)
)
