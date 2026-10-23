package service

import (
	"context"
	"fmt"
	"time"

	"github.com/edddddr/taskmanager/internal/domain"
)

const defaultPriority = 3

// TaskRepository is what the service NEEDS. Defined here, implemented elsewhere.
type TaskRepository interface {
	Create(ctx context.Context, t *domain.Task) error
	List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, error)	
}

type CreateTaskInput struct {
	Title       string
	Description string
	Priority    *int
}

type TaskService struct {
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo: repo}
}

func (s *TaskService) Create(ctx context.Context, in CreateTaskInput) (*domain.Task, error) {
	priority := defaultPriority
	if in.Priority != nil {
		priority = *in.Priority
	}

	t := &domain.Task{
		Title:       in.Title,
		Description: in.Description,
		Priority:    priority,
		CreatedAt:   time.Now().UTC(),
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, fmt.Errorf("create task: %w", err)
	}
	return t, nil
}


func (s *TaskService) List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, error) {
	tasks, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return tasks, nil
}