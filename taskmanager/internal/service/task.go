package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/edddddr/taskmanager/internal/domain"
)


const (
	defaultPriority = 3
	defaultPageSize = 20
	maxPageSize     = 100
)



type TaskRepository interface {
	 Create(ctx context.Context, t *domain.Task) error
	 List(ctx context.Context, ownerID int, f domain.TaskFilter) ([]domain.Task, error)
	 Get(ctx context.Context,  ownerID int, id int) (*domain.Task, error)
	 Update(ctx context.Context, ownerID int, id int, p domain.TaskPatch) (*domain.Task, error)
	 Delete(ctx context.Context, ownerID int, id int) error
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


type ListTasksInput struct {
	Done     *bool
	Limit    *int
	BeforeID *int 
}

type TaskPage struct {
	Items      []domain.Task
	NextCursor *int
}


func NewTaskService(repo TaskRepository) *TaskService {
	return &TaskService{repo:repo}
} 

func (r *TaskService) Create(ctx context.Context, ownerID int, in CreateInput) (*domain.Task, error){
	priority := defaultPriority

	if in.Priority != nil {
		priority = *in.Priority
	}
	
	t := &domain.Task{
		Title: in.Title,
		Description: in.Description,
		Priority: priority,
		OwnerID: ownerID,
	    CreatedAt: time.Now(),
	}

	if err := r.repo.Create(ctx, t); err != nil{
		return nil, fmt.Errorf("create task: %w", err)
	}

	return t, nil
}


func (r *TaskService) List(ctx context.Context, ownerID int, in ListTasksInput) (*TaskPage, error){
	limit := defaultPageSize


	if in.Limit != nil {
		limit = *in.Limit
		if limit < 1 || limit > maxPageSize {
			return nil, &domain.InvalidInputError{
				Field:  "limit",
				Reason: fmt.Sprintf("must be between 1 and %d", maxPageSize),
			}
		}
	}

	if in.BeforeID != nil && *in.BeforeID < 1 {
		return nil, &domain.InvalidInputError{Field: "cursor", Reason: "invalid cursor"}
	}


	tasks, err := r.repo.List(ctx, ownerID, domain.TaskFilter{
		Done:     in.Done,
		BeforeID: in.BeforeID,
		Limit:    limit + 1,
	})


	if err != nil {
		return nil, fmt.Errorf("list task: %w", err)
	} 

	page := &TaskPage{Items: tasks}
	if len(tasks) > limit {
		page.Items = tasks[:limit]
		next := tasks[limit-1].ID
		page.NextCursor = &next
	}
	return page, nil
} 


func (s *TaskService) Get(ctx context.Context, ownerID int, id int) (*domain.Task, error){
		task, err := s.repo.Get(ctx, ownerID, id)

		if err != nil {
			return nil, fmt.Errorf("get taks: %d %w", id, err)
		}

		return task, nil
}	


func (s *TaskService) Update(ctx context.Context, ownerID int, id int, in UpdateInput) (*domain.Task, error){
	patch := domain.TaskPatch{
		Description: in.Description,
		Done: in.Done,
		Priority: in.Priority,
	}

	if in.Title != nil {
		title := strings.TrimSpace(*in.Title)
		if title == "" {
			return nil, &domain.InvalidInputError{Field: "title", Reason: "must not be blank"}
		}
		patch.Title = &title
	}
	
	
	
	t, err := s.repo.Update(ctx, ownerID, id, patch)
	if err != nil{
		return nil, fmt.Errorf("update task %d %w", id, err)
	}

	return t, nil

}

func (s *TaskService) Delete(ctx context.Context, ownerID int, id int) error {
	err := s.repo.Delete(ctx, ownerID, id)

	if err != nil{
		return fmt.Errorf("delete task %d %w", id, err)
	}

	return nil
}