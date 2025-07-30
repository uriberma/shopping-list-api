package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/uriberma/go-shopping-list-api/internal/adapters/http/handlers"
	"github.com/uriberma/go-shopping-list-api/internal/adapters/http/routes"
	"github.com/uriberma/go-shopping-list-api/internal/application/services"
	"github.com/uriberma/go-shopping-list-api/internal/infrastructure/database"
	"github.com/uriberma/go-shopping-list-api/internal/infrastructure/persistence"
)

func main() {
	// Database configuration
	dbConfig := database.Config{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "postgres"),
		Password: getEnv("DB_PASSWORD", "password"),
		DBName:   getEnv("DB_NAME", "shopping_list_db"),
		SSLMode:  getEnv("DB_SSLMODE", "disable"),
	}

	// Connect to database
	db, err := database.NewPostgresConnection(dbConfig)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Note: Database migrations are now handled by the separate migrator tool
	// Run: go run ./cmd/migrator/main.go -action=up

	// Initialize repositories
	shoppingListRepo := persistence.NewPostgresShoppingListRepository(db)
	itemRepo := persistence.NewPostgresItemRepository(db)

	// Initialize services
	shoppingListService := services.NewShoppingListService(shoppingListRepo, itemRepo)
	itemService := services.NewItemService(itemRepo, shoppingListRepo)

	// Initialize handlers
	shoppingListHandler := handlers.NewShoppingListHandler(shoppingListService)
	itemHandler := handlers.NewItemHandler(itemService)

	// Setup Gin router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Setup routes
	routes.SetupRoutes(router, shoppingListHandler, itemHandler)

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Starting server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
