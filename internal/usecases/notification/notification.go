package notification

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"todoapp/internal/entity"
	mongorepo "todoapp/internal/repository/mongo"
)

type Service struct {
	repo *mongorepo.NotificationRepo
	log  *slog.Logger
}

func NewService(repo *mongorepo.NotificationRepo, log *slog.Logger) *Service {
	return &Service{
		repo: repo,
		log:  log,
	}
}

func (s *Service) CreateStatusChangedNotification(ctx context.Context, taskID int64, userID int64, status string) error {
	notification := &entity.Notification{
		UserID:    userID,
		TaskID:    taskID,
		Type:      "task.status_changed",
		Message:   fmt.Sprintf("Task %d changed status to %s", taskID, status),
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		s.log.Error(
			"create notification failed",
			slog.String("error", err.Error()),
		)
		return err
	}

	s.log.Info(
		"notification created",
		slog.Int64("task_id", taskID),
		slog.Int64("user_id", userID),
	)

	return nil
}
func (s *Service) CreateFromEvent(ctx context.Context, event entity.TaskEvent) error {

	notification := &entity.Notification{
		UserID:    event.UserID,
		TaskID:    event.TaskID,
		Type:      event.EventType,
		Message:   buildMessage(event),
		CreatedAt: time.Now(),
	}

	if err := s.repo.Create(ctx, notification); err != nil {
		s.log.Error(
			"create notification failed",
			slog.String("error", err.Error()),
		)
		return err
	}

	return nil
}

func (s *Service) ListByUserID(ctx context.Context, userID int64) ([]*entity.Notification, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func buildMessage(event entity.TaskEvent) string {
	switch event.EventType {

	case "task.created":
		return fmt.Sprintf(
			"Task %d created",
			event.TaskID,
		)

	case "task.status_changed":
		return fmt.Sprintf(
			"Task %d status changed",
			event.TaskID,
		)

	case "task.deleted":
		return fmt.Sprintf(
			"Task %d deleted",
			event.TaskID,
		)

	default:
		return fmt.Sprintf(
			"Task %d event %s",
			event.TaskID,
			event.EventType,
		)
	}
}
