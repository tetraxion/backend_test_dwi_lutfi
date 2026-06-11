package handler

import (
	"errors"
	"net/http"

	"task-tracker-backend/model"
	"task-tracker-backend/repository"

	"github.com/gin-gonic/gin"
)

// TaskRepo interface kontrak untuk repositori data task
type TaskRepo interface {
	GetAll() ([]model.Task, error)
	GetByID(id string) (model.Task, error)
	Create(req model.CreateTaskRequest) (model.Task, error)
	UpdateStatus(id string, status model.TaskStatus) (model.Task, error)
	Delete(id string) error
}

type TaskHandler struct {
	repo TaskRepo
}

func NewTaskHandler(repo TaskRepo) *TaskHandler {
	return &TaskHandler{repo: repo}
}

// RegisterRoutes registrasi route endpoint API
func (h *TaskHandler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.GET("/tasks", h.GetAll)
	rg.GET("/tasks/:id", h.GetByID)
	rg.POST("/tasks", h.Create)
	rg.PATCH("/tasks/:id/status", h.UpdateStatus)
	rg.DELETE("/tasks/:id", h.Delete)
}

// GET /api/v1/tasks
func (h *TaskHandler) GetAll(c *gin.Context) {
	tasks, err := h.repo.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":  tasks,
		"count": len(tasks),
	})
}

// GET /api/v1/tasks/:id
func (h *TaskHandler) GetByID(c *gin.Context) {
	id := c.Param("id")

	task, err := h.repo.GetByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": task})
}

// POST /api/v1/tasks
func (h *TaskHandler) Create(c *gin.Context) {
	var req model.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.repo.Create(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": task})
}

// PATCH /api/v1/tasks/:id/status
func (h *TaskHandler) UpdateStatus(c *gin.Context) {
	id := c.Param("id")

	var req model.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.repo.UpdateStatus(id, req.Status)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": task})
}

// DELETE /api/v1/tasks/:id
func (h *TaskHandler) Delete(c *gin.Context) {
	id := c.Param("id")

	if err := h.repo.Delete(id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}
