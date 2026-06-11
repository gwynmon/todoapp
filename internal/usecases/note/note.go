package note

import (
	"context"
	"errors"
	"time"
	"todoapp/internal/entity"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	noteRepo entity.NoteRepository
	taskRepo entity.TaskRepository
}

func NewService(noteRepo entity.NoteRepository, taskRepo entity.TaskRepository) *Service {
	return &Service{
		noteRepo: noteRepo,
		taskRepo: taskRepo,
	}
}

func (s *Service) Create(ctx context.Context, authorID, taskID int, text string, meta map[string]any) error {
	if err := s.checkTaskOwnership(ctx, authorID, taskID); err != nil {
		return err
	}

	if text == "" {
		return errors.New("text is required")
	}
	note := &entity.Note{
		TaskID:    taskID,
		AuthorID:  authorID,
		Text:      text,
		Meta:      meta,
		CreatedAt: time.Now(),
	}
	return s.noteRepo.Create(ctx, note)
}

func (s *Service) GetByTaskID(ctx context.Context, userID int, taskID int) ([]*entity.Note, error) {
	if err := s.checkTaskOwnership(ctx, userID, taskID); err != nil {
		return nil, err
	}

	return s.noteRepo.GetByTaskID(ctx, taskID)
}

func (s *Service) Delete(ctx context.Context, userID int, noteID bson.ObjectID) error {
	note, err := s.noteRepo.GetByID(ctx, noteID)
	if err != nil {
		return err
	}

	if note.AuthorID != userID {
		return entity.ErrAccessDenied
	}

	return s.noteRepo.Delete(ctx, noteID)
}

func (s *Service) checkTaskOwnership(ctx context.Context, userID int, taskID int) error {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		return err
	}

	if task.UserID != userID {
		return entity.ErrAccessDenied
	}

	return nil
}
