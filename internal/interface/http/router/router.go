package router

import (
    "net/http"
    "github.com/go-chi/chi/v5"
    "github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/handler"
)

func New(h *handler.OrderHandler) *chi.Mux {
    r := chi.NewRouter()
    r.Post("/orders", h.Create)
    r.Get("/orders/{id}", h.GetByID)
    r.Get("/orders", h.GetAll)
    r.Delete("/orders/{id}", h.Delete)
    r.Delete("/orders", h.Clear)

    fs := http.FileServer(http.Dir("./web"))
    r.Handle("/*", fs)
    return r
}
