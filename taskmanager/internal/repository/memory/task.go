package memory

import (
	"context"
	"sort"
	// "sort"
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


func (r *TaskRepository) Create(ctx context.Context, t *domain.Task) error {

	r.mu.Lock()
	defer r.mu.Unlock()
	
	r.nextID++
	t.ID = r.nextID
	r.tasks[t.ID] = *t 

	return nil
}

func (r *TaskRepository) List(ctx context.Context, f domain.TaskFilter) ([]domain.Task, error){
	r.mu.Lock()
	defer  r.mu.Unlock()

	out := make([]domain.Task, 0, len(r.tasks))

	for _, t := range r.tasks{
		if f.Done != nil && t.Done != *f.Done{
			continue
		}
		out = append(out, t)
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })

	return out, nil
}

