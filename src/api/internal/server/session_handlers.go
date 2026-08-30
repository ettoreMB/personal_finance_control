package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/auth"
	"github.com/ettoreMB/personal_finance_control/api/internal/config"
)

const sessionCookieName = "session"

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func loginHandler(db *gorm.DB, cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		var req loginRequest
		if err := c.BodyParser(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
		}

		var user auth.User
		if err := db.Where("email = ?", req.Email).First(&user).Error; err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}

		if !auth.ComparePassword(user.PasswordHash, req.Password) {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid credentials")
		}

		token, err := auth.GenerateSessionToken()
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to create session")
		}

		expiresAt := time.Now().Add(cfg.SessionTTL)

		session := auth.Session{
			UserID:    user.ID,
			TokenHash: auth.HashSessionToken(token),
			ExpiresAt: expiresAt,
		}

		if err := db.Create(&session).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to create session")
		}

		c.Cookie(&fiber.Cookie{
			Name:     sessionCookieName,
			Value:    token,
			Expires:  expiresAt,
			HTTPOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: fiber.CookieSameSiteLaxMode,
			Path:     "/",
		})

		return c.JSON(userResponse{ID: user.ID, Email: user.Email, CPF: user.CPF})
	}
}

func logoutHandler(db *gorm.DB, cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies(sessionCookieName)
		if token != "" {
			db.Where("token_hash = ?", auth.HashSessionToken(token)).Delete(&auth.Session{})
		}

		c.Cookie(&fiber.Cookie{
			Name:     sessionCookieName,
			Value:    "",
			Expires:  time.Now().Add(-time.Hour),
			HTTPOnly: true,
			Secure:   cfg.CookieSecure,
			SameSite: fiber.CookieSameSiteLaxMode,
			Path:     "/",
		})

		return c.SendStatus(fiber.StatusOK)
	}
}
