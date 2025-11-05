package routes

import (
	"github.com/code-zt/vidnotes/internal/handlers"
	"github.com/code-zt/vidnotes/pkg/auth"
	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	jwtManager *auth.JWTManager,
	userHandlers *handlers.UserHandlers,
	videoHandlers *handlers.VideoHandlers,
	aiHandlers *handlers.AIHandlers,
) {
	api := app.Group("/api/v1")

	// Public routes - Authentication
	authGroup := api.Group("/auth")
	{
		authGroup.Post("/register", userHandlers.Register)
		authGroup.Post("/login", userHandlers.Login)
		authGroup.Post("/refresh", userHandlers.RefreshToken)
	}

	// Protected routes (require JWT)
	protected := api.Group("", jwtManager.Middleware())
	{
		// User routes
		userGroup := protected.Group("/user")
		{
			userGroup.Get("/profile", userHandlers.GetProfile)
			userGroup.Put("/profile", userHandlers.UpdateProfile)
			userGroup.Post("/change-password", userHandlers.ChangePassword)
			userGroup.Get("/analytics", userHandlers.GetAnalytics)
		}

		// Video routes
		videosGroup := protected.Group("/videos")
		{
			videosGroup.Post("/upload", videoHandlers.UploadVideo)
			videosGroup.Get("/", videoHandlers.GetUserVideos)
			videosGroup.Get("/:id", videoHandlers.GetVideoStatus)
			videosGroup.Get("/:id/result", videoHandlers.GetVideoResult)
			videosGroup.Delete("/:id", videoHandlers.DeleteVideo)
		}

		// AI routes
		aiGroup := protected.Group("/ai")
		{
			// Основные операции с сессиями
			aiGroup.Post("/sessions", aiHandlers.CreateSession)
			aiGroup.Get("/sessions", aiHandlers.GetUserSessions)
			aiGroup.Get("/sessions/:id", aiHandlers.GetSession)
			aiGroup.Post("/sessions/:id/message", aiHandlers.SendMessage)
			aiGroup.Delete("/sessions/:id", aiHandlers.DeleteSession)

		}
	}

	// 404 Handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   "Endpoint not found",
			"message": "The requested endpoint does not exist",
		})
	})
}
