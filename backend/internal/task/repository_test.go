package task_test

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

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

	got, _, err := repo.List(ctx, task.ListQuery{Page: 1, Limit: 10, Sort: "-created_at"})
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

func TestList_Search(t *testing.T) {
	db := dbtest.New(t)
	repo := task.NewGormRepository(db)
	ctx := context.Background()

	str := func(s string) *string { return &s }
	base := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	fixtures := []task.Task{
		{Title: "Buy milk", Status: task.StatusTodo, Assignee: str("Ana")},
		{Title: "Fix login bug", Status: task.StatusInProgress, Assignee: str("Budi"), Description: str("users cannot log in")},
		{Title: "Write docs", Status: task.StatusDone, Assignee: str("Ana")},
		{Title: "Pay 100% invoice", Status: task.StatusTodo},
		{Title: "Buy bread", Status: task.StatusDone, Assignee: str("Citra")},
		{Title: `Fix C:\temp path`, Status: task.StatusInProgress},
	}
	for i := range fixtures {
		fixtures[i].CreatedAt = base.Add(time.Duration(i) * time.Hour)
		fixtures[i].UpdatedAt = fixtures[i].CreatedAt
		if err := repo.Create(ctx, &fixtures[i]); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name      string
		q         task.ListQuery
		wantTitle []string
		wantTotal int64
	}{
		{"keyword in title", task.ListQuery{Keyword: "milk"}, []string{"Buy milk"}, 1},
		{"keyword in description", task.ListQuery{Keyword: "log in"}, []string{"Fix login bug"}, 1},
		{"percent is literal", task.ListQuery{Keyword: "100%"}, []string{"Pay 100% invoice"}, 1},
		{"underscore is literal", task.ListQuery{Keyword: "_"}, []string{}, 0},
		{"backslash is literal", task.ListQuery{Keyword: `\`}, []string{`Fix C:\temp path`}, 1},
		{"status", task.ListQuery{Status: task.StatusDone}, []string{"Buy bread", "Write docs"}, 2},
		{"assignee case-insensitive", task.ListQuery{Assignee: "ana"}, []string{"Write docs", "Buy milk"}, 2},
		{"status and assignee", task.ListQuery{Status: task.StatusTodo, Assignee: "Ana"}, []string{"Buy milk"}, 1},
		{"sort by title", task.ListQuery{Sort: "title", Limit: 1}, []string{"Buy bread"}, 6},
		{"second page", task.ListQuery{Limit: 2, Page: 2}, []string{"Pay 100% invoice", "Write docs"}, 6},
		{"page past the end", task.ListQuery{Page: 99}, []string{}, 6},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := tt.q
			if q.Page == 0 {
				q.Page = 1
			}
			if q.Limit == 0 {
				q.Limit = 10
			}
			if q.Sort == "" {
				q.Sort = "-created_at"
			}
			got, total, err := repo.List(ctx, q)
			if err != nil {
				t.Fatal(err)
			}
			titles := []string{}
			for _, tk := range got {
				titles = append(titles, tk.Title)
			}
			if !reflect.DeepEqual(titles, tt.wantTitle) || total != tt.wantTotal {
				t.Fatalf("got %v (total %d), want %v (total %d)", titles, total, tt.wantTitle, tt.wantTotal)
			}
		})
	}
}

func TestUpdate_ClearsAssignee(t *testing.T) {
	repo := task.NewGormRepository(dbtest.New(t))
	ctx := context.Background()

	who := "bob"
	tk := &task.Task{Title: "Assigned", Status: task.StatusTodo, Assignee: &who}
	if err := repo.Create(ctx, tk); err != nil {
		t.Fatal(err)
	}
	tk.Assignee = nil
	if err := repo.Update(ctx, tk); err != nil {
		t.Fatal(err)
	}

	got, err := repo.GetByID(ctx, tk.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Assignee != nil {
		t.Fatalf("assignee = %q, want NULL", *got.Assignee)
	}
}

func TestDelete_SetsDeletedAt(t *testing.T) {
	db := dbtest.New(t)
	repo := task.NewGormRepository(db)
	ctx := context.Background()

	tk := &task.Task{Title: "Buy milk", Status: task.StatusTodo}
	if err := repo.Create(ctx, tk); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, tk.ID); err != nil {
		t.Fatal(err)
	}

	var count int64
	if err := db.Raw("SELECT COUNT(*) FROM tasks WHERE id = ? AND deleted_at IS NOT NULL", tk.ID).Scan(&count).Error; err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("soft-deleted rows = %d, want 1", count)
	}
	if _, err := repo.GetByID(ctx, tk.ID); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("GetByID err = %v, want ErrNotFound", err)
	}
}

func TestDelete_Twice_ReturnsNotFound(t *testing.T) {
	repo := task.NewGormRepository(dbtest.New(t))
	ctx := context.Background()

	tk := &task.Task{Title: "Buy milk", Status: task.StatusTodo}
	if err := repo.Create(ctx, tk); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, tk.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, tk.ID); !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("second delete err = %v, want ErrNotFound", err)
	}
}

func TestDelete_FreesTitle(t *testing.T) {
	repo := task.NewGormRepository(dbtest.New(t))
	ctx := context.Background()

	tk := &task.Task{Title: "Buy milk", Status: task.StatusTodo}
	if err := repo.Create(ctx, tk); err != nil {
		t.Fatal(err)
	}
	if err := repo.Delete(ctx, tk.ID); err != nil {
		t.Fatal(err)
	}
	if err := repo.Create(ctx, &task.Task{Title: "Buy milk", Status: task.StatusTodo}); err != nil {
		t.Fatalf("reusing deleted title: %v", err)
	}
}

func TestUpdate_DuplicateTitle_Repository(t *testing.T) {
	repo := task.NewGormRepository(dbtest.New(t))
	ctx := context.Background()

	if err := repo.Create(ctx, &task.Task{Title: "Buy milk", Status: task.StatusTodo}); err != nil {
		t.Fatal(err)
	}
	second := &task.Task{Title: "Buy bread", Status: task.StatusTodo}
	if err := repo.Create(ctx, second); err != nil {
		t.Fatal(err)
	}
	second.Title = "buy milk"
	if err := repo.Update(ctx, second); !errors.Is(err, apperr.ErrDuplicateTitle) {
		t.Fatalf("err = %v, want ErrDuplicateTitle", err)
	}
}
