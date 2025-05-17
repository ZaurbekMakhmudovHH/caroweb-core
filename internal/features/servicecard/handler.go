package servicecard

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type Handler struct {
	service *Service
	logger  *zap.Logger
}

func NewHandler(service *Service, logger *zap.Logger) *Handler {
	return &Handler{service: service, logger: logger}
}

func (h *Handler) Create(c *fiber.Ctx) error {
	return c.SendString("Create Ticket (to be implemented)")
}

func (h *Handler) Get(c *fiber.Ctx) error {
	return c.SendString("Get Ticket (to be implemented)")
}

func (h *Handler) List(c *fiber.Ctx) error {
	return c.SendString("List Tickets (to be implemented)")
}

func (h *Handler) Update(c *fiber.Ctx) error {
	return c.SendString("Update Ticket (to be implemented)")
}
