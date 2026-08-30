package server

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/auth"
	"github.com/ettoreMB/personal_finance_control/api/internal/config"
)

type recoveryRequest struct {
	RecoverySecret string `json:"recovery_secret"`
	NewPassword    string `json:"new_password"`
}

func recoveryHandler(db *gorm.DB, cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req recoveryRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		if cfg.RecoverySecret == "" || req.RecoverySecret != cfg.RecoverySecret {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid recovery secret")
		}

		if err := auth.ValidatePassword(req.NewPassword); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		var users []auth.User
		if err := db.Find(&users).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to look up users")
		}

		if len(users) != 1 {
			return fiber.NewError(fiber.StatusConflict, "password recovery requires exactly one existing user")
		}

		hash, err := auth.HashPassword(req.NewPassword)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to hash password")
		}

		if err := db.Model(&users[0]).Update("password_hash", hash).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to update password")
		}

		return c.SendStatus(fiber.StatusOK)
	}
}
