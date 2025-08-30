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

func writeJSON(w http.ResponseWriter, l domainrepo.Logger, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		l.Error("failed to encode JSON response", "error", err)
		http.Error(w, `{"error": "`+httperrors.ErrJSONEncodeFailed.Error()+`"}`, http.StatusInternalServerError)
	}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	order, err := h.decodeOrderRequest(r)
	if err != nil {
		h.handleDecodeError(w, err)
		return
	}

	result, err := h.svc.SaveOrder(ctx, order)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	h.logger.Info("order saved successfully", "order_id", order.OrderUID, "result", string(result))
	h.writeSuccessResponse(w, order.OrderUID, string(result))
}

func (h *OrderHandler) decodeOrderRequest(r *http.Request) (entities.Order, error) {
	var order entities.Order
	if err := json.NewDecoder(r.Body).Decode(&order); err != nil {
		return entities.Order{}, err
	}
	return order, nil
}

func (h *OrderHandler) handleDecodeError(w http.ResponseWriter, err error) {
	h.logger.Warn("failed to decode order request",
		"error", err,
		"error_type", httperrors.ErrInvalidJSON.Error(),
	)
	writeJSON(w, h.logger, http.StatusBadRequest, map[string]string{"error": httperrors.ErrInvalidJSON.Error()})
}

func (h *OrderHandler) handleServiceError(w http.ResponseWriter, err error) {
	if errors.Is(err, domain.ErrInvalidOrder) {
		h.logger.Warn("invalid order data", "error", err)
		writeJSON(w, h.logger, http.StatusBadRequest, map[string]string{"error": err.Error()})
	} else {
		h.logger.Error("failed to save order", "error", err)
		writeJSON(w, h.logger, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}
}

func (h *OrderHandler) writeSuccessResponse(w http.ResponseWriter, orderID, result string) {
	writeJSON(w, h.logger, http.StatusOK, map[string]string{
		"order_id": orderID,
		"result":   result,
	})
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	order, err := h.svc.GetOrder(ctx, id)
	if errors.Is(err, domain.ErrOrderNotFound) {
		writeJSON(w, h.logger, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	h.logger.Info("order retrieved successfully", "order_id", id)
	writeJSON(w, h.logger, http.StatusOK, order)
}

func (h *OrderHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	orders, err := h.svc.GetAllOrders(ctx)
	if err != nil {
		h.logger.Error("failed to get all orders", "error", err)
		writeJSON(w, h.logger, http.StatusInternalServerError, map[string]string{"error": "failed to get all orders"})
		return
	}
	h.logger.Info("orders retrieved successfully", "count", len(orders))
	writeJSON(w, h.logger, http.StatusOK, orders)
}

func (h *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	err := h.svc.DeleteOrder(ctx, id)
	if errors.Is(err, domain.ErrOrderNotFound) {
		writeJSON(w, h.logger, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}
	h.logger.Info("order deleted successfully", "order_id", id)
	writeJSON(w, h.logger, http.StatusOK, map[string]string{"status": "deleted", "order_id": id})
}

func (h *OrderHandler) Clear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if err := h.svc.ClearOrders(ctx); err != nil {
		h.logger.Error("failed to clear orders", "error", err)
		writeJSON(w, h.logger, http.StatusInternalServerError, map[string]string{"error": "failed to clear orders"})
		return
	}
	h.logger.Info("all orders cleared")
	writeJSON(w, h.logger, http.StatusOK, map[string]string{"status": "cleared"})
}
