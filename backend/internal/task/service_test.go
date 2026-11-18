package task_test

import (
	"context"
	"errors"
	"testing"

	"taskmanager/internal/apperr"
	"taskmanager/internal/task"
)

func ptr(s string) *string { return &s }

func seed(t *testing.T, svc *task.Service, title string) *task.Task {
	t.Helper()
	tk, err := svc.Create(context.Background(), task.CreateInput{Title: title})
	if err != nil {
		t.Fatal(err)
	}
	return tk
}

func TestUpdate_Success(t *testing.T) {
	svc := task.NewService(newFakeRepo())
	orig := seed(t, svc, "Old")

	got, err := svc.Update(context.Background(), orig.ID, task.UpdateInput{
		Title: "  New  ", Description: ptr("d"), Status: task.StatusDone, Assignee: ptr("bob"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "New" || got.Status != task.StatusDone ||
		got.Description == nil || *got.Description != "d" ||
		got.Assignee == nil || *got.Assignee != "bob" {
		t.Fatalf("unexpected task: %+v", got)
	}
	if got.UpdatedAt.Before(orig.UpdatedAt) {
		t.Fatalf("UpdatedAt %v before original %v", got.UpdatedAt, orig.UpdatedAt)
	}
}

func TestUpdate_NotFound(t *testing.T) {
	svc := task.NewService(newFakeRepo())
	_, err := svc.Update(context.Background(), 99, task.UpdateInput{Title: "x", Status: task.StatusTodo})
	if !errors.Is(err, apperr.ErrNotFound) {
		t.Fatalf("err = %v, want ErrNotFound", err)
	}
}

func TestUpdate_DuplicateTitle(t *testing.T) {
	svc := task.NewService(newFakeRepo())
	seed(t, svc, "One")
	two := seed(t, svc, "Two")

	_, err := svc.Update(context.Background(), two.ID, task.UpdateInput{Title: "one", Status: task.StatusTodo})
	if !errors.Is(err, apperr.ErrDuplicateTitle) {
		t.Fatalf("err = %v, want ErrDuplicateTitle", err)
	}
}

func TestUpdate_KeepOwnTitle(t *testing.T) {
	svc := task.NewService(newFakeRepo())
	tk := seed(t, svc, "Same")

	got, err := svc.Update(context.Background(), tk.ID, task.UpdateInput{Title: "SAME", Status: task.StatusDone})
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != "SAME" {
		t.Fatalf("title = %q", got.Title)
	}
}

func TestUpdate_WhitespaceTitle(t *testing.T) {
	svc := task.NewService(newFakeRepo())
	tk := seed(t, svc, "Real")

	_, err := svc.Update(context.Background(), tk.ID, task.UpdateInput{Title: "   ", Status: task.StatusTodo})
	var verr *apperr.ValidationError
	if !errors.As(err, &verr) {
		t.Fatalf("err = %v, want ValidationError", err)
	}
}
