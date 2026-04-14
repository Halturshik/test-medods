package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	taskdomain "example.com/taskservice/internal/domain/task"
)

type Repository struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Create(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		INSERT INTO tasks (parent_id, title, description, status, due_date, repeat_config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, parent_id, title, description, status, due_date, repeat_config, created_at, updated_at, generated_until
	`

	var repeatConfig []byte
	if task.RepeatConfig != nil {
		repeatConfig, _ = json.Marshal(task.RepeatConfig)
	}

	row := r.pool.QueryRow(ctx, query,
		task.ParentID, task.Title, task.Description, task.Status,
		task.DueDate, repeatConfig, task.CreatedAt, task.UpdatedAt,
	)

	return scanTask(row)
}

func (r *Repository) CreateIfNotExists(ctx context.Context, task *taskdomain.Task) error {
	const query = `
		INSERT INTO tasks (parent_id, title, description, status, due_date, repeat_config, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (parent_id, due_date) DO NOTHING
	`

	var repeatConfig []byte
	if task.RepeatConfig != nil {
		repeatConfig, _ = json.Marshal(task.RepeatConfig)
	}

	_, err := r.pool.Exec(ctx, query,
		task.ParentID, task.Title, task.Description, task.Status,
		task.DueDate, repeatConfig, task.CreatedAt, task.UpdatedAt,
	)
	return err
}

func (r *Repository) GetByID(ctx context.Context, id int64) (*taskdomain.Task, error) {
	const query = `
		SELECT id, parent_id, title, description, status, due_date, repeat_config,
		       created_at, updated_at, generated_until
		FROM tasks WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanTask(row)
}

func (r *Repository) Update(ctx context.Context, task *taskdomain.Task) (*taskdomain.Task, error) {
	const query = `
		UPDATE tasks
		SET title = $1, description = $2, status = $3, due_date = $4,
		    repeat_config = $5, updated_at = $6
		WHERE id = $7
		RETURNING id, parent_id, title, description, status, due_date, repeat_config,
		          created_at, updated_at, generated_until
	`

	var repeatConfig []byte
	if task.RepeatConfig != nil {
		repeatConfig, _ = json.Marshal(task.RepeatConfig)
	}

	row := r.pool.QueryRow(ctx, query,
		task.Title, task.Description, task.Status, task.DueDate,
		repeatConfig, task.UpdatedAt, task.ID,
	)

	return scanTask(row)
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	const query = `DELETE FROM tasks WHERE id = $1`
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return taskdomain.ErrNotFound
	}
	return nil
}

func (r *Repository) List(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, parent_id, title, description, status, due_date, repeat_config,
		       created_at, updated_at, generated_until
		FROM tasks ORDER BY due_date ASC, id DESC
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, rows.Err()
}

func (r *Repository) ListTemplates(ctx context.Context) ([]taskdomain.Task, error) {
	const query = `
		SELECT id, parent_id, title, description, status, due_date, repeat_config,
		       created_at, updated_at, generated_until
		FROM tasks
		WHERE parent_id IS NULL AND repeat_config IS NOT NULL
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []taskdomain.Task
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, *t)
	}
	return tasks, rows.Err()
}

func (r *Repository) CountChildren(ctx context.Context, parentID int64) (int, error) {
	const query = `SELECT COUNT(*) FROM tasks WHERE parent_id = $1`
	var count int
	err := r.pool.QueryRow(ctx, query, parentID).Scan(&count)
	return count, err
}

func (r *Repository) SetGeneratedUntil(ctx context.Context, templateID int64, t time.Time) error {
	const query = `UPDATE tasks SET generated_until = $1 WHERE id = $2`
	_, err := r.pool.Exec(ctx, query, t, templateID)
	return err
}

type taskScanner interface {
	Scan(dest ...any) error
}

func scanTask(scanner taskScanner) (*taskdomain.Task, error) {
	var (
		task           taskdomain.Task
		status         string
		rc             []byte
		parentID       *int64
		dueDate        time.Time
		generatedUntil *time.Time
	)

	if err := scanner.Scan(
		&task.ID, &parentID, &task.Title, &task.Description, &status,
		&dueDate, &rc, &task.CreatedAt, &task.UpdatedAt, &generatedUntil,
	); err != nil {
		return nil, err
	}

	task.ParentID = parentID
	task.DueDate = dueDate
	task.Status = taskdomain.Status(status)
	task.GeneratedUntil = generatedUntil

	if len(rc) > 0 {
		var repeat taskdomain.RepeatConfig
		if err := json.Unmarshal(rc, &repeat); err == nil {
			task.RepeatConfig = &repeat
		}
	}

	return &task, nil
}
