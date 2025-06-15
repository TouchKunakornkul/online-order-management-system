package main

import (
	"log"
	"net/http"
	"online-order-management-system/internal/api/http/handler"
	"online-order-management-system/internal/infra/db"
	"online-order-management-system/internal/middleware"
	"online-order-management-system/internal/usecase/order"
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
	// Load .env file if it exists (ignore error if file doesn't exist)
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found or error loading .env file: %v", err)
	} else {
		log.Printf("✅ Loaded configuration from .env file")
	}

	// Database connection using environment-based configuration
	database, err := db.NewPostgresDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Initialize repository
	orderRepo := db.NewPostgresOrderRepository(database)

	// Initialize use cases
	createOrderUC := order.NewCreateOrderUseCase(orderRepo)
	getOrderUC := order.NewGetOrderUseCase(orderRepo)
	listOrdersUC := order.NewListOrdersUseCase(orderRepo)
	updateOrderStatusUC := order.NewUpdateOrderStatusUseCase(orderRepo)

	// Initialize handler
	orderHandler := handler.NewOrderHandler(
		createOrderUC,
		getOrderUC,
		listOrdersUC,
		updateOrderStatusUC,
	)

	// Initialize Gin router
	router := gin.Default()

	// Middleware
	router.Use(middleware.GinLoggingMiddleware())
	router.Use(middleware.CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Swagger documentation endpoint
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API routes - use the handler's RegisterRoutes method
	api := router.Group("/api/v1")
	orderHandler.RegisterRoutes(api)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	log.Printf("📚 Swagger documentation available at: http://localhost:%s/swagger/index.html", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
