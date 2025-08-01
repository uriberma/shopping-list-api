package services

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
)

// MockItemRepository is a mock implementation of ItemRepository
type MockItemRepository struct {
	mock.Mock
}

func (m *MockItemRepository) Create(ctx context.Context, item *entities.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Item, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.Item), args.Error(1)
}

func (m *MockItemRepository) GetByShoppingListID(ctx context.Context, shoppingListID uuid.UUID) ([]*entities.Item, error) {
	args := m.Called(ctx, shoppingListID)
	return args.Get(0).([]*entities.Item), args.Error(1)
}

func (m *MockItemRepository) Update(ctx context.Context, item *entities.Item) error {
	args := m.Called(ctx, item)
	return args.Error(0)
}

func (m *MockItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockShoppingListRepository is a mock implementation of ShoppingListRepository
type MockShoppingListRepository struct {
	mock.Mock
}

func (m *MockShoppingListRepository) Create(ctx context.Context, list *entities.ShoppingList) error {
	args := m.Called(ctx, list)
	return args.Error(0)
}

func (m *MockShoppingListRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.ShoppingList, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*entities.ShoppingList), args.Error(1)
}

func (m *MockShoppingListRepository) GetAll(ctx context.Context) ([]*entities.ShoppingList, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*entities.ShoppingList), args.Error(1)
}

func (m *MockShoppingListRepository) Update(ctx context.Context, list *entities.ShoppingList) error {
	args := m.Called(ctx, list)
	return args.Error(0)
}

func (m *MockShoppingListRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func TestNewItemService(t *testing.T) {
	itemRepo := &MockItemRepository{}
	shoppingListRepo := &MockShoppingListRepository{}

	service := NewItemService(itemRepo, shoppingListRepo)

	assert.NotNil(t, service)
	assert.Equal(t, itemRepo, service.itemRepo)
	assert.Equal(t, shoppingListRepo, service.shoppingListRepo)
}

func TestItemService_CreateItem(t *testing.T) {
	tests := []struct {
		name           string
		itemName       string
		quantity       int
		shoppingListID uuid.UUID
		setupMocks     func(*MockItemRepository, *MockShoppingListRepository)
		expectedError  error
		expectedResult bool
	}{
		{
			name:           "successful creation",
			itemName:       "Test Item",
			quantity:       2,
			shoppingListID: uuid.New(),
			setupMocks: func(itemRepo *MockItemRepository, listRepo *MockShoppingListRepository) {
				listRepo.On("GetByID", mock.Anything, mock.Anything).Return(&entities.ShoppingList{}, nil)
				itemRepo.On("Create", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError:  nil,
			expectedResult: true,
		},
		{
			name:           "empty name should fail",
			itemName:       "",
			quantity:       2,
			shoppingListID: uuid.New(),
			setupMocks:     func(itemRepo *MockItemRepository, listRepo *MockShoppingListRepository) {},
			expectedError:  entities.ErrInvalidInput,
			expectedResult: false,
		},
		{
			name:           "shopping list not found",
			itemName:       "Test Item",
			quantity:       2,
			shoppingListID: uuid.New(),
			setupMocks: func(itemRepo *MockItemRepository, listRepo *MockShoppingListRepository) {
				listRepo.On("GetByID", mock.Anything, mock.Anything).Return((*entities.ShoppingList)(nil), entities.ErrShoppingListNotFound)
			},
			expectedError:  entities.ErrShoppingListNotFound,
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemRepo := &MockItemRepository{}
			shoppingListRepo := &MockShoppingListRepository{}
			service := NewItemService(itemRepo, shoppingListRepo)

			tt.setupMocks(itemRepo, shoppingListRepo)

			result, err := service.CreateItem(context.Background(), tt.shoppingListID, tt.itemName, tt.quantity)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				if tt.expectedResult {
					assert.NotNil(t, result)
					assert.Equal(t, tt.itemName, result.Name)
					assert.Equal(t, tt.quantity, result.Quantity)
					assert.Equal(t, tt.shoppingListID, result.ShoppingListID)
				}
			}

			itemRepo.AssertExpectations(t)
			shoppingListRepo.AssertExpectations(t)
		})
	}
}

func TestItemService_GetItem(t *testing.T) {
	itemRepo := &MockItemRepository{}
	shoppingListRepo := &MockShoppingListRepository{}
	service := NewItemService(itemRepo, shoppingListRepo)

	itemID := uuid.New()
	expectedItem := &entities.Item{ID: itemID, Name: "Test Item"}

	itemRepo.On("GetByID", mock.Anything, itemID).Return(expectedItem, nil)

	result, err := service.GetItem(context.Background(), itemID)

	assert.NoError(t, err)
	assert.Equal(t, expectedItem, result)
	itemRepo.AssertExpectations(t)
}

func TestItemService_GetItemsByShoppingListID(t *testing.T) {
	itemRepo := &MockItemRepository{}
	shoppingListRepo := &MockShoppingListRepository{}
	service := NewItemService(itemRepo, shoppingListRepo)

	shoppingListID := uuid.New()
	expectedItems := []*entities.Item{
		{ID: uuid.New(), Name: "Item 1"},
		{ID: uuid.New(), Name: "Item 2"},
	}

	itemRepo.On("GetByShoppingListID", mock.Anything, shoppingListID).Return(expectedItems, nil)

	result, err := service.GetItemsByShoppingListID(context.Background(), shoppingListID)

	assert.NoError(t, err)
	assert.Equal(t, expectedItems, result)
	itemRepo.AssertExpectations(t)
}

func TestItemService_UpdateItem(t *testing.T) {
	tests := []struct {
		name          string
		itemName      string
		quantity      int
		completed     bool
		setupMocks    func(*MockItemRepository, uuid.UUID)
		expectedError error
	}{
		{
			name:      "successful update",
			itemName:  "Updated Item",
			quantity:  5,
			completed: true,
			setupMocks: func(itemRepo *MockItemRepository, itemID uuid.UUID) {
				existingItem := &entities.Item{ID: itemID, Name: "Old Item", Quantity: 1, Completed: false}
				itemRepo.On("GetByID", mock.Anything, itemID).Return(existingItem, nil)
				itemRepo.On("Update", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "empty name should fail",
			itemName:  "",
			quantity:  5,
			completed: true,
			setupMocks: func(itemRepo *MockItemRepository, itemID uuid.UUID) {
				// No mocks needed as validation happens before repository calls
			},
			expectedError: entities.ErrInvalidInput,
		},
		{
			name:      "item not found",
			itemName:  "Updated Item",
			quantity:  5,
			completed: true,
			setupMocks: func(itemRepo *MockItemRepository, itemID uuid.UUID) {
				itemRepo.On("GetByID", mock.Anything, itemID).Return((*entities.Item)(nil), entities.ErrItemNotFound)
			},
			expectedError: entities.ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemRepo := &MockItemRepository{}
			shoppingListRepo := &MockShoppingListRepository{}
			service := NewItemService(itemRepo, shoppingListRepo)

			itemID := uuid.New()
			tt.setupMocks(itemRepo, itemID)

			result, err := service.UpdateItem(context.Background(), itemID, tt.itemName, tt.quantity, tt.completed)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.itemName, result.Name)
				assert.Equal(t, tt.quantity, result.Quantity)
				assert.Equal(t, tt.completed, result.Completed)
			}

			itemRepo.AssertExpectations(t)
		})
	}
}

func TestItemService_DeleteItem(t *testing.T) {
	itemRepo := &MockItemRepository{}
	shoppingListRepo := &MockShoppingListRepository{}
	service := NewItemService(itemRepo, shoppingListRepo)

	itemID := uuid.New()
	itemRepo.On("Delete", mock.Anything, itemID).Return(nil)

	err := service.DeleteItem(context.Background(), itemID)

	assert.NoError(t, err)
	itemRepo.AssertExpectations(t)
}

func TestItemService_ToggleItemCompletion(t *testing.T) {
	tests := []struct {
		name              string
		initialCompleted  bool
		expectedCompleted bool
	}{
		{
			name:              "toggle from incomplete to complete",
			initialCompleted:  false,
			expectedCompleted: true,
		},
		{
			name:              "toggle from complete to incomplete",
			initialCompleted:  true,
			expectedCompleted: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			itemRepo := &MockItemRepository{}
			shoppingListRepo := &MockShoppingListRepository{}
			service := NewItemService(itemRepo, shoppingListRepo)

			itemID := uuid.New()
			existingItem := &entities.Item{
				ID:        itemID,
				Name:      "Test Item",
				Completed: tt.initialCompleted,
			}

			itemRepo.On("GetByID", mock.Anything, itemID).Return(existingItem, nil)
			itemRepo.On("Update", mock.Anything, mock.Anything).Return(nil)

			result, err := service.ToggleItemCompletion(context.Background(), itemID)

			assert.NoError(t, err)
			assert.NotNil(t, result)
			assert.Equal(t, tt.expectedCompleted, result.Completed)
			itemRepo.AssertExpectations(t)
		})
	}
}
