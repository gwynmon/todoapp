package auth

import (
	"context"
	"errors"
	"time"

	"todoapp/internal/entity"
	"todoapp/pkg/jwt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo      entity.UserRepository
	jwtSecret string
	jwtExpire time.Duration
}

func NewAuthService(repo entity.UserRepository, jwtSecret string, jwtExpire time.Duration) *Service {
	return &Service{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpire: jwtExpire,
	}
}

func (s *Service) Register(ctx context.Context, input entity.RegisterInput) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &entity.User{
		Name:         input.Name,
		Email:        input.Email,
		PasswordHash: string(hash),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return err
	}

	return nil
}

func (s *Service) Login(ctx context.Context, input entity.LoginInput) (string, error) {
	user, err := s.repo.GetByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return "", entity.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)); err != nil {
		return "", entity.ErrInvalidCredentials
	}

	token, err := jwt.Generate(user.ID, s.jwtSecret, s.jwtExpire)
	if err != nil {
		return "", err
	}

	return token, nil
}
