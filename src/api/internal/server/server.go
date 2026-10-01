package server

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"

	"github.com/ettoreMB/personal_finance_control/api/internal/config"
	"github.com/ettoreMB/personal_finance_control/api/internal/whatsapp"
)

func New(db *gorm.DB, cfg config.Config) *fiber.App {
	return NewWith(db, cfg, Dependencies{
		Sender:     whatsapp.NewEvolutionSender(cfg.EvolutionBaseURL, cfg.EvolutionAPIKey, cfg.EvolutionInstanceID),
		Classifier: whatsapp.NewJevClassifier(cfg.JevAPIKey, cfg.JevModel),
		Now:        time.Now,
	})
}

func NewWith(db *gorm.DB, cfg config.Config, deps Dependencies) *fiber.App {
	app := fiber.New()

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Post("/webhooks/whatsapp", whatsappWebhook(db, cfg, deps))

	app.Post("/register", registerHandler(db))
	app.Post("/login", loginHandler(db, cfg))
	app.Post("/auth/recovery", recoveryHandler(db, cfg))

	app.Use(requireSession(db, cfg))

	app.Post("/logout", logoutHandler(db, cfg))
	app.Get("/me", meHandler)

	app.Get("/categories", listCategoriesHandler(db))
	app.Post("/categories", createCategoryHandler(db))
	app.Patch("/categories/:id", updateCategoryHandler(db))
	app.Delete("/categories/:id", deleteCategoryHandler(db))

	app.Get("/entries", listEntriesHandler(db))
	app.Post("/entries", createEntryHandler(db))
	app.Get("/entries/:id", getEntryHandler(db))
	app.Patch("/entries/:id", updateEntryHandler(db))
	app.Delete("/entries/:id", deleteEntryHandler(db))

	app.Get("/summary", summaryHandler(db))

	app.Get("/purchases", listPurchasesHandler(db))
	app.Post("/purchases", createPurchaseHandler(db))
	app.Get("/purchases/:id", getPurchaseHandler(db))
	app.Patch("/purchases/:id", updatePurchaseHandler(db))
	app.Delete("/purchases/:id", deletePurchaseHandler(db))

	return app
}
