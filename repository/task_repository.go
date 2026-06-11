package repository

import (
	"errors"
	"sort"
	"sync"
	"time"

	"task-tracker-backend/model"

	"github.com/google/uuid"
)

// ErrNotFound error jika data task tidak ditemukan
var ErrNotFound = errors.New("task not found")

// TaskRepository penyimpan data task in-memory
type TaskRepository struct {
	mu    sync.RWMutex
	tasks []model.Task
}

// NewTaskRepository inisialisasi repositori in-memory
func NewTaskRepository() *TaskRepository {
	repo := &TaskRepository{}
	repo.seed()
	return repo
}

// seed data awal (mocking)
func (r *TaskRepository) seed() {
	now := time.Now()
	r.tasks = []model.Task{
		{
			ID:          uuid.NewString(),
			Title:       "Design database schema",
			Description: "Draft the ER diagram and define all table relationships for the project.",
			Status:      model.StatusDone,
			CreatedAt:   now.Add(-72 * time.Hour),
			UpdatedAt:   now.Add(-48 * time.Hour),
		},
		{
			ID:          uuid.NewString(),
			Title:       "Implement REST API",
			Description: "Build the Gin-based HTTP layer with CRUD endpoints for task management.",
			Status:      model.StatusDone,
			CreatedAt:   now.Add(-48 * time.Hour),
			UpdatedAt:   now.Add(-24 * time.Hour),
		},
		{
			ID:          uuid.NewString(),
			Title:       "Write unit tests",
			Description: "Cover all handler and repository functions with table-driven tests.",
			Status:      model.StatusPending,
			CreatedAt:   now.Add(-24 * time.Hour),
			UpdatedAt:   now.Add(-24 * time.Hour),
		},
	}
}

// GetAll ambil semua task sorted by CreatedAt descending
func (r *TaskRepository) GetAll() ([]model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]model.Task, len(r.tasks))
	copy(result, r.tasks)

	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}

// GetByID ambil satu task berdasarkan ID
func (r *TaskRepository) GetByID(id string) (model.Task, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, t := range r.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return model.Task{}, ErrNotFound
}

// Create tambah task baru
func (r *TaskRepository) Create(req model.CreateTaskRequest) (model.Task, error) {
	now := time.Now()
	task := model.Task{
		ID:          uuid.NewString(),
		Title:       req.Title,
		Description: req.Description,
		Status:      model.StatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.tasks = append(r.tasks, task)
	return task, nil
}

// UpdateStatus update status task
func (r *TaskRepository) UpdateStatus(id string, status model.TaskStatus) (model.Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, t := range r.tasks {
		if t.ID == id {
			r.tasks[i].Status = status
			r.tasks[i].UpdatedAt = time.Now()
			return r.tasks[i], nil
		}
	}
	return model.Task{}, ErrNotFound
}

// Delete hapus task berdasarkan ID
func (r *TaskRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for i, t := range r.tasks {
		if t.ID == id {
			r.tasks[i] = r.tasks[len(r.tasks)-1]
			r.tasks = r.tasks[:len(r.tasks)-1]
			return nil
		}
	}
	return ErrNotFound
}
