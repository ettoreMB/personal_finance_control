package server

import "github.com/gofiber/fiber/v2"

func meHandler(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return fiber.NewError(fiber.StatusUnauthorized, "authentication required")
	}

	return c.JSON(userResponse{ID: user.ID, Email: user.Email, CPF: user.CPF})
}
