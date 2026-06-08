package entity

import (
	"time"
)

type Task struct {
	ID          int       `json:"id" db:"id"`
	UserID      int       `json:"user_id" db:"user_id"`
	Title       string    `json:"title" db:"title"`
	Description string    `json:"description" db:"description"`
	Notes       []*Note   `json:"notes,omitempty" db:"-"`
	Status      string    `json:"status" db:"status"`
	Deadline    time.Time `json:"deadline" db:"deadline"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
type CreateTaskInput struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Deadline    time.Time `json:"deadline"`
}

type UpdateTaskInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *string    `json:"status"`
	Deadline    *time.Time `json:"deadline"`
}

type TaskFilter struct {
	Status *string
	Limit  int
	Offset int
}
