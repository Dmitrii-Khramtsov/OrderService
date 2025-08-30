// github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/errors.go
package http

import "errors"

var (
	ErrInvalidJSON      = errors.New("invalid JSON format")
	ErrJSONEncodeFailed = errors.New("failed to encode JSON response")
)
