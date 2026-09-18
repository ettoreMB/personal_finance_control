package server

import (
	"sort"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/ledger"
)

type summaryPeriodResponse struct {
	Kind  string `json:"kind"`
	Year  int    `json:"year"`
	Month int    `json:"month"`
}

type summaryCategoryResponse struct {
	CategoryID   uint   `json:"category_id"`
	CategoryName string `json:"category_name"`
	IncomeCents  int    `json:"income_cents"`
	ExpenseCents int    `json:"expense_cents"`
	BalanceCents int    `json:"balance_cents"`
}

type summaryResponse struct {
	Period       summaryPeriodResponse     `json:"period"`
	IncomeCents  int                       `json:"income_cents"`
	ExpenseCents int                       `json:"expense_cents"`
	BalanceCents int                       `json:"balance_cents"`
	Categories   []summaryCategoryResponse `json:"categories"`
}

func summaryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		today := ledger.TodayCivil()
		year, month, _ := today.Date()
		start := time.Date(year, month, 1, 0, 0, 0, 0, time.UTC)
		end := start.AddDate(0, 1, 0)

		var entries []ledger.Entry
		if err := db.Preload("Category").
			Where("entry_date >= ? AND entry_date < ?", start, end).
			Find(&entries).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to load summary")
		}

		return c.JSON(buildSummary(entries, year, int(month)))
	}
}

func buildSummary(entries []ledger.Entry, year, month int) summaryResponse {
	out := summaryResponse{
		Period:     summaryPeriodResponse{Kind: "month", Year: year, Month: month},
		Categories: make([]summaryCategoryResponse, 0),
	}

	type bucket struct {
		id   uint
		name string
		in   int
		out  int
	}
	byID := make(map[uint]*bucket)
	order := make([]uint, 0)

	for _, entry := range entries {
		switch entry.Type {
		case ledger.EntryTypeIncome:
			out.IncomeCents += entry.AmountCents
		case ledger.EntryTypeExpense:
			out.ExpenseCents += entry.AmountCents
		}
		b, ok := byID[entry.CategoryID]
		if !ok {
			b = &bucket{id: entry.CategoryID, name: entry.Category.Name}
			byID[entry.CategoryID] = b
			order = append(order, entry.CategoryID)
		}
		if entry.Type == ledger.EntryTypeIncome {
			b.in += entry.AmountCents
		} else if entry.Type == ledger.EntryTypeExpense {
			b.out += entry.AmountCents
		}
	}
	out.BalanceCents = out.IncomeCents - out.ExpenseCents

	cats := make([]summaryCategoryResponse, 0, len(order))
	for _, id := range order {
		b := byID[id]
		cats = append(cats, summaryCategoryResponse{
			CategoryID:   b.id,
			CategoryName: b.name,
			IncomeCents:  b.in,
			ExpenseCents: b.out,
			BalanceCents: b.in - b.out,
		})
	}
	sortSummaryCategories(cats)
	out.Categories = cats
	return out
}

func sortSummaryCategories(cats []summaryCategoryResponse) {
	sort.Slice(cats, func(i, j int) bool {
		if cats[i].ExpenseCents != cats[j].ExpenseCents {
			return cats[i].ExpenseCents > cats[j].ExpenseCents
		}
		return cats[i].CategoryName < cats[j].CategoryName
	})
}
