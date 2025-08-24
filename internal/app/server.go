package app

import (
	"context"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
	"github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/handler"
)

type App struct {
	Server  *http.Server
	Cache   cache.Cache
	Logger  logger.LoggerInterface
	Service application.OrderServiceInterface
	Handler *handler.OrderHandler
}

func NewApp() (*App, error) {
	mode := logger.DEV
	if env := os.Getenv("LOG_MODE"); env == "production" {
		mode = logger.PROD
	}

	l, err := logger.NewLogger(mode)
	if err != nil {
		return nil, err
	}

	c := cache.NewOrderCache(l)
	svc := application.NewOrderService(c, l)
	h := handler.NewOrderHandler(svc, l)

	r := chi.NewRouter()
	r.Post("/orders", h.Create)
	r.Get("/orders/{id}", h.GetByID)
	r.Get("/orders", h.GetAll)
	r.Delete("/orders/{id}", h.Delete)
	r.Delete("/orders", h.Clear)

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8081"
	}

	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	return &App{
		Server:  srv,
		Cache:   c,
		Logger:  l,
		Service: svc,
		Handler: h,
	}, nil
}

func (a *App) Run() {
	a.Logger.Info("server starting", zap.String("addr", a.Server.Addr))

	go func() {
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Error("server listen failed", zap.Error(err))
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) {
	a.Logger.Info("shutdown initiated")

	if err := a.Server.Shutdown(ctx); err != nil {
		a.Logger.Error("server forced to shutdown", zap.Error(err))
	} else {
		a.Logger.Info("server stopped gracefully")
	}

	type Shutdownable interface {
		Shutdown(ctx context.Context) error
	}
	resources := []Shutdownable{a.Cache, a.Logger} // +db

	for _, r := range resources {
		if err := r.Shutdown(ctx); err != nil {
			a.Logger.Error("resource shutdown failed", zap.Error(err))
		}
	}

	a.Logger.Info("shutdown completed")
}
