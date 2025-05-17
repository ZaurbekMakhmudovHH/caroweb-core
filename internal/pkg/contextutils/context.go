package contextutils

import (
	"github.com/gofiber/fiber/v2"
)

const (
	ContextKeyUserID        = "userID"
	ContextKeyValidatedBody = "validatedBody"
	ContextKeyTraceID       = "traceID"
)

func GetUserID(c *fiber.Ctx) (string, bool) {
	id, ok := c.Locals(ContextKeyUserID).(string)
	return id, ok && id != ""
}

func GetValidatedBody[T any](c *fiber.Ctx) (*T, bool) {
	body, ok := c.Locals(ContextKeyValidatedBody).(*T)
	return body, ok
}

func GetTraceID(c *fiber.Ctx) (string, bool) {
	traceID, ok := c.Locals(ContextKeyTraceID).(string)
	return traceID, ok && traceID != ""
}
