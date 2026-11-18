package task_test

import (
	"context"
	"sort"
	"time"

	"gorm.io/gorm"

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
	t.ID = f.nextID
	f.nextID++
	t.CreatedAt = time.Now()
	t.UpdatedAt = t.CreatedAt
	f.tasks[t.ID] = *t
	return nil
}

func (f *fakeRepo) List(ctx context.Context) ([]task.Task, error) {
	out := make([]task.Task, 0, len(f.tasks))
	for _, t := range f.tasks {
		out = append(out, t)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID > out[j].ID })
	return out, nil
}

func (f *fakeRepo) GetByID(ctx context.Context, id uint64) (*task.Task, error) {
	t, ok := f.tasks[id]
	if !ok {
		return nil, gorm.ErrRecordNotFound
	}
	return &t, nil
}
