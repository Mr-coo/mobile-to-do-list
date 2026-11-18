package task_test

import (
	"context"
	"errors"
	"testing"

	"gorm.io/gorm"

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

func softDelete(t *testing.T, db *gorm.DB, id uint64) {
	t.Helper()
	if err := db.Exec("UPDATE tasks SET deleted_at = NOW(3) WHERE id = ?", id).Error; err != nil {
		t.Fatal(err)
	}
}

func TestList_ExcludesDeleted(t *testing.T) {
	db := dbtest.New(t)
	repo := task.NewGormRepository(db)
	ctx := context.Background()

	keep := &task.Task{Title: "Keep", Status: task.StatusTodo}
	gone := &task.Task{Title: "Gone", Status: task.StatusTodo}
	for _, tk := range []*task.Task{keep, gone} {
		if err := repo.Create(ctx, tk); err != nil {
			t.Fatal(err)
		}
	}
	softDelete(t, db, gone.ID)

	got, err := repo.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].ID != keep.ID {
		t.Fatalf("List = %+v, want only task %d", got, keep.ID)
	}
}

func TestGetByID_Deleted_ReturnsErrNotFound(t *testing.T) {
	db := dbtest.New(t)
	repo := task.NewGormRepository(db)
	ctx := context.Background()

	tk := &task.Task{Title: "Gone", Status: task.StatusTodo}
	if err := repo.Create(ctx, tk); err != nil {
		t.Fatal(err)
	}
	softDelete(t, db, tk.ID)

	if _, err := repo.GetByID(ctx, tk.ID); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}
