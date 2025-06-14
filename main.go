package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"online-order-management-system/internal/api/http/handler"
	"online-order-management-system/internal/infra/db"
	"online-order-management-system/internal/usecase/order"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func main() {
	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://user:password@localhost/orderdb?sslmode=disable"
	}

	database, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	// Test database connection
	if err := database.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	// Initialize repository
	orderRepo := db.NewPostgresOrderRepository(database)

	// Initialize use cases
	createOrderUC := order.NewCreateOrderUseCase(orderRepo)
	getOrderUC := order.NewGetOrderUseCase(orderRepo)
	listOrdersUC := order.NewListOrdersUseCase(orderRepo)
	updateOrderStatusUC := order.NewUpdateOrderStatusUseCase(orderRepo)
	bulkCreateOrdersUC := order.NewBulkCreateOrdersUseCase(orderRepo)

	// Initialize handler
	orderHandler := handler.NewOrderHandler(
		createOrderUC,
		getOrderUC,
		listOrdersUC,
		updateOrderStatusUC,
		bulkCreateOrdersUC,
	)

	// Initialize Gin router
	router := gin.Default()

	// Middleware
	router.Use(GinLoggingMiddleware())
	router.Use(CORSMiddleware())

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		orders := api.Group("/orders")
		{
			orders.POST("", orderHandler.CreateOrder)
			orders.POST("/bulk", orderHandler.BulkCreateOrders)
			orders.GET("", orderHandler.ListOrders)
			orders.GET("/:id", orderHandler.GetOrder)
			orders.PUT("/:id/status", orderHandler.UpdateOrderStatus)
		}
	}

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

// GinLoggingMiddleware provides request logging for Gin
func GinLoggingMiddleware() gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		return fmt.Sprintf("%s - [%s] \"%s %s %s %d %s \"%s\" %s\"\n",
			param.ClientIP,
			param.TimeStamp.Format(time.RFC1123),
			param.Method,
			param.Path,
			param.Request.Proto,
			param.StatusCode,
			param.Latency,
			param.Request.UserAgent(),
			param.ErrorMessage,
		)
	})
}

// CORSMiddleware handles Cross-Origin Resource Sharing
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
