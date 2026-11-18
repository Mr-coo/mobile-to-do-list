package task

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"taskmanager/internal/apperr"
)

type Repository interface {
	Create(ctx context.Context, t *Task) error
	List(ctx context.Context) ([]Task, error)
	GetByID(ctx context.Context, id uint64) (*Task, error)
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

func (r *GormRepository) List(ctx context.Context) ([]Task, error) {
	var tasks []Task
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
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
