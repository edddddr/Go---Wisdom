package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edddddr/taskmanager/internal/config"
	"github.com/edddddr/taskmanager/internal/handler"

	// "github.com/edddddr/taskmanager/internal/repository/memory"
	"github.com/edddddr/taskmanager/internal/repository/postgres"
	"github.com/edddddr/taskmanager/internal/service"
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

	ctx, cancle := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancle()

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create pool %w", err)
	}

	defer pool.Close()

	if err := pool.Ping(ctx); err != nil{
		return fmt.Errorf("ping database %w", err)
	}


	userRepo := postgres.NewUserRepository(pool)
	authSvc, err := service.NewAuthService(userRepo, []byte(cfg.JWTSecret), time.Hour)

	if err !=  nil{
		return fmt.Errorf("create auth service: %w", err)
	}

	authHandler := handler.NewAuthHandler(authSvc)
	taskRepo := postgres.NewTaskRepository(pool)
	svc := service.NewTaskService(taskRepo)
	taskHandler := handler.NewTaskHandler(svc)
	router := handler.NewTaskRouter(authHandler, taskHandler, authSvc)

	return router.Run(":" + cfg.Port)
}