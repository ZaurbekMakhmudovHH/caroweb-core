package main

import (
	"context"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"carowebapp/core/cmd"
	"carowebapp/core/internal/features/admin"
	"carowebapp/core/internal/features/auth"
	database "carowebapp/core/internal/infrastructure/db"
	"carowebapp/core/internal/infrastructure/db/migrations"
	"carowebapp/core/internal/infrastructure/email"
	"carowebapp/core/internal/infrastructure/logger"
	"carowebapp/core/internal/infrastructure/middleware"
	"carowebapp/core/internal/routes"
)

// rootCmd is the main command for the GreenCard CLI application.
var rootCmd = &cobra.Command{
	Use:   "GreenCard",
	Short: "GreenCard CLI Application",
}

// initCLI initializes the CLI commands using Cobra.
func initCLI() {
	rootCmd.AddCommand(cmd.CreateAdminCmd)
}

// initRedis initializes a Redis client using the specified configuration.
func initRedis() *redis.Client {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	redisURL := os.Getenv("REDIS_ADDR")
	if redisURL == "" {
		logger.Log.Fatal("REDIS_ADDR is not set")
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:     redisURL,
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Log.Fatal("failed to connect to Redis", zap.Error(err))
	}

	logger.Log.Info("Connected to Redis", zap.String("address", redisURL))
	return rdb
}

// runServer initializes and runs the HTTP server.
func runServer() {
	logger.Init(os.Getenv("ENV") == "production")

	app := fiber.New()
	app.Static("/docs/en", "./doc/en")
	app.Static("/docs/de", "./doc/de")

	app.Use(middleware.TraceID())
	app.Use(middleware.HTTPLogger(logger.Log))

	db := database.InitDB()
	migrations.RunMigrations()

	sender := email.NewMailer(logger.Log)

	authRepo := auth.NewSQLXRepository(db)
	authService := auth.NewService(authRepo, sender)

	adminRepo := admin.NewSQLXRepository(db)
	adminService := admin.NewService(adminRepo, logger.Log, sender)

	redisClient := initRedis()

	routes.RegisterAuthRoutes(app, authService, logger.Log, redisClient)
	routes.RegisterAdminRoutes(app, adminService, adminRepo, logger.Log)

	if err := app.Listen(":8080"); err != nil {
		logger.Log.Fatal("Failed to start server")
	}
}

func main() {
	initCLI()

	if len(os.Args) > 1 {
		if err := rootCmd.Execute(); err != nil {
			os.Exit(1)
		}
		return
	}

	runServer()
}
