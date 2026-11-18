package task

import (
	"context"

	"gorm.io/gorm"
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
	return r.db.WithContext(ctx).Create(t).Error
}

func (r *GormRepository) List(ctx context.Context) ([]Task, error) {
	var tasks []Task
	err := r.db.WithContext(ctx).Order("created_at DESC").Find(&tasks).Error
	return tasks, err
}

func (r *GormRepository) GetByID(ctx context.Context, id uint64) (*Task, error) {
	var t Task
	if err := r.db.WithContext(ctx).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}
