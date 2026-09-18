package server

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/ledger"
)

type purchaseRequest struct {
	Description      string `json:"description"`
	PurchaseDate     string `json:"purchase_date"`
	AmountCents      int    `json:"amount_cents"`
	InstallmentCount int    `json:"installment_count"`
	CategoryID       uint   `json:"category_id"`
}

type purchaseInstallmentResponse struct {
	ID                uint   `json:"id"`
	InstallmentNumber int    `json:"installment_number"`
	AmountCents       int    `json:"amount_cents"`
	EntryDate         string `json:"entry_date"`
	CategoryID        uint   `json:"category_id"`
}

type purchaseResponse struct {
	ID               uint                           `json:"id"`
	Description      string                         `json:"description"`
	PurchaseDate     string                         `json:"purchase_date"`
	AmountCents      int                            `json:"amount_cents"`
	InstallmentCount int                            `json:"installment_count"`
	CategoryID       uint                           `json:"category_id"`
	CategoryName     string                         `json:"category_name"`
	Category         *nestedCategoryResponse        `json:"category,omitempty"`
	Installments     []purchaseInstallmentResponse  `json:"installments"`
}

func toPurchaseResponse(purchase ledger.Purchase) purchaseResponse {
	resp := purchaseResponse{
		ID:               purchase.ID,
		Description:      purchase.Description,
		PurchaseDate:     ledger.FormatEntryDate(purchase.PurchaseDate),
		AmountCents:      purchase.AmountCents,
		InstallmentCount: purchase.InstallmentCount,
		CategoryID:       purchase.CategoryID,
		CategoryName:     purchase.CategoryName,
		Installments:     make([]purchaseInstallmentResponse, 0, len(purchase.Entries)),
	}
	if purchase.Category.ID != 0 {
		resp.Category = &nestedCategoryResponse{ID: purchase.Category.ID, Name: purchase.Category.Name}
	}
	for _, entry := range purchase.Entries {
		number := 0
		if entry.InstallmentNumber != nil {
			number = *entry.InstallmentNumber
		}
		resp.Installments = append(resp.Installments, purchaseInstallmentResponse{
			ID:                entry.ID,
			InstallmentNumber: number,
			AmountCents:       entry.AmountCents,
			EntryDate:         ledger.FormatEntryDate(entry.EntryDate),
			CategoryID:        entry.CategoryID,
		})
	}
	return resp
}

func loadPurchase(db *gorm.DB, id uint) (ledger.Purchase, error) {
	var purchase ledger.Purchase
	err := db.Preload("Category").Preload("Entries", func(tx *gorm.DB) *gorm.DB {
		return tx.Order("installment_number ASC")
	}).First(&purchase, id).Error
	return purchase, err
}

func listPurchasesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var purchases []ledger.Purchase
		if err := db.Preload("Category").Preload("Entries", func(tx *gorm.DB) *gorm.DB {
			return tx.Order("installment_number ASC")
		}).Order("purchase_date DESC, id DESC").Find(&purchases).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to list purchases")
		}

		out := make([]purchaseResponse, 0, len(purchases))
		for _, purchase := range purchases {
			out = append(out, toPurchaseResponse(purchase))
		}
		return c.JSON(out)
	}
}

func getPurchaseHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := parseIDParam(c)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "purchase not found")
		}

		purchase, err := loadPurchase(db, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "purchase not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to load purchase")
		}
		return c.JSON(toPurchaseResponse(purchase))
	}
}

func createPurchaseHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req purchaseRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		description, err := ledger.NormalizeDescription(req.Description)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		n, err := ledger.ParseInstallmentCount(req.InstallmentCount)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		amount, err := ledger.ParsePurchaseAmount(req.AmountCents, n)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		date, err := ledger.ParseEntryDate(req.PurchaseDate)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		if req.CategoryID == 0 {
			return mapEntryError(ledger.ErrCategoryRequired)
		}

		var category ledger.Category
		if err := db.First(&category, req.CategoryID).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return mapEntryError(ledger.ErrCategoryNotFound)
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to look up category")
		}

		purchase := ledger.Purchase{
			Description:      description,
			PurchaseDate:     date,
			AmountCents:      amount,
			InstallmentCount: n,
			CategoryID:       category.ID,
			CategoryName:     category.Name,
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Create(&purchase).Error; err != nil {
				return err
			}
			entries := ledger.GenerateParcelas(purchase)
			if err := tx.Create(&entries).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to create purchase")
		}

		purchase, err = loadPurchase(db, purchase.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to load purchase")
		}
		return c.Status(fiber.StatusCreated).JSON(toPurchaseResponse(purchase))
	}
}

type purchasePatchRequest struct {
	Description      *string `json:"description"`
	AmountCents      *int    `json:"amount_cents"`
	CategoryID       *uint   `json:"category_id"`
	InstallmentCount *int    `json:"installment_count"`
	PurchaseDate     *string `json:"purchase_date"`
}

func updatePurchaseHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := parseIDParam(c)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "purchase not found")
		}

		var req purchasePatchRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}
		if req.InstallmentCount != nil || req.PurchaseDate != nil {
			return fiber.NewError(fiber.StatusBadRequest, ledger.ErrPurchaseFieldImmutable.Error())
		}

		purchase, err := loadPurchase(db, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "purchase not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to load purchase")
		}

		if req.Description != nil {
			description, err := ledger.NormalizeDescription(*req.Description)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
			purchase.Description = description
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
			purchase.CategoryID = category.ID
			purchase.CategoryName = category.Name
			purchase.Category = category
			for i := range purchase.Entries {
				purchase.Entries[i].CategoryID = category.ID
				purchase.Entries[i].Category = category
			}
		}

		if req.AmountCents != nil {
			amount, err := ledger.ParsePurchaseAmount(*req.AmountCents, purchase.InstallmentCount)
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, err.Error())
			}
			if err := ledger.RedistributeRemaining(purchase.Entries, amount, ledger.TodayCivil()); err != nil {
				return fiber.NewError(fiber.StatusConflict, err.Error())
			}
			purchase.AmountCents = amount
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			if err := tx.Omit("Entries", "Category").Save(&purchase).Error; err != nil {
				return err
			}
			for i := range purchase.Entries {
				if err := tx.Omit("Category", "Purchase").Save(&purchase.Entries[i]).Error; err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to update purchase")
		}

		purchase, err = loadPurchase(db, purchase.ID)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to load purchase")
		}
		return c.JSON(toPurchaseResponse(purchase))
	}
}

func deletePurchaseHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := parseIDParam(c)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "purchase not found")
		}

		purchase, err := loadPurchase(db, id)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "purchase not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to load purchase")
		}

		if ledger.HasPastParcela(purchase.Entries, ledger.TodayCivil()) {
			return fiber.NewError(fiber.StatusConflict, ledger.ErrCannotUndoPurchase.Error())
		}

		err = db.Transaction(func(tx *gorm.DB) error {
			for i := range purchase.Entries {
				if err := tx.Delete(&purchase.Entries[i]).Error; err != nil {
					return err
				}
			}
			return tx.Delete(&purchase).Error
		})
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to delete purchase")
		}
		return c.SendStatus(fiber.StatusNoContent)
	}
}
