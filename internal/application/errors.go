// github.com/Dmitrii-Khramtsov/orderservice/internal/application/errors.go
package application

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	ErrCodeOrderSaveFailed   ErrorCode = "order_save_failed"
	ErrCodeOrderDeleteFailed ErrorCode = "order_delete_failed"
	ErrCodeOrderReadFailed   ErrorCode = "order_read_failed"
	ErrCodeOrdersReadFailed  ErrorCode = "orders_read_failed"
	ErrCodeValidation        ErrorCode = "validation_error"
)

type AppError struct {
	Code    ErrorCode
	Message string
	Op      string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Op, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Op, e.Message)
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func NewAppError(code ErrorCode, message string, op string, err error) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
		Op:      op,
		Err:     err,
	}
}

var (
	ErrOrderSaveFailed   = errors.New("failed to save order")
	ErrOrderDeleteFailed = errors.New("failed to delete order")
	ErrOrderReadFailed   = errors.New("failed to read order")
)
