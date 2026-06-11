package repository

import (
	"context"
	"errors"
	"time"

	"task-tracker-backend/model"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTaskRepository implementasi repositori Task berbasis PostgreSQL
type PostgresTaskRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresTaskRepository inisialisasi repositori PostgreSQL
func NewPostgresTaskRepository(pool *pgxpool.Pool) *PostgresTaskRepository {
	return &PostgresTaskRepository{pool: pool}
}

// GetAll ambil semua task dari DB
func (r *PostgresTaskRepository) GetAll() ([]model.Task, error) {
	const q = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		ORDER BY created_at DESC`

	rows, err := r.pool.Query(context.Background(), q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(
			&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if tasks == nil {
		return []model.Task{}, nil
	}
	return tasks, nil
}

// GetByID ambil satu task dari DB berdasarkan ID
func (r *PostgresTaskRepository) GetByID(id string) (model.Task, error) {
	const q = `
		SELECT id, title, description, status, created_at, updated_at
		FROM tasks
		WHERE id = $1`

	var t model.Task
	err := r.pool.QueryRow(context.Background(), q, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Task{}, ErrNotFound
		}
		return model.Task{}, err
	}
	return t, nil
}

// Create tambah task baru ke DB
func (r *PostgresTaskRepository) Create(req model.CreateTaskRequest) (model.Task, error) {
	now := time.Now()
	task := model.Task{
		ID:          uuid.NewString(),
		Title:       req.Title,
		Description: req.Description,
		Status:      model.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	const q = `
		INSERT INTO tasks (id, title, description, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, title, description, status, created_at, updated_at`

	err := r.pool.QueryRow(context.Background(), q,
		task.ID, task.Title, task.Description, task.Status, task.CreatedAt, task.UpdatedAt,
	).Scan(&task.ID, &task.Title, &task.Description, &task.Status, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return model.Task{}, err
	}
	return task, nil
}

// UpdateStatus update status task di DB
func (r *PostgresTaskRepository) UpdateStatus(id string, status model.TaskStatus) (model.Task, error) {
	const q = `
		UPDATE tasks
		SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, title, description, status, created_at, updated_at`

	var t model.Task
	err := r.pool.QueryRow(context.Background(), q, status, id).Scan(
		&t.ID, &t.Title, &t.Description, &t.Status, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Task{}, ErrNotFound
		}
		return model.Task{}, err
	}
	return t, nil
}

// Delete hapus task dari DB berdasarkan ID
func (r *PostgresTaskRepository) Delete(id string) error {
	const q = `DELETE FROM tasks WHERE id = $1`

	tag, err := r.pool.Exec(context.Background(), q, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
