package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	// "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

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

	c := cache.NewOrderCache(l)
	svc := application.NewOrderService(c, l)
	h := handler.NewOrderHandler(svc, l)

	r := chi.NewRouter()
	// r.Use(middleware.Logger)
	// r.Use(middleware.Recoverer)

	r.Post("/orders", h.Create)
	r.Get("/orders/{id}", h.GetByID)
	r.Get("/orders", h.GetAll)
	r.Delete("/orders/{id}", h.Delete)
	r.Delete("/orders", h.Clear)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	l.Info("server starting", zap.String("port", port))

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			l.Error("listen failed", zap.Error(err))
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	l.Info("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		l.Error("server forced to shutdown", zap.Error(err))
	} else {
		l.Info("server stopped gracefully")
	}

	type Shutdownable interface {
		Shutdown(ctx context.Context) error
	}
	shutdownResources := []Shutdownable{c, l} // +db

	for _, r := range shutdownResources {
		if err := r.Shutdown(ctx); err != nil {
			l.Error("resource shutdown failed", zap.Error(err))
		}
	}

	l.Info("Server stopped")
}
