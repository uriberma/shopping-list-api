// Package handlers contains HTTP request handlers for the shopping list API.
package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uriberma/go-shopping-list-api/internal/application/services"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
)

// ShoppingListHandler handles HTTP requests for shopping lists
type ShoppingListHandler struct {
	service services.ShoppingListServiceInterface
}

// NewShoppingListHandler creates a new shopping list handler
func NewShoppingListHandler(service services.ShoppingListServiceInterface) *ShoppingListHandler {
	return &ShoppingListHandler{service: service}
}

// CreateShoppingListRequest represents the request body for creating a shopping list
type CreateShoppingListRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// UpdateShoppingListRequest represents the request body for updating a shopping list
type UpdateShoppingListRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// CreateShoppingList creates a new shopping list
func (h *ShoppingListHandler) CreateShoppingList(c *gin.Context) {
	var req CreateShoppingListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list, err := h.service.CreateShoppingList(c.Request.Context(), req.Name, req.Description)
	if err != nil {
		if err == entities.ErrInvalidInput {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create shopping list"})
		return
	}

	c.JSON(http.StatusCreated, list)
}

// GetShoppingList retrieves a shopping list by ID
func (h *ShoppingListHandler) GetShoppingList(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	list, err := h.service.GetShoppingList(c.Request.Context(), id)
	if err != nil {
		if err == entities.ErrShoppingListNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Shopping list not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve shopping list"})
		return
	}

	c.JSON(http.StatusOK, list)
}

// GetAllShoppingLists retrieves all shopping lists
func (h *ShoppingListHandler) GetAllShoppingLists(c *gin.Context) {
	lists, err := h.service.GetAllShoppingLists(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve shopping lists"})
		return
	}

	c.JSON(http.StatusOK, lists)
}

// UpdateShoppingList updates an existing shopping list
func (h *ShoppingListHandler) UpdateShoppingList(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req UpdateShoppingListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	list, err := h.service.UpdateShoppingList(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		if err == entities.ErrShoppingListNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Shopping list not found"})
			return
		}
		if err == entities.ErrInvalidInput {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update shopping list"})
		return
	}

	c.JSON(http.StatusOK, list)
}

// DeleteShoppingList deletes a shopping list
func (h *ShoppingListHandler) DeleteShoppingList(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = h.service.DeleteShoppingList(c.Request.Context(), id)
	if err != nil {
		if err == entities.ErrShoppingListNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Shopping list not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete shopping list"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
