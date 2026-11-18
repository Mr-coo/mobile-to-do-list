package task

import (
	"time"

	"gorm.io/gorm"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
)

func (s Status) Valid() bool {
	return s == StatusTodo || s == StatusInProgress || s == StatusDone
}

type Task struct {
	ID          uint64         `json:"id" gorm:"primaryKey"`
	Title       string         `json:"title"`
	Description *string        `json:"description"`
	Status      Status         `json:"status"`
	Assignee    *string        `json:"assignee"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}
