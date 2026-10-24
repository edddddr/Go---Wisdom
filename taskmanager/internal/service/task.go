package service

import (
	"context"
	"fmt"
	"time"

	"github.com/edddddr/taskmanager/internal/domain"
)

const defaultPriority = 3


type TaskRepository interface {
	 Create(ctx context.Context, t *domain.Task) error
}


type CreateInput struct{
	Title       string
	Description string
	Priority *int 
}


type TaskService struct{
	repo TaskRepository
}

func NewTasService(repo TaskRepository) *TaskService {
	return &TaskService{repo:repo}
} 

func (r *TaskService) Create(ctx context.Context, in CreateInput) (*domain.Task, error){
	priority := defaultPriority

	if in.Priority != nil {
		priority = *in.Priority
	}

	t := &domain.Task{
		Title: in.Title,
		Description: in.Description,
		Priority: priority,
	    CreatedAt: time.Now(),
	}

	if err := r.repo.Create(ctx, t); err != nil{
		return nil, fmt.Errorf("create task: %w", err)
	}

	return t, nil
}

