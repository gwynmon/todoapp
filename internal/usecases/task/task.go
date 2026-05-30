package task

import (
	"context"
	"todoapp/internal/entity"
)

type Service struct {
	repo entity.TaskRepository
}

func NewService(repo entity.TaskRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, userID int, input entity.Task) error {
	task := &entity.Task{
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Deadline:    input.Deadline,
		Status:      "todo",
	}
	return s.repo.Create(ctx, task)
}

func (s *Service) GetByID(ctx context.Context, id int) (*entity.Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetByUser(ctx context.Context, userID int) ([]entity.Task, error) {
	return s.repo.GetByUser(ctx, userID)
}

func (s *Service) Update(ctx context.Context, userID int, input entity.Task) error {
	input.UserID = userID
	return s.repo.Update(ctx, &input)
}

func (s *Service) Delete(ctx context.Context, userID int, id int) error {
	return s.repo.Delete(ctx, id)
}
