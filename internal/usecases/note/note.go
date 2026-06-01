package note

import (
	"context"
	"errors"
	"time"
	"todoapp/internal/entity"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type Service struct {
	repo entity.NoteRepository
}

func NewService(repo entity.NoteRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, authorID, taskID int, text string, meta map[string]any) error {
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
	return s.repo.Create(ctx, note)
}

func (s *Service) GetByTaskID(ctx context.Context, taskID int) ([]entity.Note, error) {
	return s.repo.GetByTaskID(ctx, taskID)
}

func (s *Service) Delete(ctx context.Context, userID int, noteID bson.ObjectID) error {
	note, err := s.repo.GetByID(ctx, noteID)
	if err != nil {
		return err
	}

	if note.AuthorID != userID {
		return entity.ErrAccessDenied
	}

	return s.repo.Delete(ctx, noteID)
}
