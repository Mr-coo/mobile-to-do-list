package task

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"taskmanager/internal/apperr"
)

type Repository interface {
	Create(ctx context.Context, t *Task) error
	List(ctx context.Context, q ListQuery) ([]Task, int64, error)
	GetByID(ctx context.Context, id uint64) (*Task, error)
	Update(ctx context.Context, t *Task) error
}

type GormRepository struct {
	db *gorm.DB
}

func NewGormRepository(db *gorm.DB) *GormRepository {
	return &GormRepository{db: db}
}

func (r *GormRepository) Create(ctx context.Context, t *Task) error {
	err := r.db.WithContext(ctx).Create(t).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperr.ErrDuplicateTitle
	}
	return err
}

func (r *GormRepository) List(ctx context.Context, q ListQuery) ([]Task, int64, error) {
	filtered := r.db.WithContext(ctx).Model(&Task{}).
		Scopes(WithStatus(q.Status), WithAssignee(q.Assignee), WithKeyword(q.Keyword))

	var total int64
	if err := filtered.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var tasks []Task
	err := filtered.Scopes(OrderBy(q.Sort), Paginate(q.Page, q.Limit)).Find(&tasks).Error
	return tasks, total, err
}

func (r *GormRepository) GetByID(ctx context.Context, id uint64) (*Task, error) {
	var t Task
	err := r.db.WithContext(ctx).First(&t, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, apperr.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Select lists the columns explicitly so nil/empty values are written too.
func (r *GormRepository) Update(ctx context.Context, t *Task) error {
	err := r.db.WithContext(ctx).Model(t).
		Select("title", "description", "status", "assignee").Updates(t).Error
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return apperr.ErrDuplicateTitle
	}
	return err
}
