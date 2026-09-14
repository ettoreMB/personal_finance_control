package ledger

import (
	"errors"
	"strings"
	"time"
)

var ErrEmptyCategoryName = errors.New("name is required")

type Category struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Category) TableName() string {
	return "categories"
}

func NormalizeCategoryName(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", ErrEmptyCategoryName
	}
	return trimmed, nil
}
