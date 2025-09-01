// github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/handler/order_handler.go
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	domainrepo "github.com/Dmitrii-Khramtsov/orderservice/internal/domain/repository"
	httperrors "github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http"
)

type OrderHandler struct {
	svc    application.OrderServiceInterface
	logger domainrepo.Logger
}

func NewOrderHandler(s application.OrderServiceInterface, l domainrepo.Logger) *OrderHandler {
	return &OrderHandler{
		svc:    s,
		logger: l,
	}
}

func (h *OrderHandler) writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.logger.Error("failed to encode JSON response",
			"error", err,
			"status", status,
		)
		
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error"))
	}
}

func (h *OrderHandler) writeError(w http.ResponseWriter, status int, err *httperrors.HTTPError) {
	h.writeJSON(w, status, map[string]interface{}{
		"error": err,
	})
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var order entities.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		h.logger.Warn("failed to decode order request",
			"error", err,
		)
		h.writeError(w, http.StatusBadRequest, httperrors.NewHTTPError(
			httperrors.ErrCodeInvalidJSON,
			"Invalid JSON format",
			err.Error(),
		))
		return
	}

	result, err := h.svc.SaveOrder(ctx, order)
	if err != nil {
		h.handleServiceError(w, err, "failed to save order")
		return
	}

	h.logger.Info("order saved successfully",
		"order_id", order.OrderUID,
		"result", string(result),
	)

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"order_id": order.OrderUID,
		"result":   string(result),
		"status":   "success",
	})
}

func (h *OrderHandler) handleServiceError(w http.ResponseWriter, err error, context string) {
	var appErr *application.AppError

	switch {
	case errors.As(err, &appErr):
		h.logger.Error(context,
			"error", err,
			"error_code", appErr.Code,
			"operation", appErr.Op,
		)
		h.writeError(w, http.StatusInternalServerError, httperrors.NewHTTPError(
			httperrors.ErrCodeInternalError,
			"Internal server error",
			"",
		))

	case errors.Is(err, domain.ErrInvalidOrder):
		h.logger.Warn("invalid order data",
			"error", err,
		)
		h.writeError(w, http.StatusBadRequest, httperrors.NewHTTPError(
			httperrors.ErrCodeInvalidRequest,
			"Invalid order data",
			err.Error(),
		))

	case errors.Is(err, domain.ErrOrderNotFound):
		h.logger.Warn("order not found",
			"error", err,
		)
		h.writeError(w, http.StatusNotFound, httperrors.NewHTTPError(
			httperrors.ErrCodeOrderNotFound,
			"Order not found",
			"",
		))

	default:
		h.logger.Error("unexpected error",
			"error", err,
			"context", context,
		)
		h.writeError(w, http.StatusInternalServerError, httperrors.NewHTTPError(
			httperrors.ErrCodeInternalError,
			"Internal server error",
			"",
		))
	}
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		h.writeError(w, http.StatusBadRequest, httperrors.NewHTTPError(
			httperrors.ErrCodeInvalidRequest,
			"Order ID is required",
			"",
		))
		return
	}

	order, err := h.svc.GetOrder(ctx, id)
	if err != nil {
		h.handleServiceError(w, err, "failed to get order")
		return
	}

	h.logger.Info("order retrieved successfully",
		"order_id", id,
	)
	h.writeJSON(w, http.StatusOK, order)
}

func (h *OrderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	orders, err := h.svc.GetAllOrders(ctx)
	if err != nil {
		h.handleServiceError(w, err, "failed to get all orders")
		return
	}

	h.logger.Info("orders retrieved successfully",
		"count", len(orders),
	)
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"orders": orders,
		"count":  len(orders),
	})
}

func (h *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")

	if id == "" {
		h.writeError(w, http.StatusBadRequest, httperrors.NewHTTPError(
			httperrors.ErrCodeInvalidRequest,
			"Order ID is required",
			"",
		))
		return
	}

	err := h.svc.DeleteOrder(ctx, id)
	if err != nil {
		h.handleServiceError(w, err, "failed to delete order")
		return
	}

	h.logger.Info("order deleted successfully",
		"order_id", id,
	)
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status":   "deleted",
		"order_id": id,
	})
}

func (h *OrderHandler) Clear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	err := h.svc.ClearOrders(ctx)
	if err != nil {
		h.handleServiceError(w, err, "failed to clear orders")
		return
	}

	h.logger.Info("all orders cleared")
	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"status": "cleared",
		"count":  0,
	})
}
