package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/auth"
	"github.com/ettoreMB/personal_finance_control/api/internal/config"
)

const currentUserLocalsKey = "currentUser"

func requireSession(db *gorm.DB, cfg config.Config) fiber.Handler {
	return func(c *fiber.Ctx) error {
		token := c.Cookies(sessionCookieName)
		if token == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		var session auth.Session
		if err := db.Where("token_hash = ? AND expires_at > ?", auth.HashSessionToken(token), time.Now()).
			First(&session).Error; err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		var user auth.User
		if err := db.First(&user, session.UserID).Error; err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
		}

		session.ExpiresAt = time.Now().Add(cfg.SessionTTL)
		if err := db.Model(&session).Update("expires_at", session.ExpiresAt).Error; err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "failed to extend session")
		}

		c.Locals(currentUserLocalsKey, user)

		return c.Next()
	}
}

func currentUser(c *fiber.Ctx) (auth.User, bool) {
	user, ok := c.Locals(currentUserLocalsKey).(auth.User)
	return user, ok
}
