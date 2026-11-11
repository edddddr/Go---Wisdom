package domain

import (
	"errors"
	"time"
)

var   (
	ErrNotFound = errors.New("not found")
	ErrInvalidInput = errors.New("invalid input")
	ErrConflict = errors.New("conflict")
	ErrUnauthorized = errors.New("nauthorized")
)


type Task struct {
	ID          int       `json:"id"`
	OwnerID     int       `json:"owner_id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
	Updated_at  time.Time `json:"updated_at"`
}


type TaskPatch struct {
	Title *string
	Description *string
	Done *bool
	Priority *int
}

// TaskFilter: nil Done means "don't filter by done".
type TaskFilter struct {
	Done *bool
}

