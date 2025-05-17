package routes

import (
	"carowebapp/core/internal/features/auth"

	"carowebapp/core/internal/infrastructure/middleware"

	"github.com/go-redis/redis/v8"

	"github.com/gofiber/fiber/v2"

	"go.uber.org/zap"

	"os"

	"time"
)

// RegisterAuthRoutes sets up all auth-related routes under /api/v1/auth and /api/v1/authenticated.
func RegisterAuthRoutes(app *fiber.App, service *auth.Service, logger *zap.Logger, redis *redis.Client) {
	handler := auth.NewHandler(service, logger)

	// --- Public routes: /api/v1/auth
	public := app.Group("/api/v1/auth")

	loginLimiter := middleware.RateLimitByRedis(middleware.RateLimiterConfig{
		RedisClient: redis,
		MaxAttempts: 5,
		Window:      15 * time.Minute,
		Logger:      logger,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	})

	public.Post("/login",
		loginLimiter,
		middleware.ValidateBody[auth.LoginRequest](),
		handler.Login,
	)

	public.Post("/register",
		loginLimiter,
		middleware.ValidateBody[auth.RegisterRequest](),
		handler.Register,
	)

	public.Post("/reset-password-request",
		middleware.ValidateBody[auth.ResetPasswordRequest](),
		middleware.ResetPasswordLimiter(redis, logger),
		handler.RequestResetPassword,
	)

	public.Get("/reset-password-check-token",
		handler.CheckResetPasswordToken,
	)

	public.Post("/reset-password",
		middleware.ResetPasswordLimiter(redis, logger),
		middleware.ValidateBody[auth.ResetPasswordPayload](),
		handler.ResetPassword,
	)

	public.Get("/confirm",
		handler.ConfirmEmail,
	)

	// --- Protected routes: /api/v1/auth
	protected := app.Group("/api/v1")
	protected.Use(middleware.JWTMiddleware(os.Getenv("JWT_SECRET")))

	authProtected := protected.Group("/auth")

	authProtected.Post("/resend-confirmation",
		handler.ResendConfirmation,
	)

	authProtected.Post("/create-profile",
		middleware.ValidateBody[auth.CreateProfileRequest](),
		handler.CreateProfile,
	)
}
