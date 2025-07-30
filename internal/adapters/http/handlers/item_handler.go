package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/uriberma/go-shopping-list-api/internal/application/services"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
)

// ItemHandler handles HTTP requests for items
type ItemHandler struct {
	service services.ItemServiceInterface
}

// NewItemHandler creates a new item handler
func NewItemHandler(service services.ItemServiceInterface) *ItemHandler {
	return &ItemHandler{service: service}
}

// CreateItemRequest represents the request body for creating an item
type CreateItemRequest struct {
	Name     string `json:"name" binding:"required"`
	Quantity int    `json:"quantity"`
}

// UpdateItemRequest represents the request body for updating an item
type UpdateItemRequest struct {
	Name      string `json:"name" binding:"required"`
	Quantity  int    `json:"quantity"`
	Completed bool   `json:"completed"`
}

// CreateItem creates a new item in a shopping list
func (h *ItemHandler) CreateItem(c *gin.Context) {
	listIDParam := c.Param("listId")
	listID, err := uuid.Parse(listIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID format"})
		return
	}

	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	item, err := h.service.CreateItem(c.Request.Context(), listID, req.Name, req.Quantity)
	if err != nil {
		if err == entities.ErrInvalidInput {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err == entities.ErrShoppingListNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Shopping list not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create item"})
		return
	}

	c.JSON(http.StatusCreated, item)
}

// GetItem retrieves an item by ID
func (h *ItemHandler) GetItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	item, err := h.service.GetItem(c.Request.Context(), id)
	if err != nil {
		if err == entities.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// GetItemsByShoppingListID retrieves all items for a shopping list
func (h *ItemHandler) GetItemsByShoppingListID(c *gin.Context) {
	listIDParam := c.Param("listId")
	listID, err := uuid.Parse(listIDParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid list ID format"})
		return
	}

	items, err := h.service.GetItemsByShoppingListID(c.Request.Context(), listID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve items"})
		return
	}

	c.JSON(http.StatusOK, items)
}

// UpdateItem updates an existing item
func (h *ItemHandler) UpdateItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	item, err := h.service.UpdateItem(c.Request.Context(), id, req.Name, req.Quantity, req.Completed)
	if err != nil {
		if err == entities.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		if err == entities.ErrInvalidInput {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update item"})
		return
	}

	c.JSON(http.StatusOK, item)
}

// DeleteItem deletes an item
func (h *ItemHandler) DeleteItem(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	err = h.service.DeleteItem(c.Request.Context(), id)
	if err != nil {
		if err == entities.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete item"})
		return
	}

	c.JSON(http.StatusNoContent, nil)
}

// ToggleItemCompletion toggles the completion status of an item
func (h *ItemHandler) ToggleItemCompletion(c *gin.Context) {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	item, err := h.service.ToggleItemCompletion(c.Request.Context(), id)
	if err != nil {
		if err == entities.ErrItemNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Item not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to toggle item completion"})
		return
	}

	c.JSON(http.StatusOK, item)
}
