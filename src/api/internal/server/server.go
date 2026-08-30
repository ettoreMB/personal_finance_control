package server

import "github.com/gofiber/fiber/v2"

func New() *fiber.App {
	app := fiber.New()

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	return app
}
