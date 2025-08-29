// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/lifecycle.go
package bootstrap

import (
	"context"
	"net/http"

	"go.uber.org/zap"
)

func (a *App) Run() {
	a.Logger.Info("server starting", zap.String("addr", a.Server.Addr))
	a.KafkaConsumer.Start()

	go a.restoreCacheFromDB()

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

func (a *App) restoreCacheFromDB() {
	ctx := context.Background()
	
	if err := a.CacheRestorer.Restore(ctx); err != nil {
		a.Logger.Error("failed to restore cache from DB", zap.Error(err))
	} else {
		a.Logger.Info("cache restored successfully from DB")
	}
}
