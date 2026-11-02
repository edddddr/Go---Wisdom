package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/edddddr/taskmanager/internal/domain"
)


var uniqueViolation = "23505"


type UserRepository struct{
	pool *pgxpool.Pool
} 

func NewUserRepository(pool *pgxpool.Pool) *UserRepository{
	return &UserRepository{pool:pool}
}


func (r *UserRepository) Create(ctx context.Context, u domain.User) error{
	
	err := r.pool.QueryRow(ctx,
	`INSERT TO users (email, password_hash)
	VALUE($1,$2)
	RETURNING id, created_at`,
	u.Email, u.CreatedAt).Scan(&u.ID, &u.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation{
		return domain.ErrorConflict
	}

	if err != nil{
		return fmt.Errorf("insert user %w", err)
	}

	return nil
}

func (r *UserRepository) GetEmail(ctx context.Context, email string) (*domain.User, error){
	var u domain.User
    err := r.pool.QueryRow(ctx,
		`SELECT id, email, password_hash, created_at FROM users WHERE email = $1`, email).Scan(u.ID, u.Email, u.PasswordHash, u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows){
		return nil, domain.ErrNotFound
	}

	if err != nil{
		return nil, fmt.Errorf("get user by email : ",err)
	}
		

	return &u, nil
}