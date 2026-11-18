package task_test

import (
	"context"
	"sort"
	"strings"
	"time"

	"gorm.io/gorm"

	"taskmanager/internal/apperr"
	"taskmanager/internal/task"
)

type fakeRepo struct {
	tasks  map[uint64]task.Task
	nextID uint64
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{tasks: map[uint64]task.Task{}, nextID: 1}
}

func (f *fakeRepo) Create(ctx context.Context, t *task.Task) error {
	for _, existing := range f.tasks {
		if !existing.DeletedAt.Valid && strings.EqualFold(existing.Title, t.Title) {
			return apperr.ErrDuplicateTitle
		}
	}
	t.ID = f.nextID
	f.nextID++
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	f.tasks[t.ID] = *t
	return nil
}

// List filters by status and assignee only and sorts newest first; keyword
// and field sorting are covered by the repository integration tests.
func (f *fakeRepo) List(ctx context.Context, q task.ListQuery) ([]task.Task, int64, error) {
	var out []task.Task
	for _, t := range f.tasks {
		if t.DeletedAt.Valid || (q.Status != "" && t.Status != q.Status) {
			continue
		}
		if q.Assignee != "" && (t.Assignee == nil || !strings.EqualFold(*t.Assignee, q.Assignee)) {
			continue
		}
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })

	total := int64(len(out))
	start := min((q.Page-1)*q.Limit, len(out))
	end := min(start+q.Limit, len(out))
	return out[start:end], total, nil
}

func (f *fakeRepo) Update(ctx context.Context, t *task.Task) error {
	for id, existing := range f.tasks {
		if id != t.ID && !existing.DeletedAt.Valid && strings.EqualFold(existing.Title, t.Title) {
			return apperr.ErrDuplicateTitle
		}
	}
	t.UpdatedAt = time.Now()
	f.tasks[t.ID] = *t
	return nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id uint64) (*task.Task, error) {
	t, ok := f.tasks[id]
	if !ok || t.DeletedAt.Valid {
		return nil, apperr.ErrNotFound
	}
	return &t, nil
}

func (f *fakeRepo) Delete(ctx context.Context, id uint64) error {
	t, ok := f.tasks[id]
	if !ok || t.DeletedAt.Valid {
		return apperr.ErrNotFound
	}
	t.DeletedAt = gorm.DeletedAt{Time: time.Now(), Valid: true}
	f.tasks[id] = t
	return nil
}
