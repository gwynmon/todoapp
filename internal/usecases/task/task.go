package task

import (
	"context"
	"errors"
	"time"
	"todoapp/internal/entity"
	"todoapp/pkg/broker"
	"todoapp/pkg/cache"
)

type Service struct {
	taskRepo entity.TaskRepository
	noteRepo entity.NoteRepository
	cache    *cache.Cache
	producer *broker.Producer
}

func NewService(taskRepo entity.TaskRepository, noteRepo entity.NoteRepository, cache *cache.Cache, producer *broker.Producer) *Service {
	return &Service{
		taskRepo: taskRepo,
		noteRepo: noteRepo,
		cache:    cache,
		producer: producer,
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
	if err := s.taskRepo.Create(ctx, task); err != nil {
		return err
	}
	if s.cache != nil {
		s.cache.Delete(ctx, cache.TasksKey(userID))
	}
	if s.producer != nil {
		s.producer.Publish(ctx, "task.created", broker.Event{
			Type:      "task.created",
			TaskID:    task.ID,
			UserID:    userID,
			Timestamp: time.Now(),
			Payload:   map[string]any{"title": task.Title, "status": task.Status},
		})
	}
	return nil
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

func (s *Service) GetByUser(ctx context.Context, userID int, filter entity.TaskFilter) ([]entity.Task, error) {
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	if filter.Limit > 100 {
		filter.Limit = 100
	}

	if filter.Status == nil && filter.Offset == 0 && s.cache != nil {
		var tasks []entity.Task
		if err := s.cache.Get(ctx, cache.TasksKey(userID), &tasks); err == nil {
			return tasks, nil
		}
		tasks, err := s.taskRepo.GetByUser(ctx, userID, filter)
		if err != nil {
			return nil, err
		}
		s.cache.Set(ctx, cache.TasksKey(userID), tasks)
		return tasks, nil
	}
	return s.taskRepo.GetByUser(ctx, userID, filter)
}

func (s *Service) Update(ctx context.Context, userID, taskID int, input entity.UpdateTaskInput) error {
	if input.Status != nil && !isValidStatusTransition(*input.Status) {
		return entity.ErrInvalidStatus
	}
	if err := s.taskRepo.Update(ctx, userID, taskID, input); err != nil {
		return err
	}
	if s.cache != nil {
		s.cache.Delete(ctx, cache.TasksKey(userID))
	}
	if s.producer != nil && input.Status != nil {
		s.producer.Publish(ctx, "task.status_changed", broker.Event{
			Type:      "task.status_changed",
			TaskID:    taskID,
			UserID:    userID,
			Timestamp: time.Now(),
			Payload:   map[string]any{"status": *input.Status},
		})
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID int, taskID int) error {
	if err := s.taskRepo.Delete(ctx, userID, taskID); err != nil {
		if errors.Is(err, entity.ErrTaskNotFound) || errors.Is(err, entity.ErrNotFoundOrAccessDenied) {
			return entity.ErrNotFoundOrAccessDenied
		}
		return err
	}
	if s.cache != nil {
		s.cache.Delete(ctx, cache.TasksKey(userID))
	}
	if s.producer != nil {
		s.producer.Publish(ctx, "task.deleted", broker.Event{
			Type:      "task.deleted",
			TaskID:    taskID,
			UserID:    userID,
			Timestamp: time.Now(),
			Payload:   nil,
		})
	}
	return nil
}

func isValidStatusTransition(status string) bool {
	return status == "todo" || status == "in_progress" || status == "done"
}
