package main

import (
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/handler"
)

func main() {
	// 1. Инициализируем инфраструктуру
	c := cache.NewOrderCache()

	// 2. Инициализируем application-слой
	svc := application.NewOrderService(c)

	// 3. Инициализируем handler
	h := handler.NewOrderHandler(svc)

	// 4. Роутер
	r := chi.NewRouter()

	r.Post("/order", h.Create)
	r.Get("/order/{id}", h.GetByID)
	r.Get("/order", h.List)
	r.Delete("/order/{id}", h.Delete)

	// 5. Запуск сервера
	log.Println("server starting on :8081")
	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err)
	}
}
