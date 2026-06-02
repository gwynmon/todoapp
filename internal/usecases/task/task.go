package task

import (
	"context"
	"errors"
	"todoapp/internal/entity"
)

type Service struct {
	taskRepo entity.TaskRepository
	noteRepo entity.NoteRepository
}

func NewService(taskRepo entity.TaskRepository, noteRepo entity.NoteRepository) *Service {
	return &Service{
		taskRepo: taskRepo,
		noteRepo: noteRepo,
	}
}

func (s *Service) Create(ctx context.Context, userID int, input entity.CreateTaskInput) error {
	task := &entity.Task{
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Deadline:    input.Deadline,
		Status:      "todo",
	}
	return s.taskRepo.Create(ctx, task)
}

func (s *Service) GetByID(ctx context.Context, userID, taskID int) (*entity.Task, error) {
	task, err := s.taskRepo.GetByID(ctx, taskID)
	if err != nil {
		if errors.Is(err, entity.ErrTaskNotFound) {
			return nil, entity.ErrNotFoundOrAccessDenied
		}
		return nil, err
	}

	if task.UserID != userID {
		return nil, entity.ErrNotFoundOrAccessDenied
	}

	notes, err := s.noteRepo.GetByTaskID(ctx, taskID)
	if err != nil {
		task.Notes = []*entity.Note{}
	} else {
		task.Notes = notes
	}

	return task, nil
}

func (s *Service) GetByUser(ctx context.Context, userID int) ([]entity.Task, error) {
	// TODO (п. 1.4): добавить фильтрацию по статусу и пагинацию
	return s.taskRepo.GetByUser(ctx, userID)
}

func (s *Service) Update(ctx context.Context, userID, taskID int, input entity.UpdateTaskInput) error {
	if input.Status != nil && !isValidStatusTransition(*input.Status) {
		return entity.ErrInvalidStatus
	}
	return s.taskRepo.Update(ctx, userID, taskID, input)
}

func isValidStatusTransition(status string) bool {
	return status == "todo" || status == "in_progress" || status == "done"
}

func (s *Service) Delete(ctx context.Context, userID int, taskID int) error {
	err := s.taskRepo.Delete(ctx, userID, taskID)
	if err != nil {
		if errors.Is(err, entity.ErrTaskNotFound) || errors.Is(err, entity.ErrNotFoundOrAccessDenied) {
			return entity.ErrNotFoundOrAccessDenied
		}
		return err
	}
	return nil
}
