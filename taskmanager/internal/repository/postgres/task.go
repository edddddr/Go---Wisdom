package postgres

import (
	"github.com/jackc/pgx/v5/pgxpool"

)
				
type TaskRepository struct {
	pool *pgxpool.Pool
}		