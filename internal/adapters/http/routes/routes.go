package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/uriberma/go-shopping-list-api/internal/adapters/http/handlers"
)

// SetupRoutes configures all API routes with versioning
func SetupRoutes(
	router *gin.Engine,
	shoppingListHandler *handlers.ShoppingListHandler,
	itemHandler *handlers.ItemHandler,
) {
	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Shopping list routes
		v1.POST("/lists", shoppingListHandler.CreateShoppingList)
		v1.GET("/lists", shoppingListHandler.GetAllShoppingLists)
		v1.GET("/lists/:id", shoppingListHandler.GetShoppingList)
		v1.PUT("/lists/:id", shoppingListHandler.UpdateShoppingList)
		v1.DELETE("/lists/:id", shoppingListHandler.DeleteShoppingList)

		// Items within a specific shopping list (using different path to avoid conflicts)
		v1.POST("/shopping-lists/:listId/items", itemHandler.CreateItem)
		v1.GET("/shopping-lists/:listId/items", itemHandler.GetItemsByShoppingListID)

		// Item routes (for direct item operations)
		v1.GET("/items/:id", itemHandler.GetItem)
		v1.PUT("/items/:id", itemHandler.UpdateItem)
		v1.DELETE("/items/:id", itemHandler.DeleteItem)
		v1.PATCH("/items/:id/toggle", itemHandler.ToggleItemCompletion)
	}

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"service": "shopping-list-api",
			"version": "v1.0.0",
		})
	})
}
