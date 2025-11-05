package main

import (
	"log"
	"os"
	"time"

	"github.com/code-zt/vidnotes/config"
	"github.com/code-zt/vidnotes/internal/handlers"
	"github.com/code-zt/vidnotes/internal/repository"
	"github.com/code-zt/vidnotes/internal/routes"
	"github.com/code-zt/vidnotes/internal/services"
	"github.com/code-zt/vidnotes/pkg/auth"
	"github.com/code-zt/vidnotes/pkg/database"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// Загрузка .env файла
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Инициализация MongoDB
	mongoURI := os.Getenv("MONGODB_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "vidnotes"
	}

	mongoClient, err := database.NewMongoClient(mongoURI, dbName)
	if err != nil {
		log.Fatal("Failed to connect to MongoDB:", err)
	}
	defer mongoClient.Close()

	log.Println("Successfully connected to MongoDB")

	// Инициализация gRPC клиента
	grpcAddr := os.Getenv("GRPC_SERVER_ADDR")
	if grpcAddr == "" {
		// В Docker используем имя сервиса из docker-compose
		if os.Getenv("DOCKER_ENV") == "true" {
			grpcAddr = "processor:50051" // Имя сервиса в docker-compose
		} else {
			grpcAddr = "localhost:50051"
		}
	}

	log.Printf("Connecting to gRPC server at: %s", grpcAddr)

	grpcConn, err := grpc.Dial(grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(500*1024*1024),
			grpc.MaxCallSendMsgSize(500*1024*1024),
		),
	)
	if err != nil {
		log.Fatal("Failed to connect to gRPC server:", err)
	}
	defer grpcConn.Close()

	log.Println("Successfully connected to gRPC server")

	// Инициализация репозиториев
	userRepo := repository.NewUserRepository(mongoClient.DB)
	videoRepo := repository.NewVideoRepository(mongoClient.DB)
	sessionRepo := repository.NewAISessionRepository(mongoClient.DB)

	// Инициализация сервисов
	userService := services.NewUserService(userRepo)
	videoService := services.NewVideoService(videoRepo, userService, grpcConn)

	// Инициализация AI сервиса
	openRouterConfig := config.GetOpenRouterConfig()
	aiService := services.NewOpenRouterService(openRouterConfig)

	// Инициализация JWT
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "your-default-jwt-secret-change-in-production"
	}

	jwtRefreshSecret := os.Getenv("JWT_REFRESH_SECRET")
	if jwtRefreshSecret == "" {
		jwtRefreshSecret = "your-default-refresh-secret-change-in-production"
	}

	jwtManager := auth.NewJWTManager(
		jwtSecret,
		jwtRefreshSecret,
		24*time.Hour,   // Access token duration
		7*24*time.Hour, // Refresh token duration
	)

	// Инициализация handlers
	userHandlers := handlers.NewUserHandlers(userService, jwtManager)
	videoHandlers := handlers.NewVideoHandlers(videoService)
	aiHandlers := handlers.NewAIHandlers(aiService, sessionRepo, videoRepo)

	// Создание Fiber приложения
	app := fiber.New(fiber.Config{
		BodyLimit: 500 * 1024 * 1024, // 500MB для загрузки видео
		AppName:   "VidNotes API",
	})

	// CORS только для разработки
	if os.Getenv("DOCKER_ENV") != "true" {
		app.Use(cors.New(cors.Config{
			AllowOrigins:     "http://localhost:3000,http://127.0.0.1:3000,http://localhost:80",
			AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Content-Length",
			AllowMethods:     "GET, POST, PUT, DELETE, OPTIONS",
			AllowCredentials: true,
		}))
	}

	// Middleware логгера
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${method} ${path}\n",
	}))

	// Health check
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":    "ok",
			"service":   "VidNotes API",
			"timestamp": time.Now().UTC(),
		})
	})

	// Docs (OpenAPI + Redoc)
	routes.SetupDocs(app)

	// Настройка маршрутов
	routes.SetupRoutes(app, jwtManager, userHandlers, videoHandlers, aiHandlers)

	// Запуск сервера
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// В Docker всегда слушаем 0.0.0.0
	listenAddr := ":" + port
	if os.Getenv("DOCKER_ENV") == "true" {
		listenAddr = "0.0.0.0:" + port
	}

	log.Printf("Server starting on %s", listenAddr)
	if err := app.Listen(listenAddr); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
