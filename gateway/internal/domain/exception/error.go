package exception

import (
	"errors"
	"fmt"
	"net/http"
)

type ApplicationError struct {
	Code       ErrorCode
	Message    string
	Details    string
	StatusCode int
	Err        error
	Metadata   map[string]interface{}
}

func (e *ApplicationError) Error() string {
	if e == nil {
		return ""
	}
	if e.Details != "" {
		return fmt.Sprintf("[%s] %s: %s", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("[%s] %s", e.Code, e.Message)
}

func (e *ApplicationError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

func NewApplicationError(code ErrorCode, message string, err error) *ApplicationError {
	return &ApplicationError{
		Code:       code,
		Message:    message,
		StatusCode: ErrorCodeToHTTPStatus(code),
		Err:        err,
		Metadata:   make(map[string]interface{}),
	}
}

func NewApplicationErrorWithStatus(code ErrorCode, message string, statusCode int, err error) *ApplicationError {
	return &ApplicationError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Err:        err,
		Metadata:   make(map[string]interface{}),
	}
}

func (e *ApplicationError) WithDetails(details string) *ApplicationError {
	if e == nil {
		return e
	}
	e.Details = details
	return e
}

func (e *ApplicationError) WithMetadata(key string, value interface{}) *ApplicationError {
	if e == nil {
		return e
	}
	if e.Metadata == nil {
		e.Metadata = make(map[string]interface{})
	}
	e.Metadata[key] = value
	return e
}

func ErrorCodeToHTTPStatus(code ErrorCode) int {
	switch code {
	case ErrorCodeUnauthorized, ErrorCodeInvalidToken, ErrorCodeTokenExpired, ErrorCodeMissingToken:
		return http.StatusUnauthorized
	case ErrorCodeForbidden, ErrorCodeInsufficientRole:
		return http.StatusForbidden
	case ErrorCodeValidationFailed, ErrorCodeInvalidInput, ErrorCodeMissingField, ErrorCodeInvalidFormat:
		return http.StatusBadRequest
	case ErrorCodeNotFound:
		return http.StatusNotFound
	case ErrorCodeAlreadyExists, ErrorCodeConflict:
		return http.StatusConflict
	case ErrorCodeInsufficientFunds, ErrorCodeInvalidTransaction, ErrorCodeInvalidState, ErrorCodeOperationFailed:
		return http.StatusUnprocessableEntity
	case ErrorCodeServiceUnavailable:
		return http.StatusServiceUnavailable
	case ErrorCodeDatabaseError, ErrorCodeExternalService, ErrorCodeInternal:
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

func IsApplicationError(err error) bool {
	var appErr *ApplicationError
	return errors.As(err, &appErr)
}

func GetApplicationError(err error) (*ApplicationError, bool) {
	var appErr *ApplicationError
	ok := errors.As(err, &appErr)
	return appErr, ok
}
