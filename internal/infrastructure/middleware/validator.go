package middleware

import (
	"carowebapp/core/internal/infrastructure/logger"

	"github.com/go-playground/validator/v10"

	"github.com/gofiber/fiber/v2"

	"go.uber.org/zap"
)

var validate = validator.New()

func ValidateBody[T any]() fiber.Handler {
	return func(c *fiber.Ctx) error {
		payload := new(T)

		if err := c.BodyParser(payload); err != nil {
			logger.Log.Warn("Failed to parse request body",
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.Error(err),
			)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "invalid input format",
			})
		}

		if err := validate.Struct(payload); err != nil {
			logger.Log.Warn("Validation failed",
				zap.String("path", c.Path()),
				zap.String("method", c.Method()),
				zap.Any("payload", payload),
				zap.Error(err),
			)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "validation failed",
				"details": err.Error(),
			})
		}

		// сохраняем *T для удобного кастинга
		c.Locals("validatedBody", payload)
		return c.Next()
	}
}
