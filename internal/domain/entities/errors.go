// github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities/errors.go
package entities

import "errors"

var (
	ErrOrderUIDRequired     = errors.New("order_uid is required")
	ErrTrackNumberRequired  = errors.New("track_number is required")
	ErrItemsEmpty           = errors.New("items cannot be empty")
	ErrInvalidPaymentAmount = errors.New("payment amount must be > 0")
	ErrInvalidEmailFormat   = errors.New("invalid email format")
	ErrInvalidPhoneFormat   = errors.New("phone must start with +")
)
