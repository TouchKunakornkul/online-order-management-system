package main

import (
	"net/http"
	"online-order-management-system/internal/api/http/handler"
	"online-order-management-system/internal/infra/db"
	"online-order-management-system/internal/middleware"
	"online-order-management-system/internal/usecase/order"
	"online-order-management-system/pkg/logger"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	// Swagger imports
	_ "online-order-management-system/docs" // This will be generated

	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// @title           Online Order Management System API
// @version         1.0
// @description     A high-performance order management system built with Go, featuring concurrent order processing and Clean Architecture.
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.swagger.io/support
// @contact.email  support@swagger.io

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.basic  BasicAuth

// @externalDocs.description  OpenAPI
// @externalDocs.url          https://swagger.io/resources/open-api/
func main() {
	// Initialize structured logger
	appLogger := logger.New("order-management-system", "1.0.0")

	// Load .env file if it exists (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		appLogger.WithError(err).Warn("No .env file found or error loading .env file")
	} else {
		appLogger.Info("Loaded configuration from .env file")
	}

	// Database connection using environment-based configuration
	database, err := db.NewPostgresDB()
	if err != nil {
		appLogger.WithError(err).Fatal("Failed to connect to database")
	}
	defer func() {
		if err := database.Close(); err != nil {
			appLogger.WithError(err).Error("Failed to close database connection")
		}
	}()

	appLogger.Info("Successfully connected to database")

	// Initialize repository
	orderRepo := db.NewPostgresOrderRepository(database)

	// Initialize use cases
	createOrderUC := order.NewCreateOrderUseCase(orderRepo)
	getOrderUC := order.NewGetOrderUseCase(orderRepo)
	listOrdersUC := order.NewListOrdersUseCase(orderRepo)
	updateOrderStatusUC := order.NewUpdateOrderStatusUseCase(orderRepo)

	appLogger.Info("Initialized all use cases")

	// Initialize handler
	orderHandler := handler.NewOrderHandler(
		createOrderUC,
		getOrderUC,
		listOrdersUC,
		updateOrderStatusUC,
	)

	appLogger.Info("Initialized handlers")

	// Initialize Gin router
	router := gin.Default()

	// Middleware
	router.Use(middleware.GinLoggingMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "order-management-system",
			"version": "1.0.0",
		})
	})

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes - use the handler's RegisterRoutes method
	api := router.Group("/api/v1")
	orderHandler.RegisterRoutes(api)

	appLogger.Info("Registered all routes and middleware")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	appLogger.WithFields(map[string]interface{}{
		"port":        port,
		"swagger_url": "http://localhost:" + port + "/swagger/index.html",
	}).Info("Starting server")

	if err := router.Run(":" + port); err != nil {
		appLogger.WithError(err).WithField("port", port).Fatal("Failed to start server")
	}
}
