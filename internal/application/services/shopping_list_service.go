// Package services contains the application layer business logic services.
package services

import (
	"context"

	"github.com/google/uuid"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
	"github.com/uriberma/go-shopping-list-api/internal/domain/repositories"
)

// ShoppingListService handles business logic for shopping lists
type ShoppingListService struct {
	shoppingListRepo repositories.ShoppingListRepository
	itemRepo         repositories.ItemRepository
}

// NewShoppingListService creates a new shopping list service
func NewShoppingListService(shoppingListRepo repositories.ShoppingListRepository, itemRepo repositories.ItemRepository) *ShoppingListService {
	return &ShoppingListService{
		shoppingListRepo: shoppingListRepo,
		itemRepo:         itemRepo,
	}
}

// CreateShoppingList creates a new shopping list
func (s *ShoppingListService) CreateShoppingList(ctx context.Context, name, description string) (*entities.ShoppingList, error) {
	if name == "" {
		return nil, entities.ErrInvalidInput
	}

	list := entities.NewShoppingList(name, description)
	if err := s.shoppingListRepo.Create(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

// GetShoppingList retrieves a shopping list by ID
func (s *ShoppingListService) GetShoppingList(ctx context.Context, id uuid.UUID) (*entities.ShoppingList, error) {
	list, err := s.shoppingListRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Load items for the shopping list
	items, err := s.itemRepo.GetByShoppingListID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Convert slice of pointers to slice of values
	list.Items = make([]entities.Item, len(items))
	for i, item := range items {
		list.Items[i] = *item
	}

	return list, nil
}

// GetAllShoppingLists retrieves all shopping lists
func (s *ShoppingListService) GetAllShoppingLists(ctx context.Context) ([]*entities.ShoppingList, error) {
	lists, err := s.shoppingListRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	// Load items for each shopping list
	for _, list := range lists {
		items, err := s.itemRepo.GetByShoppingListID(ctx, list.ID)
		if err != nil {
			return nil, err
		}

		list.Items = make([]entities.Item, len(items))
		for i, item := range items {
			list.Items[i] = *item
		}
	}

	return lists, nil
}

// UpdateShoppingList updates an existing shopping list
func (s *ShoppingListService) UpdateShoppingList(ctx context.Context, id uuid.UUID, name, description string) (*entities.ShoppingList, error) {
	if name == "" {
		return nil, entities.ErrInvalidInput
	}

	list, err := s.shoppingListRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	list.Name = name
	list.Description = description

	if err := s.shoppingListRepo.Update(ctx, list); err != nil {
		return nil, err
	}

	return list, nil
}

// DeleteShoppingList deletes a shopping list
func (s *ShoppingListService) DeleteShoppingList(ctx context.Context, id uuid.UUID) error {
	return s.shoppingListRepo.Delete(ctx, id)
}
