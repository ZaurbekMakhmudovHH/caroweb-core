package servicecard

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func RegisterRoutes(app fiber.Router, service *Service, logger *zap.Logger) {
	handler := NewHandler(service, logger)
	routes := app.Group("/tickets")

	routes.Post("/create", handler.Create)
	routes.Get("/:id", handler.Get)
	routes.Get("/", handler.List)
	routes.Put("/:id", handler.Update)
}
