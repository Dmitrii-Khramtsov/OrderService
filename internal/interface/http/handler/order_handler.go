// github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/order_handler.go
package handler

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/domain/entities"
	"github.com/go-chi/chi/v5"
)

type OrderHandler struct {
	svc application.OrderServiceInterface
}

func NewOrderHandler(s application.OrderServiceInterface) *OrderHandler {
	return &OrderHandler{
		svc: s,
	}
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	resp, exist := h.svc.GetOrder(id)
	if !exist {
		WriteJSON(w, http.StatusNotFound, Operation{
			Operation: "get",
			Status:    false,
			Message:   "order not found",
		})

		fmt.Printf("Заказа с таким ID не существует: %v\n", id)
		return
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
		fmt.Println("Анкодинг JSON не сработал")
		return
	}
}

func (h *OrderHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	orders := h.svc.GetAllOrder()

	if len(orders) == 0 {
		if err := json.NewEncoder(w).Encode([]entities.Order{}); err != nil {
			http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
			fmt.Println("Анкодинг JSON не сработал")
			return
		}
		return
	}

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "failed ti encod JSON", http.StatusInternalServerError)
		fmt.Println("Анкодинг JSON не сработал")
	}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req entities.Order
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		fmt.Println("Декодирование JSON не сработало")
		return
	}
	defer r.Body.Close()

	if req.OrderUID == "" {
		WriteJSON(w, http.StatusBadRequest, Operation{
			Operation: "post",
			Status:    false,
			Message:   "order_id is required",
		})
		fmt.Printf("order_id не может быть пустым: %v\n", req.OrderUID)
		return
	}

	result := h.svc.SaveOrder(req)

	statusHTTP := http.StatusInternalServerError
	message := "unexpected status"

	switch result {
	case application.OrderCreated:
		statusHTTP = http.StatusCreated
		message = "order created"
	case application.OrderUpdated:
		statusHTTP = http.StatusOK
		message = "order updated"
	case application.OrderExists:
		statusHTTP = http.StatusNotModified
		message = "order already exists"
	}

	WriteJSON(w, statusHTTP, Operation{
		Operation: "post",
		Status:    true,
		Message:   message,
	})

	b, _ := json.MarshalIndent(req, "", "  ")
	fmt.Println("Результат операции:", message, string(b))
}

func (h *OrderHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	id := chi.URLParam(r, "id")

	del := h.svc.DelOrder(id)
	if !del {
		WriteJSON(w, http.StatusNotFound, Operation{
			Operation: "delete",
			Status:    false,
			Message:   "order not found",
		})

		fmt.Printf("Заказа с таким ID не существует: %v\n", id)
		return
	}

	WriteJSON(w, http.StatusOK, Operation{
		Operation: "delete",
		Status:    true,
		Message:   "successfully deleted",
	})
}

func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode JSON", http.StatusInternalServerError)
		fmt.Println("Анкодинг JSON не сработал")
	}
}

type Operation struct {
	Operation string `json:"operation"`
	Status    bool   `json:"status"`
	Message   string `json:"message"`
}
