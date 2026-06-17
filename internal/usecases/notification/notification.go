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
