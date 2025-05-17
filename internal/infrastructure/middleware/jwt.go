package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"

	"github.com/golang-jwt/jwt/v5"

	"strings"
)

func JWTMiddleware(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "missing or invalid token"})
		}

		token := strings.TrimPrefix(auth, "Bearer ")
		claims, err := ParseJWT(token, secret)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "invalid or expired token"})
		}

		userID := claims["sub"].(string)
		c.Locals("userID", userID)

		return c.Next()
	}
}

func ParseJWT(tokenStr, secret string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(_ *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})

	if err != nil || !token.Valid {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}
