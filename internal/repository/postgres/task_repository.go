package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (r *TaskRepo) GetByUser(ctx context.Context, userID int, filter entity.TaskFilter) ([]entity.Task, error) {
	query := `SELECT id, user_id, title, description, status, deadline, created_at, updated_at 
              FROM tasks WHERE user_id = $1`
	args := []any{userID}
	idx := 2

	if filter.Status != nil {
		query += fmt.Sprintf(" AND status = $%d", idx)
		args = append(args, *filter.Status)
		idx++
	}

	query += " ORDER BY created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, filter.Limit, filter.Offset)

	var tasks []entity.Task
	if err := r.db.SelectContext(ctx, &tasks, query, args...); err != nil {
		return nil, err
	}
	return tasks, nil
}
func (r *TaskRepo) Update(ctx context.Context, userID, taskID int, input entity.UpdateTaskInput) error {
	query := `UPDATE tasks SET updated_at = NOW()`
	args := []any{taskID, userID}
	idx := 3

	if input.Title != nil {
		query += fmt.Sprintf(", title = $%d", idx)
		args = append(args, *input.Title)
		idx++
	}
	if input.Description != nil {
		query += fmt.Sprintf(", description = $%d", idx)
		args = append(args, *input.Description)
		idx++
	}
	if input.Status != nil {
		query += fmt.Sprintf(", status = $%d", idx)
		args = append(args, *input.Status)
		idx++
	}
	if input.Deadline != nil {
		query += fmt.Sprintf(", deadline = $%d", idx)
		args = append(args, *input.Deadline)
		idx++
	}

	if idx == 3 {
		return nil
	}

	query += ` WHERE id = $1 AND user_id = $2`

	res, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return entity.ErrNotFoundOrAccessDenied
	}
	return nil
}

func (r *TaskRepo) Delete(ctx context.Context, userID int, id int) error {
	res, err := r.db.ExecContext(ctx, "DELETE FROM tasks WHERE id = $1 AND user_id = $2", id, userID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return entity.ErrTaskNotFound
	}
	return nil
}
