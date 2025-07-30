package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
)

// ShoppingListServiceInterface defines the interface for shopping list service
type ShoppingListServiceInterface interface {
	CreateShoppingList(ctx context.Context, name, description string) (*entities.ShoppingList, error)
	GetShoppingList(ctx context.Context, id uuid.UUID) (*entities.ShoppingList, error)
	GetAllShoppingLists(ctx context.Context) ([]*entities.ShoppingList, error)
	UpdateShoppingList(ctx context.Context, id uuid.UUID, name, description string) (*entities.ShoppingList, error)
	DeleteShoppingList(ctx context.Context, id uuid.UUID) error
}

// ItemServiceInterface defines the interface for item service
type ItemServiceInterface interface {
	CreateItem(ctx context.Context, shoppingListID uuid.UUID, name string, quantity int) (*entities.Item, error)
	GetItem(ctx context.Context, id uuid.UUID) (*entities.Item, error)
	GetItemsByShoppingListID(ctx context.Context, shoppingListID uuid.UUID) ([]*entities.Item, error)
	UpdateItem(ctx context.Context, id uuid.UUID, name string, quantity int, completed bool) (*entities.Item, error)
	DeleteItem(ctx context.Context, id uuid.UUID) error
	ToggleItemCompletion(ctx context.Context, id uuid.UUID) (*entities.Item, error)
}

// Ensure that the concrete services implement the interfaces
var _ ShoppingListServiceInterface = (*ShoppingListService)(nil)
var _ ItemServiceInterface = (*ItemService)(nil)
