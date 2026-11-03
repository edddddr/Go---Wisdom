package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edddddr/taskmanager/internal/domain"
)


const taskColumns = `id, owner_id, title, description, done, priority, created_at, updated_at`

type TaskRepository struct {
	pool *pgxpool.Pool
}		


func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository{
	return &TaskRepository{pool:pool}
} 


func scanTasks(row pgx.Row) (domain.Task, error){
	var t  domain.Task
	err := row.Scan(&t.ID, &t.OwnerID, &t.Title, &t.Description, &t.Done, &t.Priority, &t.CreatedAt, &t.Updated_at)
	
	return t, err
}

func (r *TaskRepository) Create(ctx context.Context, t *domain.Task) error{
	err :=r.pool.QueryRow(ctx, 
		`INSERT INTO tasks (owner_id, title, description, priority, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		t.owner_id, t.Title, t.Description, t.Priority, t.CreatedAt,).Scan(&t.ID)
		
	if err != nil{	
		return fmt.Errorf("insert task: %w", err)
	} 

	return nil
}

func (r *TaskRepository) List(ctx context.Context, ownerID int, f domain.TaskFilter) ([]domain.Task, error){

	rows, err := r.pool.Query(ctx,
	`SELECT `+taskColumns+`
	FROM tasks
	WHERE owner_id = $1  AND ($2::boolean is NULL OR done = $2)
	ORDER BY id DESC`,
	ownerID, f.Done)

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


func (r *TaskRepository) Get(ctx context.Context, ownerID,  id int) (*domain.Task, error){
	t, err := scanTasks(r.pool.QueryRow(ctx,
	`SELECT `+taskColumns+` FROM tasks WHERE owner_id = $1 AND  id = $2  `, ownerID, id))

	if errors.Is(err, pgx.ErrNoRows){
		return nil, domain.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get task: %w", err)
	}
	
	return &t, nil
}


func (r *TaskRepository) Update(ctx context.Context, ownerID int, id int, p domain.TaskPatch) (*domain.Task, error){
	t, err := scanTasks(r.pool.QueryRow(ctx, 
	`UPDATE tasks SET
	title = COALESCE($3, title),
	description = COALESCE($4, description),
	done = COALESCE($5, done),
	priority = COALESCE($6, priority),
	updated_at = now(),
	WHERE id = $1 AND owner_id = $2 RETURNING `+taskColumns, id,ownerID, p.Title, p.Description, p.Done, p.Priority))

	if errors.Is(err, pgx.ErrNoRows){
		return nil, domain.ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("update task: %w", err)
	}
	return &t, nil
}

func (r *TaskRepository) Delete(ctx context.Context, ownerID int, id int) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1 AND owner_id = $2`, id, ownerID)
	if err != nil {
		return fmt.Errorf("delete task: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}