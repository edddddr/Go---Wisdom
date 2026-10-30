package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edddddr/taskmanager/internal/domain"
)


const taskColumns = `title, description, done, priority, created_at`

type TaskRepository struct {
	pool *pgxpool.Pool
}		


func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository{
	return &TaskRepository{pool:pool}
} 


func scanTasks(row pgx.Row) (domain.Task, error){
	var t  domain.Task
	err := row.Scan(&t.ID, &t.Description, &t.Done, &t.Priority, &t.CreatedAt)
	
	return t, err
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

func (r *TaskRepository) List(ctx context.Context, f *domain.TaskFilter) ([]domain.Task, error){

	rows, err := r.pool.Query(ctx,
	`SELECT `+taskColumns+`
	FROM tasks
	WHERE ($1::boolean is NULL OR data = $1)
	ORDER BY ID DESC`,
	f.Done)

	if err != nil {
		return nil, fmt.Errorf("query tasks %w", err)
	}

	defer rows.Close()

	tasks := make([]domain.Task, 0)

	for rows.Next(){
		task, err := scanTasks(rows)

		if err != nil {
			return nil, fmt.Errorf("scan task : %w", err)
		}

		tasks = append(tasks, task)
	}

	if err := rows.Err(); err != nil{
		return nil, fmt.Errorf("iterate task %w", err)
	}

	return tasks, nil
}


func (r *TaskRepository) Get(ctx context.Context, id int) (*domain.Task, error){
	t, err := scanTasks(r.pool.QueryRow(ctx,
	`SELECT `+taskColumns+` FROM tasks WHERE id = $1  `, id))

	if errors.Is(err, pgx.ErrNoRows){
		return nil, domain.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	
	return &t, nil
}