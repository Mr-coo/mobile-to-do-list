package task

import "context"

type CreateInput struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
	Status      Status  `json:"status"`
	Assignee    *string `json:"assignee"`
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, in CreateInput) (*Task, error) {
	t := &Task{
		Title:       in.Title,
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

func (s *Service) List(ctx context.Context) ([]Task, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id uint64) (*Task, error) {
	return s.repo.GetByID(ctx, id)
}
