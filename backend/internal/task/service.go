package task

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"taskmanager/internal/apperr"
)

type CreateInput struct {
	Title       string
	Description *string
	Status      Status
	Assignee    *string
}

type UpdateInput struct {
	Title       string
	Description *string
	Status      Status
	Assignee    *string
}

type Service struct {
	repo  Repository
	cache ListCache // nil disables caching
}

func NewService(repo Repository, cache ListCache) *Service {
	return &Service{repo: repo, cache: cache}
}

// invalidate runs after a successful write. The write already succeeded, so a
// cache failure is only logged; stale lists expire on their own after the TTL.
// It ignores cancellation so a client that disconnects after the commit still
// clears the cache.
func (s *Service) invalidate(ctx context.Context) {
	if s.cache == nil {
		return
	}
	if err := s.cache.InvalidateLists(context.WithoutCancel(ctx)); err != nil {
		slog.Warn("invalidate list cache", "error", err)
	}
}

// normaliseAssignee trims spaces and maps blank to nil, because the list
// filter trims too and the collation would not match " Ana " to "Ana".
func normaliseAssignee(a *string) *string {
	if a == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*a)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, &apperr.ValidationError{Details: map[string]string{"title": "is required"}}
	}
	t := &Task{
		Title:       title,
		Description: in.Description,
		Status:      in.Status,
		Assignee:    normaliseAssignee(in.Assignee),
	}
	if t.Status == "" {
		t.Status = StatusTodo
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	s.invalidate(ctx)
	return t, nil
}

// List reports whether the result came from the cache.
func (s *Service) List(ctx context.Context, q ListQuery) (ListResult, bool, error) {
	if s.cache == nil {
		res, err := s.listFromRepo(ctx, q)
		return res, false, err
	}
	key := q.CacheKey()
	if raw, found, err := s.cache.Get(ctx, key); err != nil {
		slog.Warn("read list cache", "error", err)
	} else if found {
		var res ListResult
		if err := json.Unmarshal(raw, &res); err == nil {
			return res, true, nil
		}
		slog.Warn("decode cached list", "key", key)
	}

	res, err := s.listFromRepo(ctx, q)
	if err != nil {
		return ListResult{}, false, err
	}
	if raw, err := json.Marshal(res); err != nil {
		slog.Warn("encode list for cache", "error", err)
	} else if err := s.cache.Set(ctx, key, raw, listCacheTTL); err != nil {
		slog.Warn("write list cache", "error", err)
	}
	return res, false, nil
}

func (s *Service) listFromRepo(ctx context.Context, q ListQuery) (ListResult, error) {
	tasks, total, err := s.repo.List(ctx, q)
	if err != nil {
		return ListResult{}, err
	}
	return ListResult{
		Data: tasks,
		Meta: Meta{
			Page:       q.Page,
			Limit:      q.Limit,
			Total:      total,
			TotalPages: int((total + int64(q.Limit) - 1) / int64(q.Limit)),
		},
	}, nil
}

func (s *Service) Get(ctx context.Context, id uint64) (*Task, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) Update(ctx context.Context, id uint64, in UpdateInput) (*Task, error) {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return nil, &apperr.ValidationError{Details: map[string]string{"title": "is required"}}
	}
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Title = title
	t.Description = in.Description
	t.Status = in.Status
	t.Assignee = normaliseAssignee(in.Assignee)
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	s.invalidate(ctx)
	return t, nil
}

func (s *Service) Delete(ctx context.Context, id uint64) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	s.invalidate(ctx)
	return nil
}
