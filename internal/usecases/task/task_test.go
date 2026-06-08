package task

import (
	"context"
	"testing"
	"time"
	"todoapp/internal/entity"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type mockTaskRepo struct {
	task *entity.Task
	err  error
}

func (m *mockTaskRepo) Create(ctx context.Context, task *entity.Task) error { return m.err }
func (m *mockTaskRepo) GetByID(ctx context.Context, id int) (*entity.Task, error) {
	return m.task, m.err
}
func (m *mockTaskRepo) GetByUser(ctx context.Context, userID int, filter entity.TaskFilter) ([]entity.Task, error) {
	return nil, m.err
}
func (m *mockTaskRepo) Update(ctx context.Context, userID, taskID int, input entity.UpdateTaskInput) error {
	return m.err
}
func (m *mockTaskRepo) Delete(ctx context.Context, userID, id int) error { return m.err }

type mockNoteRepo struct{}

func (m *mockNoteRepo) Create(ctx context.Context, note *entity.Note) error { return nil }
func (m *mockNoteRepo) GetByTaskID(ctx context.Context, taskID int) ([]*entity.Note, error) {
	return nil, nil
}
func (m *mockNoteRepo) GetByID(ctx context.Context, noteID bson.ObjectID) (*entity.Note, error) {
	return nil, nil
}
func (m *mockNoteRepo) Delete(ctx context.Context, noteID bson.ObjectID) error { return nil }

func newTestService(taskRepo entity.TaskRepository) *Service {
	return NewService(taskRepo, &mockNoteRepo{}, nil, nil)
}

func TestUpdate_InvalidStatus(t *testing.T) {
	svc := newTestService(&mockTaskRepo{})

	invalidStatus := "invalid_status"
	err := svc.Update(context.Background(), 1, 1, entity.UpdateTaskInput{
		Status: &invalidStatus,
	})

	if err == nil {
		t.Fatal("expected error for invalid status, got nil")
	}
	if err != entity.ErrInvalidStatus {
		t.Fatalf("expected ErrInvalidStatus, got %v", err)
	}
}

func TestGetByID_AccessDenied(t *testing.T) {
	svc := newTestService(&mockTaskRepo{
		task: &entity.Task{
			ID:        1,
			UserID:    1,
			Title:     "secret task",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	})

	_, err := svc.GetByID(context.Background(), 2, 1)

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err != entity.ErrNotFoundOrAccessDenied {
		t.Fatalf("expected ErrNotFoundOrAccessDenied, got %v", err)
	}
}
