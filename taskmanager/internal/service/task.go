package service

import (
	"context"
	"fmt"
	"time"
	"strings"

	"github.com/edddddr/taskmanager/internal/domain"
)

const defaultPriority = 3


type TaskRepository interface {
	 Create(ctx context.Context, t *domain.Task) error
	 List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, error)
	 Get(ctx context.Context, id int) (*domain.Task, error)
	 Update(ctx context.Context, id int, p domain.TaskPatch) (*domain.Task, error)
	 Delete(ctx context.Context, id int) error
}


type CreateInput struct{
	Title       string
	Description string
	Priority *int 
}
type UpdateInput struct{
	Title       *string
	Description *string
	Priority    *int 
	Done        *bool
}


type TaskService struct{
	repo TaskRepository
}

func NewTaskService(repo TaskRepository) *TaskService {
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


func (r *TaskService) List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, error){
	tasks, err := r.repo.List(ctx, f)
	if err != nil {
		return nil, fmt.Errorf("list task: %w", err)
	} 

	return tasks, nil
} 


func (s *TaskService) Get(ctx context.Context, id int) (*domain.Task, error){
		task, err := s.repo.Get(ctx, id)

		if err != nil {
			return nil, fmt.Errorf("get taks: %d %w", id, err)
		}

		return task, nil
}	


func (s *TaskService) Update(ctx context.Context, id int, in UpdateInput) (*domain.Task, error){
	patch := domain.TaskPatch{
		Description: in.Description,
		Done: in.Done,
		Priority: in.Priority,
	}

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return nil, fmt.Errorf("title must not be blank: %w", domain.ErrInvalidInput)
		}
		patch.Title = &title
	}
	
	
	
	t, err := s.repo.Update(ctx, id, patch)
	if err != nil{
		return nil, fmt.Errorf("update task %d %w", id, err)
	}

	return t, nil

}

func (s *TaskService) Delete(ctx context.Context, id int) error {
	err := s.repo.Delete(ctx, id)

	if err != nil{
		return fmt.Errorf("delete task %d %w", id, err)
	}

	return nil
}