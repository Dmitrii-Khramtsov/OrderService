package main

import (
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
	"github.com/joho/godotenv"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/handler"
)

func main() {
	// пока не через докер
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: failed to load .env file: %v", err)
	}

	mode := logger.DEV
	if env := os.Getenv("LOG_MODE"); env == "production" {
		mode = logger.PROD
	}

	l, err := logger.NewLogger(mode)
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}

	c := cache.NewOrderCache()
	svc := application.NewOrderService(c, l)
	h := handler.NewOrderHandler(svc, l)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Post("/order", h.Create)
	r.Get("/order/{id}", h.GetByID)
	r.Get("/order", h.GetAll)
	r.Delete("/order/{id}", h.Delete)
	r.Delete("/orders", h.Clear)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	l.Info("server starting", zap.String("port", port))
	if err := http.ListenAndServe(":"+port, r); err != nil {
		l.Error("server failed", zap.Error(err))
	}
}
