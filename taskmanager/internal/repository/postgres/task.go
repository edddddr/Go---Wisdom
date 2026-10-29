package postgres

import (
	"context"
	"fmt"

	"github.com/edddddr/taskmanager/internal/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	pool *pgxpool.Pool
}		


func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository{
	return &TaskRepository{pool:pool}
} 


func (r *TaskRepository) Create(ctx context.Context, t *domain.Task) error{
	err :=r.pool.QueryRow(ctx, 
		`INSERT INTO tasks (title, description, priority, created_at)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id`,
		t.Title, t.Description, t.CreatedAt, t.CreatedAt,).Scan(&t.ID)
		
	if err != nil{
		return fmt.Errorf("insert task: %w", err)
	} 

	return nil
}