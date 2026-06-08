package entity

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, id int) (*Task, error)
	GetByUser(ctx context.Context, userID int, filter TaskFilter) ([]Task, error)
	Update(ctx context.Context, userID int, taskID int, input UpdateTaskInput) error
	Delete(ctx context.Context, userID int, id int) error
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByID(ctx context.Context, id int) (*User, error)
}

type NoteRepository interface {
	Create(ctx context.Context, note *Note) error
	GetByTaskID(ctx context.Context, taskID int) ([]*Note, error)
	GetByID(ctx context.Context, noteID bson.ObjectID) (*Note, error)
	Delete(ctx context.Context, noteID bson.ObjectID) error
}
