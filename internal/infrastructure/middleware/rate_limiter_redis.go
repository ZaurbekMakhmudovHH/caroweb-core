package middleware

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
	"golang.org/x/net/context"
)

var ctx = context.Background()

type RateLimiterConfig struct {
	RedisClient  *redis.Client
	MaxAttempts  int
	Window       time.Duration
	KeyGenerator func(*fiber.Ctx) string
	Logger       *zap.Logger
}

func RateLimitByRedis(cfg RateLimiterConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		key := fmt.Sprintf("rl:%s", cfg.KeyGenerator(c))

		attempts, err := cfg.RedisClient.Get(ctx, key).Int()
		if err != nil && !errors.Is(err, redis.Nil) {
			cfg.Logger.Error("Redis error during rate limiting",
				zap.String("key", key),
				zap.Error(err),
			)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Rate limiter error",
			})
		}

		if attempts >= cfg.MaxAttempts {
			cfg.Logger.Warn("Rate limit exceeded",
				zap.String("key", key),
				zap.Int("attempts", attempts),
				zap.Duration("window", cfg.Window),
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
			)
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error": "Too many attempts. Please try again later.",
			})
		}

		pipe := cfg.RedisClient.TxPipeline()
		pipe.Incr(ctx, key)
		pipe.Expire(ctx, key, cfg.Window)
		_, _ = pipe.Exec(ctx)

		cfg.Logger.Info("Rate limit attempt recorded",
			zap.String("key", key),
			zap.Int("current_attempts", attempts+1),
			zap.String("ip", c.IP()),
			zap.String("path", c.Path()),
		)

		return c.Next()
	}
}

func ResetPasswordLimiter(redisClient *redis.Client, logger *zap.Logger) fiber.Handler {
	const (
		maxAttempts   = 5
		shortWindow   = 1 * time.Minute
		blockDuration = 1 * time.Hour
	)

	return func(c *fiber.Ctx) error {
		var body struct {
			Email string `json:"email"`
		}
		if err := c.BodyParser(&body); err != nil || body.Email == "" {
			logger.Info("Missing or invalid email in reset-password request",
				zap.String("ip", c.IP()),
				zap.String("path", c.Path()),
			)
			return fiber.NewError(fiber.StatusBadRequest, "missing email")
		}

		email := body.Email
		keyAttempts := fmt.Sprintf("rl:reset:%s:count", email)
		keyBlock := fmt.Sprintf("rl:reset:%s:block", email)

		// Проверка блокировки
		blocked, err := redisClient.Exists(ctx, keyBlock).Result()
		if err != nil {
			logger.Error("Redis error checking block key", zap.Error(err))
			return fiber.ErrInternalServerError
		}
		if blocked > 0 {
			logger.Info("Password reset blocked due to rate limiting",
				zap.String("email", email),
				zap.String("ip", c.IP()),
			)
			return fiber.NewError(fiber.StatusTooManyRequests, "Too many attempts, try again later.")
		}

		// Инкремент попытки
		attempts, err := redisClient.Incr(ctx, keyAttempts).Result()
		if err != nil {
			logger.Error("Redis error incrementing attempts", zap.Error(err))
			return fiber.ErrInternalServerError
		}

		if attempts == 1 {
			_ = redisClient.Expire(ctx, keyAttempts, shortWindow).Err()
		}

		if attempts > maxAttempts {
			_ = redisClient.Set(ctx, keyBlock, 1, blockDuration).Err()
			_ = redisClient.Del(ctx, keyAttempts).Err()

			logger.Info("Password reset rate limit exceeded",
				zap.String("email", email),
				zap.String("ip", c.IP()),
				zap.Int64("attempts", attempts),
			)

			return fiber.NewError(fiber.StatusTooManyRequests, "Too many attempts, try again in 1 hour.")
		}

		logger.Info("Password reset attempt recorded",
			zap.String("email", email),
			zap.String("ip", c.IP()),
			zap.Int64("attempt", attempts),
		)

		return c.Next()
	}
}
