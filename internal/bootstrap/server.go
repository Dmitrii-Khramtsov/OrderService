// // github.com/Dmitrii-Khramtsov/orderservice/internal/bootstrap/server.go
package bootstrap

// import (
// 	"context"
// 	"fmt"
// 	"net/http"
// 	"os"
// 	"time"

// 	"github.com/cenkalti/backoff/v4"
// 	"github.com/go-chi/chi/v5"
// 	"go.uber.org/zap"
// 	"github.com/Dmitrii-Khramtsov/orderservice/internal/application"
// 	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/cache"
// 	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/config"
// 	repo "github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/database"
// 	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/kafka"
// 	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/logger"
// 	"github.com/Dmitrii-Khramtsov/orderservice/internal/infrastructure/migrations"
// 	"github.com/Dmitrii-Khramtsov/orderservice/internal/interface/http/handler"
// 	"github.com/jmoiron/sqlx"
// )

// type Shutdownable interface {
// 	Shutdown(ctx context.Context) error
// }

// type DBWrapper struct {
// 	DB *sqlx.DB
// }

// func (dw *DBWrapper) Shutdown(ctx context.Context) error {
// 	return dw.DB.Close()
// }

// type App struct {
// 	Server        *http.Server
// 	Cache         cache.Cache
// 	Logger        logger.LoggerInterface
// 	Service       application.OrderServiceInterface
// 	Handler       *handler.OrderHandler
// 	Repo          repo.OrderRepository
// 	KafkaConsumer kafka.ConsumerInterface
// 	DB            Shutdownable
// }

// func NewApp() (*App, error) {
// 	cfg, err := config.LoadConfig("config.yml")
// 	if err != nil {
// 		return nil, fmt.Errorf("failed to load config: %w", err)
// 	}

// 	l, err := newLogger(cfg)
// 	if err != nil {
// 		return nil, err
// 	}

// 	c := newCache(l, cfg.Cache.Capacity)

// 	db, err := newDatabaseConnection(l, cfg)
// 	if err != nil {
// 		return nil, err
// 	}

// 	ctx := context.Background()
// 	if err := runMigrations(ctx, db, cfg.Migrations.MigrationsPath, l); err != nil {
// 		l.Error("failed to run migrations", zap.Error(err))
// 		return nil, err
// 	}

// 	rp, err := newPostgresOrderRepository(db, l)
// 	if err != nil {
// 		return nil, err
// 	}

// 	svc := newOrderService(c, l, rp, cfg.Cache.GetAllLimit)
// 	h := newOrderHandler(svc, l)

// 	kc := newKafkaConsumer(cfg.Kafka.Brokers, cfg.Kafka.Topic, cfg.Kafka.GroupID, svc, l)

// 	r := newRouter(h)
// 	srv := newHTTPServer(cfg.Server.Port, r)

// 	return &App{
// 		Server:        srv,
// 		Cache:         c,
// 		Logger:        l,
// 		Service:       svc,
// 		Handler:       h,
// 		Repo:          rp,
// 		KafkaConsumer: kc,
// 		DB:            &DBWrapper{DB: db},
// 	}, nil
// }

// func newLogger(cfg *config.Config) (logger.LoggerInterface, error) {
// 	mode := logger.DEV
// 	if envLogMode := os.Getenv("LOG_MODE"); envLogMode == "production" {
// 		mode = logger.PROD
// 	}
// 	return logger.NewLogger(mode)
// }

// func newCache(l logger.LoggerInterface, capacity int) cache.Cache {
// 	return cache.NewOrderLRUCache(l, capacity)
// }

// func newDatabaseConnection(l logger.LoggerInterface, cfg *config.Config) (*sqlx.DB, error) {
// 	db, err := sqlx.Connect("postgres", cfg.Database.DSN)
// 	if err != nil {
// 		l.Error("failed to connect to db", zap.Error(err))
// 		return nil, repo.ErrDatabaseConnectionFailed
// 	}

// 	db.SetMaxOpenConns(cfg.Database.MaxOpenConns)
// 	db.SetMaxIdleConns(cfg.Database.MaxIdleConns)
// 	db.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

// 	return db, nil
// }

// func runMigrations(ctx context.Context, db *sqlx.DB, migrationsPath string, l logger.LoggerInterface) error {
// 	operation := func() error {
// 		return migrations.RunMigrations(ctx, db.DB, migrationsPath, l)
// 	}

// 	retryPolicy := backoff.NewExponentialBackOff()
// 	retryPolicy.MaxElapsedTime = 30 * time.Second

// 	err := backoff.Retry(operation, retryPolicy)
// 	if err != nil {
// 		l.Error("failed to run migrations after retries", zap.Error(err))
// 		return err
// 	}

// 	l.Info("migrations completed successfully")
// 	return nil
// }

// func newPostgresOrderRepository(db *sqlx.DB, l logger.LoggerInterface) (repo.OrderRepository, error) {
// 	return repo.NewPostgresOrderRepository(db, l)
// }

// func newOrderService(c cache.Cache, l logger.LoggerInterface, rp repo.OrderRepository, limit int) application.OrderServiceInterface {
// 	return application.NewOrderService(c, l, rp, limit)
// }

// func newOrderHandler(svc application.OrderServiceInterface, l logger.LoggerInterface) *handler.OrderHandler {
// 	return handler.NewOrderHandler(svc, l)
// }

// func newKafkaConsumer(brokers []string, topic, groupID string, svc application.OrderServiceInterface, l logger.LoggerInterface) kafka.ConsumerInterface {
// 	kc := kafka.NewConsumer(brokers, topic, groupID, svc, l)
// 	return kc
// }

// func newRouter(h *handler.OrderHandler) *chi.Mux {
// 	r := chi.NewRouter()
// 	r.Post("/orders", h.Create)
// 	r.Get("/orders/{id}", h.GetByID)
// 	r.Get("/orders", h.GetAll)
// 	r.Delete("/orders/{id}", h.Delete)
// 	r.Delete("/orders", h.Clear)
// 	fs := http.FileServer(http.Dir("./web"))
// 	r.Handle("/*", fs)
// 	return r
// }

// func newHTTPServer(port string, r *chi.Mux) *http.Server {
// 	return &http.Server{
// 		Addr:    ":" + port,
// 		Handler: r,
// 	}
// }

// func (a *App) Run() {
// 	a.Logger.Info("server starting", zap.String("addr", a.Server.Addr))
// 	a.KafkaConsumer.Start()
// 	go a.RestoreCacheFromDB(context.Background())
// 	go func() {
// 		if err := a.Server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
// 			a.Logger.Error("server listen failed", zap.Error(err))
// 		}
// 	}()
// }

// func (a *App) RestoreCacheFromDB(ctx context.Context) {
// 	operation := func() error {
// 		orders, err := a.Service.GetAllFromDB(ctx)
// 		if err != nil {
// 			a.Logger.Error("failed to restore cache from DB", zap.Error(err))
// 			return err
// 		}
// 		for _, o := range orders {
// 			a.Cache.Set(o.OrderUID, o)
// 		}
// 		a.Logger.Info("cache restored from DB", zap.Int("count", len(orders)))
// 		return nil
// 	}

// 	retryPolicy := backoff.NewExponentialBackOff()
// 	retryPolicy.MaxElapsedTime = 30 * time.Second

// 	err := backoff.Retry(operation, retryPolicy)
// 	if err != nil {
// 		a.Logger.Error("failed to restore cache from DB after retries", zap.Error(err))
// 	}
// }

// func (a *App) Shutdown(ctx context.Context) {
// 	a.Logger.Info("shutdown initiated")

// 	if err := a.KafkaConsumer.Shutdown(ctx); err != nil {
// 		a.Logger.Error("failed to shutdown kafka consumer", zap.Error(err))
// 	} else {
// 		a.Logger.Info("kafka consumer stopped gracefully")
// 	}

// 	if err := a.Server.Shutdown(ctx); err != nil {
// 		a.Logger.Error("server forced to shutdown", zap.Error(err))
// 	} else {
// 		a.Logger.Info("server stopped gracefully")
// 	}

// 	resources := []struct {
// 		name string
// 		res  Shutdownable
// 	}{
// 		{"cache", a.Cache},
// 		{"logger", a.Logger},
// 		{"repository", a.Repo},
// 		{"database", a.DB},
// 	}
// 	for _, resource := range resources {
// 		if err := resource.res.Shutdown(ctx); err != nil {
// 			a.Logger.Error("failed to shutdown resource",
// 				zap.String("resource", resource.name),
// 				zap.Error(err),
// 			)
// 		} else {
// 			a.Logger.Info("resource stopped gracefully",
// 				zap.String("resource", resource.name),
// 			)
// 		}
// 	}
// 	a.Logger.Info("shutdown completed")
// }
