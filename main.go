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
)

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

	// API routes - use the handler's RegisterRoutes method
	api := router.Group("/api/v1")
	orderHandler.RegisterRoutes(api)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}
