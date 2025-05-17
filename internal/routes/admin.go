package routes

import (
	"carowebapp/core/internal/features/admin"

	"carowebapp/core/internal/infrastructure/adapter"

	"carowebapp/core/internal/infrastructure/middleware"

	"github.com/gofiber/fiber/v2"

	"go.uber.org/zap"

	"os"
)

// RegisterAdminRoutes sets up admin-specific endpoints under /api/v1/admin.
func RegisterAdminRoutes(app *fiber.App, service *admin.Service, repo admin.Repository, logger *zap.Logger) {
	userProvider := &adapter.AdminUserProvider{Repo: repo}

	handler := &admin.Handler{
		Service:      service,
		Logger:       logger,
		UserProvider: userProvider,
	}

	adminGroup := app.Group("/api/v1/admin")

	adminGroup.Use(
		middleware.JWTMiddleware(os.Getenv("JWT_SECRET")),
		middleware.RequireAdmin(userProvider, logger),
	)

	adminGroup.Post("/approve-user",
		middleware.ValidateBody[admin.ApproveUserRequest](),
		handler.ApproveUser,
	)

	adminGroup.Post("/reject-user",
		middleware.ValidateBody[admin.RejectUserRequest](),
		handler.RejectUser,
	)

	adminGroup.Get("/pending-users", handler.ListPendingUsers)

	adminGroup.Get("/user-profile/:id", handler.GetUserProfile)
}
