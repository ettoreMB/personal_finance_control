package server

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/config"
)

func New(db *gorm.DB, cfg config.Config) *fiber.App {
	app := fiber.New()

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Post("/register", registerHandler(db))
	app.Post("/login", loginHandler(db, cfg))
	app.Post("/auth/recovery", recoveryHandler(db, cfg))

	app.Use(requireSession(db, cfg))

	app.Post("/logout", logoutHandler(db, cfg))
	app.Get("/me", meHandler)

	return app
}
