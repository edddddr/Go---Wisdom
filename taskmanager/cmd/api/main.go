package main

import (
	"log"

	"github.com/edddddr/taskmanager/internal/config"
	"github.com/edddddr/taskmanager/internal/handler"
	"github.com/edddddr/taskmanager/internal/repository/memory"
	"github.com/edddddr/taskmanager/internal/service"
)

func main() {
	cfg := config.Load()

	repo := memory.NewTasRepostory()
	svc := service.NewTasService(repo)
	taskHandler := handler.NewTaskHandler(svc)
	router := handler.NewTaskRouter(taskHandler)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}


