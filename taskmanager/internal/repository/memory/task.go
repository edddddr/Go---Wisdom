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


func NewTaskRepostory() *TaskRepository{
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


func (r *TaskRepository) List(ctx context.Context, ownerID int, f domain.TaskFilter) ([]domain.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]domain.Task, 0, len(r.tasks))
	for _, t := range r.tasks {
		if t.OwnerID != ownerID {
			continue
		}
		if f.Done != nil && t.Done != *f.Done {
			continue
		}
		if f.BeforeID != nil && t.ID >= *f.BeforeID {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	if f.Limit > 0 && len(out) > f.Limit {
		out = out[:f.Limit]
	}
	return out, nil
}


func (r *TaskRepository)  Get(ctx context.Context, id int) (*domain.Task, error){
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tasks[id]
	if !ok{
		return nil, domain.ErrNotFound
	}

	return &t, nil
}	


func (r *TaskRepository) Update(ctx context.Context, id int, p domain.TaskPatch) (*domain.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tasks[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	if p.Title != nil {
		t.Title = *p.Title
	}
	if p.Description != nil {
		t.Description = *p.Description
	}
	if p.Done != nil {
		t.Done = *p.Done
	}
	if p.Priority != nil {
		t.Priority = *p.Priority
	}
	r.tasks[id] = t
	return &t, nil
}

// func (r *TaskRepository) Delete(ctx context.Context, id int) error {
// 	r.mu.Lock()
// 	defer r.mu.Unlock()

// 	_, ok := r.tasks[id]

// 	if !ok{
// 		return domain.ErrNotFound
// 	}

// 	delete(r.tasks, id)

// 	return nil
// }


func (r *TaskRepository) Delete(ctx context.Context, userID, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	t, ok := r.tasks[id]
	if !ok || t.OwnerID != userID {
		return domain.ErrNotFound
	}
	delete(r.tasks, id)
	return nil
}