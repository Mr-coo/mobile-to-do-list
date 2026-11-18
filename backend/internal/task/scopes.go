package task

import (
	"strings"

	"gorm.io/gorm"
)

var likeEscaper = strings.NewReplacer(`\`, `\`, `%`, `\%`, `_`, `\_`)

func WithStatus(s Status) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if s == "" {
			return db
		}
		return db.Where("status = ?", s)
	}
}

// Assignee matching is case-insensitive through the column's collation.
func WithAssignee(a string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if a == "" {
			return db
		}
		return db.Where("assignee = ?", a)
	}
}

// WithKeyword escapes \, % and _ so user input is matched literally.
func WithKeyword(k string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if k == "" {
			return db
		}
		pattern := "%" + likeEscaper.Replace(k) + "%"
		return db.Where("(title LIKE ? OR description LIKE ?)", pattern, pattern)
	}
}

// OrderBy takes the column from a Go whitelist, never from the request, and
// adds id as a tie-breaker so pages stay stable.
func OrderBy(sort string) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if sort == "" {
			return db
		}
		direction := "ASC"
		if strings.HasPrefix(sort, "-") {
			direction = "DESC"
		}
		column, ok := sortFields[strings.TrimPrefix(sort, "-")]
		if !ok {
			return db
		}
		return db.Order(column + " " + direction).Order("id " + direction)
	}
}

func Paginate(page, limit int) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if page < 1 || limit < 1 {
			return db
		}
		return db.Offset((page - 1) * limit).Limit(limit)
	}
}
