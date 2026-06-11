package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"task-tracker-backend/model"
	"task-tracker-backend/repository"

	"github.com/gin-gonic/gin"
)

// ── Mock repository ───────────────────────────────────────────────────────────

// mockRepo is a minimal in-memory implementation of TaskRepo used in tests.
type mockRepo struct {
	tasks []model.Task
	err   error
}

func newMock(tasks ...model.Task) *mockRepo {
	return &mockRepo{tasks: tasks}
}

func sampleTask(id, title string, status model.TaskStatus) model.Task {
	now := time.Now()
	return model.Task{
		ID:          id,
		Title:       title,
		Description: "desc",
		Status:      status,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func (m *mockRepo) GetAll() ([]model.Task, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.tasks, nil
}

func (m *mockRepo) GetByID(id string) (model.Task, error) {
	for _, t := range m.tasks {
		if t.ID == id {
			return t, nil
		}
	}
	return model.Task{}, repository.ErrNotFound
}

func (m *mockRepo) Create(req model.CreateTaskRequest) (model.Task, error) {
	if m.err != nil {
		return model.Task{}, m.err
	}
	t := model.Task{
		ID:          "new-id",
		Title:       req.Title,
		Description: req.Description,
		Status:      model.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	m.tasks = append(m.tasks, t)
	return t, nil
}

func (m *mockRepo) UpdateStatus(id string, status model.TaskStatus) (model.Task, error) {
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks[i].Status = status
			return m.tasks[i], nil
		}
	}
	return model.Task{}, repository.ErrNotFound
}

func (m *mockRepo) Delete(id string) error {
	for i, t := range m.tasks {
		if t.ID == id {
			m.tasks = append(m.tasks[:i], m.tasks[i+1:]...)
			return nil
		}
	}
	return repository.ErrNotFound
}

// ── Test helpers ──────────────────────────────────────────────────────────────

func setupRouter(repo TaskRepo) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	h := NewTaskHandler(repo)
	v1 := r.Group("/api/v1")
	h.RegisterRoutes(v1)
	return r
}

func performRequest(r *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		_ = json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// ── GetAll ────────────────────────────────────────────────────────────────────

func TestGetAll_ReturnsAllTasks(t *testing.T) {
	tasks := []model.Task{
		sampleTask("1", "Task A", model.StatusPending),
		sampleTask("2", "Task B", model.StatusDone),
	}
	r := setupRouter(newMock(tasks...))
	w := performRequest(r, http.MethodGet, "/api/v1/tasks", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	count := int(resp["count"].(float64))
	if count != 2 {
		t.Errorf("expected count=2, got %d", count)
	}
}

func TestGetAll_EmptyStore(t *testing.T) {
	r := setupRouter(newMock())
	w := performRequest(r, http.MethodGet, "/api/v1/tasks", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if int(resp["count"].(float64)) != 0 {
		t.Error("expected count=0")
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestGetByID_Found(t *testing.T) {
	r := setupRouter(newMock(sampleTask("abc", "My Task", model.StatusPending)))
	w := performRequest(r, http.MethodGet, "/api/v1/tasks/abc", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	r := setupRouter(newMock())
	w := performRequest(r, http.MethodGet, "/api/v1/tasks/missing", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_ValidPayload(t *testing.T) {
	r := setupRouter(newMock())
	body := map[string]string{"title": "New Task", "description": "Some description"}
	w := performRequest(r, http.MethodPost, "/api/v1/tasks", body)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestCreate_MissingTitle(t *testing.T) {
	r := setupRouter(newMock())
	body := map[string]string{"description": "No title provided"}
	w := performRequest(r, http.MethodPost, "/api/v1/tasks", body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreate_MissingDescription(t *testing.T) {
	r := setupRouter(newMock())
	body := map[string]string{"title": "No desc"}
	w := performRequest(r, http.MethodPost, "/api/v1/tasks", body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreate_EmptyBody(t *testing.T) {
	r := setupRouter(newMock())
	w := performRequest(r, http.MethodPost, "/api/v1/tasks", nil)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

// ── UpdateStatus ──────────────────────────────────────────────────────────────

func TestUpdateStatus_ValidTransition(t *testing.T) {
	r := setupRouter(newMock(sampleTask("t1", "T", model.StatusPending)))
	body := map[string]string{"status": "done"}
	w := performRequest(r, http.MethodPatch, "/api/v1/tasks/t1/status", body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d — body: %s", w.Code, w.Body.String())
	}
}

func TestUpdateStatus_InvalidStatus(t *testing.T) {
	r := setupRouter(newMock(sampleTask("t1", "T", model.StatusPending)))
	body := map[string]string{"status": "in_progress"}
	w := performRequest(r, http.MethodPatch, "/api/v1/tasks/t1/status", body)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestUpdateStatus_NotFound(t *testing.T) {
	r := setupRouter(newMock())
	body := map[string]string{"status": "done"}
	w := performRequest(r, http.MethodPatch, "/api/v1/tasks/ghost/status", body)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	r := setupRouter(newMock(sampleTask("del-1", "Delete me", model.StatusPending)))
	w := performRequest(r, http.MethodDelete, "/api/v1/tasks/del-1", nil)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	var resp map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["message"] != "task deleted" {
		t.Errorf("unexpected message: %v", resp["message"])
	}
}

func TestDelete_NotFound(t *testing.T) {
	r := setupRouter(newMock())
	w := performRequest(r, http.MethodDelete, "/api/v1/tasks/no-exist", nil)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

// ── Error propagation ─────────────────────────────────────────────────────────

func TestGetAll_InternalError(t *testing.T) {
	mock := newMock()
	mock.err = repository.ErrNotFound // Mock generic error
	r := setupRouter(mock)
	w := performRequest(r, http.MethodGet, "/api/v1/tasks", nil)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestCreate_InternalError(t *testing.T) {
	mock := newMock()
	mock.err = repository.ErrNotFound // Mock generic error
	r := setupRouter(mock)
	body := map[string]string{"title": "Fail Task", "description": "Some description"}
	w := performRequest(r, http.MethodPost, "/api/v1/tasks", body)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
