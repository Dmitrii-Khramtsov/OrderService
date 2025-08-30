// github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/lifecycle.go
package bootstrap

import (
	"context"
	"net/http"
)

func (a *App) Run() {
	a.Logger.Info("server starting", "addr", a.Server.Addr)
	a.KafkaConsumer.Start()

	go a.restoreCacheFromDB()

	go func() {
		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.Logger.Error("server listen failed", "error", err)
		}
	}()
}

func (a *App) Shutdown(ctx context.Context) {
	a.Logger.Info("shutdown initiated")

	if err := a.KafkaConsumer.Shutdown(ctx); err != nil {
		a.Logger.Error("failed to shutdown kafka consumer", "error", err)
	}

	if err := a.Server.Shutdown(ctx); err != nil {
		a.Logger.Error("server forced to shutdown", "error", err)
	}

	resources := []struct {
		name string
		res  Shutdownable
	}{
		{"repository", a.Repo},
		{"database", a.DB},
	}

	for _, resource := range resources {
		if resource.res == nil {
			continue
		}
		if err := resource.res.Shutdown(ctx); err != nil {
			a.Logger.Error("failed to shutdown resource",
				"resource", resource.name,
				"error", err,
			)
		} else {
			a.Logger.Info("resource stopped gracefully",
				"resource", resource.name,
			)
		}
	}

	a.Logger.Info("shutdown completed")
}

func (a *App) restoreCacheFromDB() {
	ctx := context.Background()

	if err := a.CacheRestorer.Restore(ctx); err != nil {
		a.Logger.Error("failed to restore cache from DB", "error", err)
	} else {
		a.Logger.Info("cache restored successfully from DB")
	}
}
