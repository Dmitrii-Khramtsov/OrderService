// github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/errors.go
package http

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrCodeInvalidJSON      ErrorCode = "invalid_json"
	ErrCodeJSONEncodeFailed ErrorCode = "json_encode_failed"
	ErrCodeInvalidRequest   ErrorCode = "invalid_request"
	ErrCodeOrderNotFound    ErrorCode = "order_not_found"
	ErrCodeInternalError    ErrorCode = "internal_error"
)

type HTTPError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details string    `json:"details,omitempty"`
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewHTTPError(code ErrorCode, message string, details string) *HTTPError {
	return &HTTPError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

var (
	ErrInvalidJSON      = errors.New("invalid JSON format")
	ErrJSONEncodeFailed = errors.New("failed to encode JSON response")
)
