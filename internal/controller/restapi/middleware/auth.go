package middleware

import (
	"context"
	"errors"
	"time"

	"todoapp/internal/entity"
	"todoapp/pkg/jwt"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      entity.UserRepository
	jwtSecret string
	jwtExpire time.Duration
}

func NewAuthService(repo entity.UserRepository, jwtSecret string, jwtExpire time.Duration) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: jwtSecret,
		jwtExpire: jwtExpire,
	}
}

func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

func (s *AuthService) Register(ctx context.Context, input entity.RegisterInput) error {
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
		if isUniqueViolation(err) {
			return entity.ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

func (s *AuthService) Login(ctx context.Context, input entity.LoginInput) (string, error) {
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
