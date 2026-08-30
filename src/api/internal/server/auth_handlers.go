package server

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/auth"
)

type registerRequest struct {
	Email           string `json:"email"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
	CPF             string `json:"cpf"`
}

type userResponse struct {
	ID    uint   `json:"id"`
	Email string `json:"email"`
	CPF   string `json:"cpf"`
}

func registerHandler(db *gorm.DB) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req registerRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		if req.Password != req.ConfirmPassword {
			return fiber.NewError(fiber.StatusBadRequest, "passwords do not match")
		}

		if err := auth.ValidatePassword(req.Password); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		cpf, err := auth.NormalizeCPF(req.CPF)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}

		email := strings.TrimSpace(req.Email)
		if email == "" {
			return fiber.NewError(fiber.StatusBadRequest, "email is required")
		}

		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to hash password")
		}

		user := auth.User{
			Email:        email,
			PasswordHash: hash,
			CPF:          cpf,
		}

		if err := db.Create(&user).Error; err != nil {
			if isUniqueConstraintErr(err) {
				return fiber.NewError(fiber.StatusConflict, "email or cpf already registered")
			}
			return fiber.NewError(fiber.StatusInternalServerError, "failed to create user")
		}

		return c.Status(fiber.StatusCreated).JSON(userResponse{
			ID:    user.ID,
			Email: user.Email,
			CPF:   user.CPF,
		})
	}
}

func isUniqueConstraintErr(err error) bool {
	return errors.Is(err, gorm.ErrDuplicatedKey) || strings.Contains(err.Error(), "UNIQUE constraint failed")
}
