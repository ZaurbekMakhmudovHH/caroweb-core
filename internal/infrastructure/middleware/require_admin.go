package middleware

import (
	"carowebapp/core/internal/domain/user"

	"carowebapp/core/internal/infrastructure/response"

	"carowebapp/core/internal/pkg/contextutils"

	"github.com/gofiber/fiber/v2"

	"go.uber.org/zap"
)

// RequireAdmin checks if the current user is an admin.
// It expects JWT middleware to already have set the user ID in context.
func RequireAdmin(provider user.Provider, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := contextutils.GetUserID(c)
		if !ok {
			_ = response.JSONErrorInfoLog(c, logger, fiber.StatusUnauthorized, response.ErrMsgUnauthorized)
			return fiber.ErrUnauthorized
		}

		u, err := provider.GetByID(c.Context(), userID)
		if err != nil {
			_ = response.JSONErrorInfoLog(c, logger, fiber.StatusInternalServerError, "failed to fetch user",
				zap.String("user_id", userID),
				zap.Error(err),
			)
			return fiber.ErrInternalServerError
		}
		if u == nil || !u.IsAdmin() {
			_ = response.JSONErrorInfoLog(c, logger, fiber.StatusForbidden, "only admin can perform this action",
				zap.String("user_id", userID),
				zap.String("email", u.Email),
				zap.String("role", u.Role),
			)
			return fiber.ErrForbidden
		}

		return c.Next()
	}
}
