package ledger

import (
	"errors"
	"time"

	"gorm.io/gorm"
)

const (
	EntryTypeIncome  = "income"
	EntryTypeExpense = "expense"
	entryDateLayout  = "2006-01-02"
)

var (
	ErrInvalidEntryType   = errors.New("type must be income or expense")
	ErrInvalidAmount      = errors.New("amount_cents must be greater than 0")
	ErrInvalidEntryDate   = errors.New("entry_date must be YYYY-MM-DD")
	ErrCategoryRequired   = errors.New("category_id is required")
	ErrCategoryNotFound   = errors.New("category not found")
)

type Entry struct {
	ID          uint `gorm:"primaryKey"`
	Type        string
	AmountCents int
	EntryDate   time.Time
	CategoryID  uint
	Category    Category `gorm:"foreignKey:CategoryID"`
	DeletedAt   gorm.DeletedAt
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Entry) TableName() string {
	return "entries"
}

func ParseEntryType(value string) (string, error) {
	if value == EntryTypeIncome || value == EntryTypeExpense {
		return value, nil
	}
	return "", ErrInvalidEntryType
}

func ParseAmountCents(amount int) (int, error) {
	if amount <= 0 {
		return 0, ErrInvalidAmount
	}
	return amount, nil
}

func ParseEntryDate(value string) (time.Time, error) {
	if value == "" {
		return time.Time{}, ErrInvalidEntryDate
	}
	parsed, err := time.ParseInLocation(entryDateLayout, value, time.UTC)
	if err != nil {
		return time.Time{}, ErrInvalidEntryDate
	}
	return parsed, nil
}

func FormatEntryDate(value time.Time) string {
	return value.UTC().Format(entryDateLayout)
}
