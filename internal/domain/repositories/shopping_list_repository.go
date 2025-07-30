package repositories

import (
	"context"

	"github.com/google/uuid"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
)

// ShoppingListRepository defines the contract for shopping list persistence
type ShoppingListRepository interface {
	Create(ctx context.Context, list *entities.ShoppingList) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.ShoppingList, error)
	GetAll(ctx context.Context) ([]*entities.ShoppingList, error)
	Update(ctx context.Context, list *entities.ShoppingList) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ItemRepository defines the contract for item persistence
type ItemRepository interface {
	Create(ctx context.Context, item *entities.Item) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Item, error)
	GetByShoppingListID(ctx context.Context, shoppingListID uuid.UUID) ([]*entities.Item, error)
	Update(ctx context.Context, item *entities.Item) error
	Delete(ctx context.Context, id uuid.UUID) error
}
