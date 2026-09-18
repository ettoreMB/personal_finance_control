package server

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/ledger"
)

type entryRequest struct {
	Type        string `json:"type"`
	AmountCents int    `json:"amount_cents"`
	EntryDate   string `json:"entry_date"`
	CategoryID  uint   `json:"category_id"`
	PurchaseID  *uint  `json:"purchase_id"`
}

type nestedCategoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type nestedPurchaseResponse struct {
	ID               uint   `json:"id"`
	Description      string `json:"description"`
	InstallmentCount int    `json:"installment_count"`
}

type entryResponse struct {
	ID                uint                     `json:"id"`
	Type              string                   `json:"type"`
	AmountCents       int                      `json:"amount_cents"`
	EntryDate         string                   `json:"entry_date"`
	CategoryID        uint                     `json:"category_id"`
	Category          *nestedCategoryResponse  `json:"category,omitempty"`
	PurchaseID        *uint                    `json:"purchase_id,omitempty"`
	InstallmentNumber *int                     `json:"installment_number,omitempty"`
	Purchase          *nestedPurchaseResponse  `json:"purchase,omitempty"`
}

func toEntryResponse(entry ledger.Entry) entryResponse {
	resp := entryResponse{
		ID:                entry.ID,
		Type:              entry.Type,
		AmountCents:       entry.AmountCents,
		EntryDate:         ledger.FormatEntryDate(entry.EntryDate),
		CategoryID:        entry.CategoryID,
		PurchaseID:        entry.PurchaseID,
		InstallmentNumber: entry.InstallmentNumber,
	}
	if entry.Category.ID != 0 {
		resp.Category = &nestedCategoryResponse{ID: entry.Category.ID, Name: entry.Category.Name}
	}
	if entry.Purchase != nil && entry.Purchase.ID != 0 {
		resp.Purchase = &nestedPurchaseResponse{
			ID:               entry.Purchase.ID,
			Description:      entry.Purchase.Description,
			InstallmentCount: entry.Purchase.InstallmentCount,
		}
	}
	return resp
}

func listEntriesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var entries []ledger.Entry
		if err := db.Preload("Category").Preload("Purchase").Order("entry_date DESC, id DESC").Find(&entries).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to list entries")
		}

		out := make([]entryResponse, 0, len(entries))
		for _, entry := range entries {
			out = append(out, toEntryResponse(entry))
		}
		return c.JSON(out)
	}
}

func createEntryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req entryRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
		if req.PurchaseID != nil {
			return fiber.NewError(fiber.StatusBadRequest, ledger.ErrPurchaseIDNotAllowed.Error())
		}

		entry, err := buildEntry(db, req)
		if err != nil {
			return mapEntryError(err)
		}

		if err := db.Create(&entry).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to create entry")
		}
		if err := db.Preload("Category").First(&entry, entry.ID).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to load entry")
		}

		return c.Status(fiber.StatusCreated).JSON(toEntryResponse(entry))
	}
}

func getEntryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		entry, err := findEntry(db, c)
		if err != nil {
			return err
		}
		return c.JSON(toEntryResponse(entry))
	}
}

type entryPatchRequest struct {
	Type        *string `json:"type"`
	AmountCents *int    `json:"amount_cents"`
	EntryDate   *string `json:"entry_date"`
	CategoryID  *uint   `json:"category_id"`
}

func updateEntryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		entry, err := findEntry(db, c)
		if err != nil {
			return err
		}
		if entry.PurchaseID != nil {
			return fiber.NewError(fiber.StatusConflict, ledger.ErrParcelaImmutable.Error())
		}

		var req entryPatchRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		if req.Type != nil {
			entryType, err := ledger.ParseEntryType(*req.Type)
			if err != nil {
				return mapEntryError(err)
			}
			entry.Type = entryType
		}
		if req.AmountCents != nil {
			amount, err := ledger.ParseAmountCents(*req.AmountCents)
			if err != nil {
				return mapEntryError(err)
			}
			entry.AmountCents = amount
		}
		if req.EntryDate != nil {
			date, err := ledger.ParseEntryDate(*req.EntryDate)
			if err != nil {
				return mapEntryError(err)
			}
			entry.EntryDate = date
		}
		if req.CategoryID != nil {
			if *req.CategoryID == 0 {
				return mapEntryError(ledger.ErrCategoryRequired)
			}
			var category ledger.Category
			if err := db.First(&category, *req.CategoryID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return mapEntryError(ledger.ErrCategoryNotFound)
				}
				return fiber.NewError(fiber.StatusInternalServerError, "failed to look up category")
			}
			entry.CategoryID = category.ID
			entry.Category = category
		}

		if err := db.Save(&entry).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to update entry")
		}
		if err := db.Preload("Category").First(&entry, entry.ID).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to load entry")
		}

		return c.JSON(toEntryResponse(entry))
	}
}

func deleteEntryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		entry, err := findEntry(db, c)
		if err != nil {
			return err
		}
		if entry.PurchaseID != nil {
			return fiber.NewError(fiber.StatusConflict, ledger.ErrParcelaImmutable.Error())
		}

		if err := db.Delete(&entry).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to delete entry")
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

func findEntry(db *gorm.DB, c *fiber.Ctx) (ledger.Entry, error) {
	id, err := parseIDParam(c)
	if err != nil {
		return ledger.Entry{}, fiber.NewError(fiber.StatusNotFound, "entry not found")
	}

	var entry ledger.Entry
	if err := db.Preload("Category").Preload("Purchase").First(&entry, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ledger.Entry{}, fiber.NewError(fiber.StatusNotFound, "entry not found")
		}
		return ledger.Entry{}, fiber.NewError(fiber.StatusInternalServerError, "failed to look up entry")
	}

	return entry, nil
}

func buildEntry(db *gorm.DB, req entryRequest) (ledger.Entry, error) {
	entryType, err := ledger.ParseEntryType(req.Type)
	if err != nil {
		return ledger.Entry{}, err
	}
	amount, err := ledger.ParseAmountCents(req.AmountCents)
	if err != nil {
		return ledger.Entry{}, err
	}
	date, err := ledger.ParseEntryDate(req.EntryDate)
	if err != nil {
		return ledger.Entry{}, err
	}
	if req.CategoryID == 0 {
		return ledger.Entry{}, ledger.ErrCategoryRequired
	}

	var category ledger.Category
	if err := db.First(&category, req.CategoryID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ledger.Entry{}, ledger.ErrCategoryNotFound
		}
		return ledger.Entry{}, err
	}

	return ledger.Entry{
		Type:        entryType,
		AmountCents: amount,
		EntryDate:   date,
		CategoryID:  category.ID,
		Category:    category,
	}, nil
}

func mapEntryError(err error) error {
	switch {
	case errors.Is(err, ledger.ErrInvalidEntryType),
		errors.Is(err, ledger.ErrInvalidAmount),
		errors.Is(err, ledger.ErrInvalidEntryDate),
		errors.Is(err, ledger.ErrCategoryRequired),
		errors.Is(err, ledger.ErrCategoryNotFound):
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "failed to process entry")
	}
}
