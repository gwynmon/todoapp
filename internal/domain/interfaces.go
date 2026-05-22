package domain

import "context"

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id int) (*Task, error)
	GetByUser(ctx context.Context, userID int) ([]Task, error)
	Update(ctx context.Context, task *Task) error
	Delete(ctx context.Context, id int) error
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int) (*User, error)
}

type NoteRepository interface {
	Create(ctx context.Context, note *Note) error
	GetByTaskID(ctx context.Context, taskID int) ([]Note, error)
	Delete(ctx context.Context, id string) error
	DeleteByTaskID(ctx context.Context, taskID int) error
}
