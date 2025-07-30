package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
	"github.com/uriberma/go-shopping-list-api/internal/domain/repositories"
)

// ItemService handles business logic for items
type ItemService struct {
	itemRepo         repositories.ItemRepository
	shoppingListRepo repositories.ShoppingListRepository
}

// NewItemService creates a new item service
func NewItemService(itemRepo repositories.ItemRepository, shoppingListRepo repositories.ShoppingListRepository) *ItemService {
	return &ItemService{
		itemRepo:         itemRepo,
		shoppingListRepo: shoppingListRepo,
	}
}

// CreateItem creates a new item in a shopping list
func (s *ItemService) CreateItem(ctx context.Context, shoppingListID uuid.UUID, name string, quantity int) (*entities.Item, error) {
	if name == "" {
		return nil, entities.ErrInvalidInput
	}

	// Verify shopping list exists
	_, err := s.shoppingListRepo.GetByID(ctx, shoppingListID)
	if err != nil {
		return nil, entities.ErrShoppingListNotFound
	}

	item := entities.NewItem(name, quantity)
	item.ShoppingListID = shoppingListID

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// GetItem retrieves an item by ID
func (s *ItemService) GetItem(ctx context.Context, id uuid.UUID) (*entities.Item, error) {
	return s.itemRepo.GetByID(ctx, id)
}

// GetItemsByShoppingListID retrieves all items for a shopping list
func (s *ItemService) GetItemsByShoppingListID(ctx context.Context, shoppingListID uuid.UUID) ([]*entities.Item, error) {
	return s.itemRepo.GetByShoppingListID(ctx, shoppingListID)
}

// UpdateItem updates an existing item
func (s *ItemService) UpdateItem(ctx context.Context, id uuid.UUID, name string, quantity int, completed bool) (*entities.Item, error) {
	if name == "" {
		return nil, entities.ErrInvalidInput
	}

	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	item.Name = name
	item.Quantity = quantity
	item.Completed = completed

	if err := s.itemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}

// DeleteItem deletes an item
func (s *ItemService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	return s.itemRepo.Delete(ctx, id)
}

// ToggleItemCompletion toggles the completion status of an item
func (s *ItemService) ToggleItemCompletion(ctx context.Context, id uuid.UUID) (*entities.Item, error) {
	item, err := s.itemRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if item.Completed {
		item.MarkIncomplete()
	} else {
		item.MarkCompleted()
	}

	if err := s.itemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	return item, nil
}
