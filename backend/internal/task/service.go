package task

import (
	"context"
	"strings"

	"taskmanager/internal/apperr"
)

type CreateInput struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	Status      Status  `json:"status"`
	Assignee    *string `json:"assignee"`
}

type UpdateInput struct {
	Title       string
	Description *string
	Status      Status
	Assignee    *string
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
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
		Assignee:    in.Assignee,
	}
	if t.Status == "" {
		t.Status = StatusTodo
	}
	if err := s.repo.Create(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) List(ctx context.Context, q ListQuery) (ListResult, error) {
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
	t.Assignee = in.Assignee
	if err := s.repo.Update(ctx, t); err != nil {
		return nil, err
	}
	return t, nil
}

func (s *Service) Delete(ctx context.Context, id uint64) error {
	return s.repo.Delete(ctx, id)
}
