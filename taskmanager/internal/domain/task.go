package domain

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("not found")
var ErrInvalidInput = errors.New("Invalid input")

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Done        bool      `json:"done"`
	Priority    int       `json:"priority"`
	CreatedAt   time.Time `json:"created_at"`
}

// TaskFilter: nil Done means "don't filter by done".
type TaskFilter struct {
	Done *bool
}