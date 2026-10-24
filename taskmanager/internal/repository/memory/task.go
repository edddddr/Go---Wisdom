package memory

import (
	"context"
	// "sort"
	"context"
	"sync"

	"github.com/edddddr/taskmanager/internal/domain"
)

type TaskRepository struct{
	mu sync.RWMutex
	nextID int
	tasks map[int]domain.Task
}


func NewTasRepostory() *TaskRepository{
	return &TaskRepository{tasks: make(map[int]domain.Task)}
}


func (r *TaskRepository) create(ctx context.Context, t *domain.Task) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.nextID++
	t.ID = r.nextID
	r.tasks[t.ID] = *t 

	return nil
}
