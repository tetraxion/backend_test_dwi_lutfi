package model

import "time"

// TaskStatus tipe status task
type TaskStatus string

const (
	StatusPending TaskStatus = "pending"
	StatusDone    TaskStatus = "done"
)

// Task model data core
type Task struct {
	ID          string     `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      TaskStatus `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// CreateTaskRequest request payload untuk membuat task baru
type CreateTaskRequest struct {
	Title       string `json:"title"       binding:"required,min=1,max=100"`
	Description string `json:"description" binding:"required,min=1,max=500"`
}

// UpdateStatusRequest request payload untuk update status task
type UpdateStatusRequest struct {
	Status TaskStatus `json:"status" binding:"required,oneof=pending done"`
}
