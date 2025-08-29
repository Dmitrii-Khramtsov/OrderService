package bootstrap

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

func (a *App) Run() {
	a.Logger.Info("server starting", zap.String("addr", a.Server.Addr))
	a.KafkaConsumer.Start()

	go a.restoreCacheFromDB(context.Background())

	go func() {
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Error("server listen failed", zap.Error(err))
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) {
	a.Logger.Info("shutdown initiated")

	if err := a.KafkaConsumer.Shutdown(ctx); err != nil {
		a.Logger.Error("failed to shutdown kafka consumer", zap.Error(err))
	}

	if err := a.Server.Shutdown(ctx); err != nil {
		a.Logger.Error("server forced to shutdown", zap.Error(err))
	}

	resources := []struct {
		name string
		res  Shutdownable
	}{
		{"cache", a.Cache},
		{"logger", a.Logger},
		{"repository", a.Repo},
		{"database", a.DB},
	}

	for _, resource := range resources {
		if resource.res == nil {
			continue
		}
		if err := resource.res.Shutdown(ctx); err != nil {
			a.Logger.Error("failed to shutdown resource",
				zap.String("resource", resource.name),
				zap.Error(err),
			)
		} else {
			a.Logger.Info("resource stopped gracefully",
				zap.String("resource", resource.name),
			)
		}
	}

	a.Logger.Info("shutdown completed")
}

// восстановление кеша из БД
func (a *App) restoreCacheFromDB(ctx context.Context) {
	orders, err := a.Service.GetAllFromDB(ctx)
	if err != nil {
		a.Logger.Error("failed to restore cache from DB", zap.Error(err))
		return
	}

	for _, o := range orders {
		a.Cache.Set(o.OrderUID, o)
	}

	a.Logger.Info("cache restored from DB", zap.Int("count", len(orders)))
}
