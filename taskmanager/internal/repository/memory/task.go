package memory

import (
	// "context"
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


func (r TaskRepository) create(ctx context.Context, t *domain.Task) error {

}