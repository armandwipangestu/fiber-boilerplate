package server

import (
	"github.com/gofiber/fiber/v2"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
	"github.com/armandwipangestu/fiber-boilerplate/internal/middleware"
)

// New builds and configures the Fiber application with core middleware.
func New(cfg config.Config) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:           cfg.AppName,
		EnablePrintRoutes: false,
		BodyLimit:         10 * 1024 * 1024, // 10MB
	})

	app.Use(middleware.RequestID())

	api := app.Group("/api")
	v1 := api.Group("/v1")

	// Health & liveness
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.JSON(map[string]any{"message": "pong"})
	})

	_ = v1

	return app
}
