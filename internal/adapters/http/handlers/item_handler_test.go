package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uriberma/go-shopping-list-api/internal/application/services"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
)

// MockItemService is a mock implementation of the item service interface
type MockItemService struct {
	mock.Mock
}

// Ensure MockItemService implements the interface
var _ services.ItemServiceInterface = (*MockItemService)(nil)

func (m *MockItemService) CreateItem(ctx context.Context, shoppingListID uuid.UUID, name string, quantity int) (*entities.Item, error) {
	args := m.Called(ctx, shoppingListID, name, quantity)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Item), args.Error(1)
}

func (m *MockItemService) GetItem(ctx context.Context, id uuid.UUID) (*entities.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Item), args.Error(1)
}

func (m *MockItemService) GetItemsByShoppingListID(ctx context.Context, shoppingListID uuid.UUID) ([]*entities.Item, error) {
	args := m.Called(ctx, shoppingListID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.Item), args.Error(1)
}

func (m *MockItemService) UpdateItem(ctx context.Context, id uuid.UUID, name string, quantity int, completed bool) (*entities.Item, error) {
	args := m.Called(ctx, id, name, quantity, completed)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Item), args.Error(1)
}

func (m *MockItemService) DeleteItem(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockItemService) ToggleItemCompletion(ctx context.Context, id uuid.UUID) (*entities.Item, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.Item), args.Error(1)
}

func TestNewItemHandler(t *testing.T) {
	mockService := &MockItemService{}
	handler := NewItemHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
}

func TestItemHandler_CreateItem(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		requestBody    interface{}
		mockSetup      func(*MockItemService)
		expectedStatus int
		expectedBody   func(*testing.T, map[string]interface{})
	}{
		{
			name:   "successfully creates item",
			listID: uuid.New().String(),
			requestBody: CreateItemRequest{
				Name:     "Milk",
				Quantity: 2,
			},
			mockSetup: func(m *MockItemService) {
				expectedItem := &entities.Item{
					ID:       uuid.New(),
					Name:     "Milk",
					Quantity: 2,
				}
				m.On("CreateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Milk", 2).Return(expectedItem, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Milk", body["name"])
				assert.Equal(t, float64(2), body["quantity"])
			},
		},
		{
			name:   "creates item with default quantity when zero",
			listID: uuid.New().String(),
			requestBody: CreateItemRequest{
				Name:     "Bread",
				Quantity: 0,
			},
			mockSetup: func(m *MockItemService) {
				expectedItem := &entities.Item{
					ID:       uuid.New(),
					Name:     "Bread",
					Quantity: 1,
				}
				m.On("CreateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Bread", 1).Return(expectedItem, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Bread", body["name"])
				assert.Equal(t, float64(1), body["quantity"])
			},
		},
		{
			name:   "creates item with default quantity when negative",
			listID: uuid.New().String(),
			requestBody: CreateItemRequest{
				Name:     "Eggs",
				Quantity: -5,
			},
			mockSetup: func(m *MockItemService) {
				expectedItem := &entities.Item{
					ID:       uuid.New(),
					Name:     "Eggs",
					Quantity: 1,
				}
				m.On("CreateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Eggs", 1).Return(expectedItem, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Eggs", body["name"])
				assert.Equal(t, float64(1), body["quantity"])
			},
		},
		{
			name:           "fails with invalid list ID",
			listID:         "invalid-uuid",
			requestBody:    CreateItemRequest{Name: "Test", Quantity: 1},
			mockSetup:      func(m *MockItemService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Invalid list ID format", body["error"])
			},
		},
		{
			name:           "fails with missing name",
			listID:         uuid.New().String(),
			requestBody:    map[string]interface{}{"quantity": 1},
			mockSetup:      func(m *MockItemService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "required")
			},
		},
		{
			name:   "fails with invalid input error from service",
			listID: uuid.New().String(),
			requestBody: CreateItemRequest{
				Name:     "ValidName",
				Quantity: 1,
			},
			mockSetup: func(m *MockItemService) {
				m.On("CreateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "ValidName", 1).Return(nil, entities.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, entities.ErrInvalidInput.Error(), body["error"])
			},
		},
		{
			name:   "fails with shopping list not found",
			listID: uuid.New().String(),
			requestBody: CreateItemRequest{
				Name:     "Test Item",
				Quantity: 1,
			},
			mockSetup: func(m *MockItemService) {
				m.On("CreateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Test Item", 1).Return(nil, entities.ErrShoppingListNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Shopping list not found", body["error"])
			},
		},
		{
			name:   "fails with internal server error",
			listID: uuid.New().String(),
			requestBody: CreateItemRequest{
				Name:     "Test Item",
				Quantity: 1,
			},
			mockSetup: func(m *MockItemService) {
				m.On("CreateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Test Item", 1).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Failed to create item", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.mockSetup(mockService)

			handler := NewItemHandler(mockService)
			router := setupTestRouter()
			router.POST("/shopping-lists/:listId/items", handler.CreateItem)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/shopping-lists/"+tt.listID+"/items", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.expectedBody(t, responseBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestItemHandler_GetItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		mockSetup      func(*MockItemService)
		expectedStatus int
		expectedBody   func(*testing.T, map[string]interface{})
	}{
		{
			name:   "successfully gets item",
			itemID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				expectedItem := &entities.Item{
					ID:       uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Name:     "Test Item",
					Quantity: 2,
					Completed: false,
				}
				m.On("GetItem", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(expectedItem, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Test Item", body["name"])
				assert.Equal(t, float64(2), body["quantity"])
				assert.Equal(t, false, body["completed"])
			},
		},
		{
			name:           "fails with invalid UUID",
			itemID:         "invalid-uuid",
			mockSetup:      func(m *MockItemService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Invalid ID format", body["error"])
			},
		},
		{
			name:   "fails with not found error",
			itemID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				m.On("GetItem", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, entities.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Item not found", body["error"])
			},
		},
		{
			name:   "fails with internal server error",
			itemID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				m.On("GetItem", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Failed to retrieve item", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.mockSetup(mockService)

			handler := NewItemHandler(mockService)
			router := setupTestRouter()
			router.GET("/items/:id", handler.GetItem)

			req := httptest.NewRequest(http.MethodGet, "/items/"+tt.itemID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.expectedBody(t, responseBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestItemHandler_GetItemsByShoppingListID(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		mockSetup      func(*MockItemService)
		expectedStatus int
		expectedBody   func(*testing.T, interface{})
	}{
		{
			name:   "successfully gets items for shopping list",
			listID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				expectedItems := []*entities.Item{
					{
						ID:       uuid.New(),
						Name:     "Milk",
						Quantity: 2,
						Completed: false,
					},
					{
						ID:       uuid.New(),
						Name:     "Bread",
						Quantity: 1,
						Completed: true,
					},
				}
				m.On("GetItemsByShoppingListID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(expectedItems, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body interface{}) {
				items, ok := body.([]interface{})
				require.True(t, ok)
				assert.Len(t, items, 2)
			},
		},
		{
			name:   "successfully gets empty items list",
			listID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				m.On("GetItemsByShoppingListID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return([]*entities.Item{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body interface{}) {
				items, ok := body.([]interface{})
				require.True(t, ok)
				assert.Empty(t, items)
			},
		},
		{
			name:           "fails with invalid list ID",
			listID:         "invalid-uuid",
			mockSetup:      func(m *MockItemService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body interface{}) {
				bodyMap, ok := body.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Invalid list ID format", bodyMap["error"])
			},
		},
		{
			name:   "fails with internal server error",
			listID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				m.On("GetItemsByShoppingListID", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body interface{}) {
				bodyMap, ok := body.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Failed to retrieve items", bodyMap["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.mockSetup(mockService)

			handler := NewItemHandler(mockService)
			router := setupTestRouter()
			router.GET("/shopping-lists/:listId/items", handler.GetItemsByShoppingListID)

			req := httptest.NewRequest(http.MethodGet, "/shopping-lists/"+tt.listID+"/items", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody interface{}
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.expectedBody(t, responseBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestItemHandler_UpdateItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		requestBody    interface{}
		mockSetup      func(*MockItemService)
		expectedStatus int
		expectedBody   func(*testing.T, map[string]interface{})
	}{
		{
			name:   "successfully updates item",
			itemID: uuid.New().String(),
			requestBody: UpdateItemRequest{
				Name:      "Updated Milk",
				Quantity:  3,
				Completed: true,
			},
			mockSetup: func(m *MockItemService) {
				expectedItem := &entities.Item{
					ID:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Name:      "Updated Milk",
					Quantity:  3,
					Completed: true,
				}
				m.On("UpdateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Updated Milk", 3, true).Return(expectedItem, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Updated Milk", body["name"])
				assert.Equal(t, float64(3), body["quantity"])
				assert.Equal(t, true, body["completed"])
			},
		},
		{
			name:   "updates item with default quantity when zero",
			itemID: uuid.New().String(),
			requestBody: UpdateItemRequest{
				Name:      "Test Item",
				Quantity:  0,
				Completed: false,
			},
			mockSetup: func(m *MockItemService) {
				expectedItem := &entities.Item{
					ID:        uuid.New(),
					Name:      "Test Item",
					Quantity:  1,
					Completed: false,
				}
				m.On("UpdateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Test Item", 1, false).Return(expectedItem, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Test Item", body["name"])
				assert.Equal(t, float64(1), body["quantity"])
			},
		},
		{
			name:           "fails with invalid UUID",
			itemID:         "invalid-uuid",
			requestBody:    UpdateItemRequest{Name: "Test", Quantity: 1},
			mockSetup:      func(m *MockItemService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Invalid ID format", body["error"])
			},
		},
		{
			name:           "fails with missing name",
			itemID:         uuid.New().String(),
			requestBody:    map[string]interface{}{"quantity": 1},
			mockSetup:      func(m *MockItemService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "required")
			},
		},
		{
			name:   "fails with not found error",
			itemID: uuid.New().String(),
			requestBody: UpdateItemRequest{
				Name:     "Test Item",
				Quantity: 1,
			},
			mockSetup: func(m *MockItemService) {
				m.On("UpdateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Test Item", 1, false).Return(nil, entities.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Item not found", body["error"])
			},
		},
		{
			name:   "fails with invalid input error",
			itemID: uuid.New().String(),
			requestBody: UpdateItemRequest{
				Name:     "ValidName",
				Quantity: 1,
			},
			mockSetup: func(m *MockItemService) {
				m.On("UpdateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "ValidName", 1, false).Return(nil, entities.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, entities.ErrInvalidInput.Error(), body["error"])
			},
		},
		{
			name:   "fails with internal server error",
			itemID: uuid.New().String(),
			requestBody: UpdateItemRequest{
				Name:     "Test Item",
				Quantity: 1,
			},
			mockSetup: func(m *MockItemService) {
				m.On("UpdateItem", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Test Item", 1, false).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Failed to update item", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.mockSetup(mockService)

			handler := NewItemHandler(mockService)
			router := setupTestRouter()
			router.PUT("/items/:id", handler.UpdateItem)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/items/"+tt.itemID, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.expectedBody(t, responseBody)
			mockService.AssertExpectations(t)
		})
	}
}

func TestItemHandler_DeleteItem(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		mockSetup      func(*MockItemService)
		expectedStatus int
		expectedBody   func(*testing.T, []byte)
	}{
		{
			name:   "successfully deletes item",
			itemID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				m.On("DeleteItem", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody: func(t *testing.T, body []byte) {
				// NoContent responses typically have empty body
				assert.True(t, len(body) == 0 || string(body) == "null")
			},
		},
		{
			name:           "fails with invalid UUID",
			itemID:         "invalid-uuid",
			mockSetup:      func(m *MockItemService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body []byte) {
				var responseBody map[string]interface{}
				err := json.Unmarshal(body, &responseBody)
				require.NoError(t, err)
				assert.Equal(t, "Invalid ID format", responseBody["error"])
			},
		},
		{
			name:   "fails with not found error",
			itemID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				m.On("DeleteItem", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(entities.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body []byte) {
				var responseBody map[string]interface{}
				err := json.Unmarshal(body, &responseBody)
				require.NoError(t, err)
				assert.Equal(t, "Item not found", responseBody["error"])
			},
		},
		{
			name:   "fails with internal server error",
			itemID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				m.On("DeleteItem", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body []byte) {
				var responseBody map[string]interface{}
				err := json.Unmarshal(body, &responseBody)
				require.NoError(t, err)
				assert.Equal(t, "Failed to delete item", responseBody["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.mockSetup(mockService)

			handler := NewItemHandler(mockService)
			router := setupTestRouter()
			router.DELETE("/items/:id", handler.DeleteItem)

			req := httptest.NewRequest(http.MethodDelete, "/items/"+tt.itemID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.expectedBody(t, w.Body.Bytes())
			mockService.AssertExpectations(t)
		})
	}
}

func TestItemHandler_ToggleItemCompletion(t *testing.T) {
	tests := []struct {
		name           string
		itemID         string
		mockSetup      func(*MockItemService)
		expectedStatus int
		expectedBody   func(*testing.T, map[string]interface{})
	}{
		{
			name:   "successfully toggles item completion",
			itemID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				expectedItem := &entities.Item{
					ID:        uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Name:      "Test Item",
					Quantity:  1,
					Completed: true,
				}
				m.On("ToggleItemCompletion", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(expectedItem, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Test Item", body["name"])
				assert.Equal(t, true, body["completed"])
			},
		},
		{
			name:           "fails with invalid UUID",
			itemID:         "invalid-uuid",
			mockSetup:      func(m *MockItemService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Invalid ID format", body["error"])
			},
		},
		{
			name:   "fails with not found error",
			itemID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				m.On("ToggleItemCompletion", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, entities.ErrItemNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Item not found", body["error"])
			},
		},
		{
			name:   "fails with internal server error",
			itemID: uuid.New().String(),
			mockSetup: func(m *MockItemService) {
				m.On("ToggleItemCompletion", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Failed to toggle item completion", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockItemService{}
			tt.mockSetup(mockService)

			handler := NewItemHandler(mockService)
			router := setupTestRouter()
			router.PATCH("/items/:id/toggle", handler.ToggleItemCompletion)

			req := httptest.NewRequest(http.MethodPatch, "/items/"+tt.itemID+"/toggle", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var responseBody map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			tt.expectedBody(t, responseBody)
			mockService.AssertExpectations(t)
		})
	}
}
