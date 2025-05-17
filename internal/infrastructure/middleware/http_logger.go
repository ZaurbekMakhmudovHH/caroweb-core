package middleware

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func HTTPLogger(log *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()
		method := c.Method()
		path := c.OriginalURL()
		userID, _ := c.Locals("userID").(string)

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("duration", duration),
		}
		if userID != "" {
			fields = append(fields, zap.String("user_id", userID))
		}
		if err != nil {
			fields = append(fields, zap.Error(err))
			log.Error("request failed", fields...)
		} else {
			log.Info("request handled", fields...)
		}

		return err
	}
}
