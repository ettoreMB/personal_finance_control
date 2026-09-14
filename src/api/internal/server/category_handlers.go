package server

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/ledger"
)

type categoryResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type categoryRequest struct {
	Name string `json:"name"`
}

func listCategoriesHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var categories []ledger.Category
		if err := db.Order("name").Find(&categories).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to list categories")
		}

		out := make([]categoryResponse, 0, len(categories))
		for _, cat := range categories {
			out = append(out, categoryResponse{ID: cat.ID, Name: cat.Name})
		}

		return c.JSON(out)
	}
}

func createCategoryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req categoryRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		name, err := ledger.NormalizeCategoryName(req.Name)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		category := ledger.Category{Name: name}
		if err := db.Create(&category).Error; err != nil {
			if isUniqueConstraintErr(err) {
				return fiber.NewError(fiber.StatusConflict, "category name already exists")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to create category")
		}

		return c.Status(fiber.StatusCreated).JSON(categoryResponse{ID: category.ID, Name: category.Name})
	}
}

func updateCategoryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := parseIDParam(c)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "category not found")
		}

		var req categoryRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		name, err := ledger.NormalizeCategoryName(req.Name)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		var category ledger.Category
		if err := db.First(&category, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "category not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to look up category")
		}

		category.Name = name
		if err := db.Save(&category).Error; err != nil {
			if isUniqueConstraintErr(err) {
				return fiber.NewError(fiber.StatusConflict, "category name already exists")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to update category")
		}

		return c.JSON(categoryResponse{ID: category.ID, Name: category.Name})
	}
}

func deleteCategoryHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id, err := parseIDParam(c)
		if err != nil {
			return fiber.NewError(fiber.StatusNotFound, "category not found")
		}

		var category ledger.Category
		if err := db.First(&category, id).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fiber.NewError(fiber.StatusNotFound, "category not found")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to look up category")
		}

		var activeCount int64
		if err := db.Model(&ledger.Entry{}).Where("category_id = ?", id).Count(&activeCount).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to check category usage")
		}
		if activeCount > 0 {
			return fiber.NewError(fiber.StatusConflict, "category is in use")
		}

		result := db.Delete(&category)
		if result.Error != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to delete category")
		}
		if result.RowsAffected == 0 {
			return fiber.NewError(fiber.StatusNotFound, "category not found")
		}

		return c.SendStatus(fiber.StatusNoContent)
	}
}

func parseIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return 0, errInvalidID
	}
	return uint(id), nil
}

var errInvalidID = errors.New("invalid id")
