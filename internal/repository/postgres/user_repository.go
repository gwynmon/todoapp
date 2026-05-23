package postgres

import (
	"context"
	"database/sql"
	"errors"
	"todoapp/internal/entity"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

const pgUniqueViolationCode = "23505"

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

var _ entity.UserRepository = (*UserRepository)(nil)

func (r *UserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `INSERT INTO users (name, email, password_hash, created_at) 
              VALUES (:name, :email, :password_hash, NOW())`

	_, err := r.db.NamedExecContext(ctx, query, user)
	if err == nil {
		return nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolationCode {
		return entity.ErrUserAlreadyExists
	}
	return err
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entity.User, error) {
	var u entity.User
	err := r.db.GetContext(ctx, &u, `
		SELECT id, name, email, password_hash, created_at 
		FROM users WHERE email = $1`, email)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.ErrUserNotFound
	}
	return &u, err
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*entity.User, error) {
	var u entity.User
	err := r.db.GetContext(ctx, &u, `
		SELECT id, name, email, password_hash, created_at 
		FROM users WHERE id = $1`, id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.ErrUserNotFound
	}
	return &u, err
}
