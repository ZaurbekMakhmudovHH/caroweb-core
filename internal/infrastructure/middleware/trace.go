package middleware

import (
	"carowebapp/core/internal/pkg/contextutils"

	"github.com/gofiber/fiber/v2"

	"github.com/google/uuid"
)

func TraceID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		traceID := uuid.New().String()
		c.Locals(contextutils.ContextKeyTraceID, traceID)
		c.Set("X-Trace-ID", traceID)
		return c.Next()
	}
}
