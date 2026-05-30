package postgres

import (
	"context"
	"database/sql"
	"errors"
	"todoapp/internal/entity"

	_ "github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

type TaskRepo struct {
	db *sqlx.DB
}

func NewTaskRepo(db *sqlx.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(ctx context.Context, task *entity.Task) error {
	query := `INSERT INTO tasks (user_id, title, description, status, deadline) 
              VALUES (:user_id, :title, :description, :status, :deadline) RETURNING id`

	rows, err := r.db.NamedQueryContext(ctx, query, task)
	if err != nil {
		return err
	}
	defer rows.Close()

	if rows.Next() {
		rows.Scan(&task.ID)
	}
	return nil
}

func (r *TaskRepo) GetByID(ctx context.Context, id int) (*entity.Task, error) {
	var task entity.Task
	err := r.db.GetContext(ctx, &task, `SELECT * FROM tasks WHERE id = $1`, id)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, entity.ErrTaskNotFound
	}
	return &task, err
}

func (r *TaskRepo) GetByUser(ctx context.Context, userID int) ([]entity.Task, error) {
	var tasks []entity.Task
	err := r.db.SelectContext(ctx, &tasks,
		`SELECT id, user_id, title, description, status, deadline, created_at, updated_at 
		 FROM tasks WHERE user_id = $1 ORDER BY created_at DESC`, userID)

	if err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *TaskRepo) Update(ctx context.Context, task *entity.Task) error {
	query := `UPDATE tasks SET title = :title, description = :description, 
              status = :status, deadline = :deadline, updated_at = NOW()
              WHERE id = :id AND user_id = :user_id`

	res, err := r.db.NamedExecContext(ctx, query, task)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return entity.ErrTaskNotFound
	}
	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return entity.ErrTaskNotFound
	}
	return nil
}
