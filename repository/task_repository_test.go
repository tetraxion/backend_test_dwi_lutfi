package repository

import (
	"testing"
	"time"

	"task-tracker-backend/model"
)

// newTestRepo returns a repository without seed data for isolated tests.
func newTestRepo() *TaskRepository {
	return &TaskRepository{}
}

// ── GetAll ────────────────────────────────────────────────────────────────────

func TestGetAll_Empty(t *testing.T) {
	repo := newTestRepo()
	tasks, err := repo.GetAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("expected 0 tasks, got %d", len(tasks))
	}
}

func TestGetAll_SortedNewestFirst(t *testing.T) {
	repo := newTestRepo()
	// Create two tasks; add a small sleep so timestamps are distinct.
	first, err := repo.Create(model.CreateTaskRequest{Title: "First", Description: "desc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(2 * time.Millisecond)
	second, err := repo.Create(model.CreateTaskRequest{Title: "Second", Description: "desc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tasks, err := repo.GetAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(tasks) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(tasks))
	}
	// Newest (second) should come first.
	if tasks[0].ID != second.ID {
		t.Errorf("expected newest task first, got title=%q", tasks[0].Title)
	}
	if tasks[1].ID != first.ID {
		t.Errorf("expected oldest task second, got title=%q", tasks[1].Title)
	}
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestGetByID_Found(t *testing.T) {
	repo := newTestRepo()
	created, err := repo.Create(model.CreateTaskRequest{Title: "T", Description: "D"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != created.ID {
		t.Errorf("ID mismatch: want %s, got %s", created.ID, got.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	repo := newTestRepo()
	_, err := repo.GetByID("non-existent-id")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ── Create ────────────────────────────────────────────────────────────────────

func TestCreate_DefaultsPendingStatus(t *testing.T) {
	repo := newTestRepo()
	task, err := repo.Create(model.CreateTaskRequest{Title: "New", Description: "Desc"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if task.Status != model.StatusPending {
		t.Errorf("expected status %q, got %q", model.StatusPending, task.Status)
	}
	if task.ID == "" {
		t.Error("expected non-empty ID")
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected non-zero CreatedAt")
	}
	if task.UpdatedAt.IsZero() {
		t.Error("expected non-zero UpdatedAt")
	}
}

func TestCreate_StoresTask(t *testing.T) {
	repo := newTestRepo()
	req := model.CreateTaskRequest{Title: "Store me", Description: "persisted"}
	task, err := repo.Create(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 task in store, got %d", len(all))
	}
	if all[0].ID != task.ID {
		t.Errorf("stored task ID mismatch")
	}
}

// ── UpdateStatus ──────────────────────────────────────────────────────────────

func TestUpdateStatus_Success(t *testing.T) {
	repo := newTestRepo()
	task, err := repo.Create(model.CreateTaskRequest{Title: "T", Description: "D"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Sleep briefly so UpdatedAt will be strictly after CreatedAt/original UpdatedAt.
	time.Sleep(2 * time.Millisecond)

	updated, err := repo.UpdateStatus(task.ID, model.StatusDone)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if updated.Status != model.StatusDone {
		t.Errorf("expected status %q, got %q", model.StatusDone, updated.Status)
	}
	if !updated.UpdatedAt.After(task.UpdatedAt) {
		t.Error("expected UpdatedAt to be refreshed")
	}
}

func TestUpdateStatus_NotFound(t *testing.T) {
	repo := newTestRepo()
	_, err := repo.UpdateStatus("ghost-id", model.StatusDone)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	repo := newTestRepo()
	task, err := repo.Create(model.CreateTaskRequest{Title: "Delete me", Description: "D"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.Delete(task.ID); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := repo.GetByID(task.ID); err != ErrNotFound {
		t.Error("expected task to be gone after deletion")
	}
}

func TestDelete_NotFound(t *testing.T) {
	repo := newTestRepo()
	err := repo.Delete("no-such-id")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestDelete_LeavesOtherTasksIntact(t *testing.T) {
	repo := newTestRepo()
	a, err := repo.Create(model.CreateTaskRequest{Title: "A", Description: "D"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	b, err := repo.Create(model.CreateTaskRequest{Title: "B", Description: "D"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.Delete(a.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	all, err := repo.GetAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1 remaining task, got %d", len(all))
	}
	if all[0].ID != b.ID {
		t.Errorf("wrong task remaining: got %s, want %s", all[0].ID, b.ID)
	}
}
