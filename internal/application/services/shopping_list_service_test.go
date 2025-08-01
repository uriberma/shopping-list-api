package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
)

func TestNewShoppingListService(t *testing.T) {
	itemRepo := &MockItemRepository{}
	shoppingListRepo := &MockShoppingListRepository{}

	service := NewShoppingListService(shoppingListRepo, itemRepo)

	assert.NotNil(t, service)
	assert.Equal(t, shoppingListRepo, service.shoppingListRepo)
	assert.Equal(t, itemRepo, service.itemRepo)
}

func TestShoppingListService_CreateShoppingList(t *testing.T) {
	tests := []struct {
		name          string
		listName      string
		description   string
		setupMocks    func(*MockShoppingListRepository)
		expectedError error
		expectedResult bool
	}{
		{
			name:        "successful creation",
			listName:    "Test List",
			description: "Test Description",
			setupMocks: func(listRepo *MockShoppingListRepository) {
				listRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError:  nil,
			expectedResult: true,
		},
		{
			name:        "empty name should fail",
			listName:    "",
			description: "Test Description",
			setupMocks:  func(listRepo *MockShoppingListRepository) {},
			expectedError:  entities.ErrInvalidInput,
			expectedResult: false,
		},
		{
			name:        "creation with empty description should succeed",
			listName:    "Test List",
			description: "",
			setupMocks: func(listRepo *MockShoppingListRepository) {
				listRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError:  nil,
			expectedResult: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemRepo := &MockItemRepository{}
			shoppingListRepo := &MockShoppingListRepository{}
			service := NewShoppingListService(shoppingListRepo, itemRepo)

			tt.setupMocks(shoppingListRepo)

			result, err := service.CreateShoppingList(context.Background(), tt.listName, tt.description)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectedResult {
					assert.NotNil(t, result)
					assert.Equal(t, tt.listName, result.Name)
					assert.Equal(t, tt.description, result.Description)
				}
			}

			shoppingListRepo.AssertExpectations(t)
		})
	}
}

func TestShoppingListService_GetShoppingList(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockShoppingListRepository, *MockItemRepository, uuid.UUID)
		hasItems   bool
	}{
		{
			name: "get shopping list with items",
			setupMocks: func(listRepo *MockShoppingListRepository, itemRepo *MockItemRepository, listID uuid.UUID) {
				expectedList := &entities.ShoppingList{
					ID:          listID,
					Name:        "Test List",
					Description: "Test Description",
				}
				expectedItems := []*entities.Item{
					{ID: uuid.New(), Name: "Item 1", ShoppingListID: listID},
					{ID: uuid.New(), Name: "Item 2", ShoppingListID: listID},
				}

				listRepo.On("GetByID", mock.Anything, listID).Return(expectedList, nil)
				itemRepo.On("GetByShoppingListID", mock.Anything, listID).Return(expectedItems, nil)
			},
			hasItems: true,
		},
		{
			name: "get shopping list without items",
			setupMocks: func(listRepo *MockShoppingListRepository, itemRepo *MockItemRepository, listID uuid.UUID) {
				expectedList := &entities.ShoppingList{
					ID:          listID,
					Name:        "Empty List",
					Description: "Empty Description",
				}
				expectedItems := []*entities.Item{}

				listRepo.On("GetByID", mock.Anything, listID).Return(expectedList, nil)
				itemRepo.On("GetByShoppingListID", mock.Anything, listID).Return(expectedItems, nil)
			},
			hasItems: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemRepo := &MockItemRepository{}
			shoppingListRepo := &MockShoppingListRepository{}
			service := NewShoppingListService(shoppingListRepo, itemRepo)

			listID := uuid.New()
			tt.setupMocks(shoppingListRepo, itemRepo, listID)

			result, err := service.GetShoppingList(context.Background(), listID)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, listID, result.ID)

			if tt.hasItems {
				assert.Len(t, result.Items, 2)
			} else {
				assert.Len(t, result.Items, 0)
			}

			shoppingListRepo.AssertExpectations(t)
			itemRepo.AssertExpectations(t)
		})
	}
}

func TestShoppingListService_GetShoppingList_NotFound(t *testing.T) {
	itemRepo := &MockItemRepository{}
	shoppingListRepo := &MockShoppingListRepository{}
	service := NewShoppingListService(shoppingListRepo, itemRepo)

	listID := uuid.New()
	shoppingListRepo.On("GetByID", mock.Anything, listID).Return((*entities.ShoppingList)(nil), entities.ErrShoppingListNotFound)

	result, err := service.GetShoppingList(context.Background(), listID)

	assert.Error(t, err)
	assert.Equal(t, entities.ErrShoppingListNotFound, err)
	assert.Nil(t, result)
	shoppingListRepo.AssertExpectations(t)
}

func TestShoppingListService_GetAllShoppingLists(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(*MockShoppingListRepository, *MockItemRepository)
		listCount  int
	}{
		{
			name: "get multiple lists with items",
			setupMocks: func(listRepo *MockShoppingListRepository, itemRepo *MockItemRepository) {
				list1ID := uuid.New()
				list2ID := uuid.New()
				
				expectedLists := []*entities.ShoppingList{
					{ID: list1ID, Name: "List 1"},
					{ID: list2ID, Name: "List 2"},
				}

				items1 := []*entities.Item{
					{ID: uuid.New(), Name: "Item 1", ShoppingListID: list1ID},
				}
				items2 := []*entities.Item{
					{ID: uuid.New(), Name: "Item 2", ShoppingListID: list2ID},
					{ID: uuid.New(), Name: "Item 3", ShoppingListID: list2ID},
				}

				listRepo.On("GetAll", mock.Anything).Return(expectedLists, nil)
				itemRepo.On("GetByShoppingListID", mock.Anything, list1ID).Return(items1, nil)
				itemRepo.On("GetByShoppingListID", mock.Anything, list2ID).Return(items2, nil)
			},
			listCount: 2,
		},
		{
			name: "get empty list",
			setupMocks: func(listRepo *MockShoppingListRepository, itemRepo *MockItemRepository) {
				expectedLists := []*entities.ShoppingList{}
				listRepo.On("GetAll", mock.Anything).Return(expectedLists, nil)
			},
			listCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemRepo := &MockItemRepository{}
			shoppingListRepo := &MockShoppingListRepository{}
			service := NewShoppingListService(shoppingListRepo, itemRepo)

			tt.setupMocks(shoppingListRepo, itemRepo)

			result, err := service.GetAllShoppingLists(context.Background())

			assert.NoError(t, err)
			assert.Len(t, result, tt.listCount)

			shoppingListRepo.AssertExpectations(t)
			itemRepo.AssertExpectations(t)
		})
	}
}

func TestShoppingListService_UpdateShoppingList(t *testing.T) {
	tests := []struct {
		name          string
		listName      string
		description   string
		setupMocks    func(*MockShoppingListRepository, uuid.UUID)
		expectedError error
	}{
		{
			name:        "successful update",
			listName:    "Updated List",
			description: "Updated Description",
			setupMocks: func(listRepo *MockShoppingListRepository, listID uuid.UUID) {
				existingList := &entities.ShoppingList{
					ID:          listID,
					Name:        "Old List",
					Description: "Old Description",
				}
				listRepo.On("GetByID", mock.Anything, listID).Return(existingList, nil)
				listRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "empty name should fail",
			listName:    "",
			description: "Updated Description",
			setupMocks: func(listRepo *MockShoppingListRepository, listID uuid.UUID) {
				// No mocks needed as validation happens before repository calls
			},
			expectedError: entities.ErrInvalidInput,
		},
		{
			name:        "list not found",
			listName:    "Updated List",
			description: "Updated Description",
			setupMocks: func(listRepo *MockShoppingListRepository, listID uuid.UUID) {
				listRepo.On("GetByID", mock.Anything, listID).Return((*entities.ShoppingList)(nil), entities.ErrShoppingListNotFound)
			},
			expectedError: entities.ErrShoppingListNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemRepo := &MockItemRepository{}
			shoppingListRepo := &MockShoppingListRepository{}
			service := NewShoppingListService(shoppingListRepo, itemRepo)

			listID := uuid.New()
			tt.setupMocks(shoppingListRepo, listID)

			result, err := service.UpdateShoppingList(context.Background(), listID, tt.listName, tt.description)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.listName, result.Name)
				assert.Equal(t, tt.description, result.Description)
			}

			shoppingListRepo.AssertExpectations(t)
		})
	}
}

func TestShoppingListService_DeleteShoppingList(t *testing.T) {
	itemRepo := &MockItemRepository{}
	shoppingListRepo := &MockShoppingListRepository{}
	service := NewShoppingListService(shoppingListRepo, itemRepo)

	listID := uuid.New()
	shoppingListRepo.On("Delete", mock.Anything, listID).Return(nil)

	err := service.DeleteShoppingList(context.Background(), listID)

	assert.NoError(t, err)
	shoppingListRepo.AssertExpectations(t)
}

func TestShoppingListService_DeleteShoppingList_NotFound(t *testing.T) {
	itemRepo := &MockItemRepository{}
	shoppingListRepo := &MockShoppingListRepository{}
	service := NewShoppingListService(shoppingListRepo, itemRepo)

	listID := uuid.New()
	shoppingListRepo.On("Delete", mock.Anything, listID).Return(entities.ErrShoppingListNotFound)

	err := service.DeleteShoppingList(context.Background(), listID)

	assert.Error(t, err)
	assert.Equal(t, entities.ErrShoppingListNotFound, err)
	shoppingListRepo.AssertExpectations(t)
}
