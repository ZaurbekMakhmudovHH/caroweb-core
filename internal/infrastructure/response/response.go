package response

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func JSONError(c *fiber.Ctx, status int, err error) error {
	return c.Status(status).JSON(fiber.Map{"error": err.Error()})
}

func JSONErrorWithLog(c *fiber.Ctx, logger *zap.Logger, status int, msg string, fields ...zap.Field) error {
	logger.Error(msg, fields...)
	return c.Status(status).JSON(fiber.Map{"error": msg})
}

func JSONErrorInfoLog(c *fiber.Ctx, logger *zap.Logger, status int, msg string, fields ...zap.Field) error {
	logger.Info(msg, fields...)
	return c.Status(status).JSON(fiber.Map{"error": msg})
}

func JSONSuccess(c *fiber.Ctx, status int, payload interface{}) error {
	return c.Status(status).JSON(payload)
}
