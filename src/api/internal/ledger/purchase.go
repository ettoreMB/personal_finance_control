package ledger

import (
	"errors"
	"strings"
	"time"

	"gorm.io/gorm"
)

const (
	MinInstallmentCount = 2
	MaxInstallmentCount = 24
)

var (
	ErrEmptyPurchaseDescription = errors.New("description is required")
	ErrInvalidInstallmentCount  = errors.New("installment_count must be between 2 and 24")
	ErrAmountSmallerThanN       = errors.New("amount_cents must be at least installment_count")
	ErrParcelaImmutable         = errors.New("parcela cannot be edited as a standalone lançamento")
	ErrPurchaseIDNotAllowed     = errors.New("purchase_id is not allowed")
	ErrCannotUndoPurchase       = errors.New("cannot undo a compra with past parcelas")
	ErrCannotEditTotal          = errors.New("cannot edit total: no mutable parcelas or leftover too small")
	ErrPurchaseFieldImmutable   = errors.New("installment_count and purchase_date cannot be changed")
)

type Purchase struct {
	ID               uint `gorm:"primaryKey"`
	Description      string
	PurchaseDate     time.Time
	AmountCents      int
	InstallmentCount int
	CategoryID       uint
	CategoryName     string
	Category         Category `gorm:"foreignKey:CategoryID"`
	Entries          []Entry         `gorm:"foreignKey:PurchaseID"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func (Purchase) TableName() string {
	return "purchases"
}

func NormalizeDescription(description string) (string, error) {
	trimmed := strings.TrimSpace(description)
	if trimmed == "" {
		return "", ErrEmptyPurchaseDescription
	}
	return trimmed, nil
}

func ParseInstallmentCount(n int) (int, error) {
	if n < MinInstallmentCount || n > MaxInstallmentCount {
		return 0, ErrInvalidInstallmentCount
	}
	return n, nil
}

func ParsePurchaseAmount(amountCents, n int) (int, error) {
	if amountCents < n {
		return 0, ErrAmountSmallerThanN
	}
	if _, err := ParseAmountCents(amountCents); err != nil {
		return 0, err
	}
	return amountCents, nil
}

func RedistributeRemaining(entries []Entry, newTotal int, today time.Time) error {
	mutable := make([]int, 0, len(entries))
	pastSum := 0
	for i, entry := range entries {
		if IsPastMonth(entry.EntryDate, today) {
			pastSum += entry.AmountCents
			continue
		}
		mutable = append(mutable, i)
	}
	if len(mutable) == 0 || newTotal-pastSum < len(mutable) {
		return ErrCannotEditTotal
	}
	parts := SplitCents(newTotal-pastSum, len(mutable))
	for i, idx := range mutable {
		entries[idx].AmountCents = parts[i]
	}
	return nil
}

func HasPastParcela(entries []Entry, today time.Time) bool {
	for _, entry := range entries {
		if IsPastMonth(entry.EntryDate, today) {
			return true
		}
	}
	return false
}

func GenerateParcelas(purchase Purchase) []Entry {
	amounts := SplitCents(purchase.AmountCents, purchase.InstallmentCount)
	entries := make([]Entry, 0, purchase.InstallmentCount)
	purchaseID := purchase.ID
	for i := 0; i < purchase.InstallmentCount; i++ {
		number := i + 1
		entries = append(entries, Entry{
			Type:              EntryTypeExpense,
			AmountCents:       amounts[i],
			EntryDate:         AddMonthsClamped(purchase.PurchaseDate, i),
			CategoryID:        purchase.CategoryID,
			PurchaseID:        &purchaseID,
			InstallmentNumber: &number,
		})
	}
	return entries
}
