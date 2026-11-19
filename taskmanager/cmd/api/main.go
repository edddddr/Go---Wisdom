package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edddddr/taskmanager/internal/config"
	"github.com/edddddr/taskmanager/internal/handler"
	"github.com/edddddr/taskmanager/internal/logging"

	// "github.com/edddddr/taskmanager/internal/repository/memory"
	"github.com/edddddr/taskmanager/internal/repository/postgres"
	"github.com/edddddr/taskmanager/internal/service"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	if err := run(); err != nil{
		log.Fatal("fata: %v", err)
	}
}


func run() error{

	cfg, err := config.Load()
	if err != nil{
		return fmt.Errorf("load confing %w", err)
	}


	var level slog.Level
	if err := level.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		return fmt.Errorf("parse LOG_LEVEL %q: %w", cfg.LogLevel, err)
	}

	production := cfg.Env == "production"
	slog.SetDefault(logging.New(os.Stdout, level, production))
	if production {
		gin.SetMode(gin.ReleaseMode)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	startCtx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	pool, err := pgxpool.New(startCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create pool %w", err)
	}

	defer pool.Close()

	if err := pool.Ping(startCtx); err != nil{
		return fmt.Errorf("ping database %w", err)
	}

	 
	authSvc, err := service.NewAuthService(postgres.NewUserRepository(pool), []byte(cfg.JWTSecret), time.Hour)
	if err !=  nil{
		return fmt.Errorf("create auth service: %w", err)
	}

	taskSvc := service.NewTaskService(postgres.NewTaskRepository(pool))

	
	router, err := handler.NewRouter(handler.Options{
		Auth:           handler.NewAuthHandler(authSvc),
		Tasks:          handler.NewTaskHandler(taskSvc),
		Tokens:         authSvc,
		DB:             pool,
		GlobalLimiter:  handler.NewIPLimiter(ctx, 20, 40),  // 20 req/s sustained, bursts of 40
		AuthLimiter:    handler.NewIPLimiter(ctx, 0.2, 5),  // 1 per 5s sustained, bursts of 5
		RequestTimeout: cfg.RequestTimeout,
		MaxBodyBytes:   1 << 20, // 1 MiB
		TrustedProxies: cfg.TrustedProxies,
	})
	if err != nil {
		return fmt.Errorf("create router: %w", err)
	}


	


	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second, 
		ReadTimeout:       10 * time.Second, 
		WriteTimeout:      15 * time.Second, 
		IdleTimeout:       60 * time.Second, 
	}


	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", promhttp.Handler())
	metricsSrv := &http.Server{
		Addr:              ":" + cfg.MetricsPort,
		Handler:           metricsMux,
		ReadHeaderTimeout: 5 * time.Second,
	}


	errCh := make(chan error, 2)
	go serve("api", srv, errCh)
	go serve("metrics", metricsSrv, errCh)


	select {
	case err := <-errCh:
		return fmt.Errorf("server failed: %w", err)
	case <-ctx.Done():
		slog.Info("shutdown signal received, draining requests")
	}

	shutdownCtx, cancelShutdown := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)

	defer cancelShutdown()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		_ = srv.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	
	slog.Info("server stopped cleanly")

	return nil
}


func serve(name string, srv *http.Server, errCh chan<- error) {
	slog.Info("server starting", "name", name, "addr", srv.Addr)
	if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		errCh <- fmt.Errorf("%s: %w", name, err)
	}
}