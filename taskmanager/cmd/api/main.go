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

	repo := memory.NewTaskRepostory()
	svc := service.NewTaskService(repo)
	taskHandler := handler.NewTaskHandler(svc)
	router := handler.NewTaskRouter(taskHandler)

	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}


