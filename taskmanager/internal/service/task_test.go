package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/edddddr/taskmanager/internal/repository/memory"
	"github.com/edddddr/taskmanager/internal/service"
)

func ptr[T any](v T) *T { return &v }

func newService() *service.TaskService {
	return service.NewTaskService(memory.NewTaskRepostory())
}
