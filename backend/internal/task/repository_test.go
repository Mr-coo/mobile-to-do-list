package task_test

import (
	"context"
	"errors"
	"testing"

	"taskmanager/internal/apperr"
	"taskmanager/internal/database/dbtest"
	"taskmanager/internal/task"
)

func TestCreate_DuplicateTitle_ReturnsErrDuplicateTitle(t *testing.T) {
	repo := task.NewGormRepository(dbtest.New(t))
	ctx := context.Background()

	if err := repo.Create(ctx, &task.Task{Title: "Buy milk", Status: task.StatusTodo}); err != nil {
		t.Fatal(err)
	}
	err := repo.Create(ctx, &task.Task{Title: "buy milk", Status: task.StatusTodo})
	if !errors.Is(err, apperr.ErrDuplicateTitle) {
		t.Fatalf("err = %v, want ErrDuplicateTitle", err)
	}
}

func TestCreate_ReuseTitleOfDeletedTask(t *testing.T) {
	db := dbtest.New(t)
	repo := task.NewGormRepository(db)
	ctx := context.Background()

	first := &task.Task{Title: "Buy milk", Status: task.StatusTodo}
	if err := repo.Create(ctx, first); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec("UPDATE tasks SET deleted_at = NOW(3) WHERE id = ?", first.ID).Error; err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, &task.Task{Title: "Buy milk", Status: task.StatusTodo}); err != nil {
		t.Fatalf("reusing a deleted task's title: %v", err)
	}
}
