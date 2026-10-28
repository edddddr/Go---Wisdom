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

func TestTaskService_Create_Priority(t *testing.T) {
	tests := []struct {
		name  string
		input *int
		want  int
	}{
		{"defaults to 3 when omitted", nil, 3},
		{"uses provided priority", ptr(5), 5},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			svc := newService()

			task, err := svc.Create(context.Background(), service.CreateInput{
				Title:    "write tests",
				Priority: tc.input,
			})
			if err != nil {
				t.Fatalf("Create() unexpected error: %v", err)
			}
			if task.Priority != tc.want {
				t.Errorf("Priority = %d, want %d", task.Priority, tc.want)
			}
		})
	}
}

func TestTaskService_Create_BlankTitle(t *testing.T) {
	svc := newService()

	_, err := svc.Create(context.Background(), service.CreateInput{Title: "   "})

	if !errors.Is(err, domain.ErrInvalidInput) {
		t.Errorf("err = %v, want ErrInvalidInput", err)
	}
}

