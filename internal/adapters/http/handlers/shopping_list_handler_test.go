package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/uriberma/go-shopping-list-api/internal/application/services"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
)

// MockShoppingListService is a mock implementation of the shopping list service interface
type MockShoppingListService struct {
	mock.Mock
}

// Ensure MockShoppingListService implements the interface
var _ services.ShoppingListServiceInterface = (*MockShoppingListService)(nil)

func (m *MockShoppingListService) CreateShoppingList(
	ctx context.Context,
	name, description string,
) (*entities.ShoppingList, error) {
	args := m.Called(ctx, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ShoppingList), args.Error(1)
}

func (m *MockShoppingListService) GetShoppingList(ctx context.Context, id uuid.UUID) (*entities.ShoppingList, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ShoppingList), args.Error(1)
}

func (m *MockShoppingListService) GetAllShoppingLists(ctx context.Context) ([]*entities.ShoppingList, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.ShoppingList), args.Error(1)
}

func (m *MockShoppingListService) UpdateShoppingList(
	ctx context.Context,
	id uuid.UUID,
	name, description string,
) (*entities.ShoppingList, error) {
	args := m.Called(ctx, id, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.ShoppingList), args.Error(1)
}

func (m *MockShoppingListService) DeleteShoppingList(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewShoppingListHandler(t *testing.T) {
	mockService := &MockShoppingListService{}
	handler := NewShoppingListHandler(mockService)

	assert.NotNil(t, handler)
	assert.Equal(t, mockService, handler.service)
}

func TestShoppingListHandler_CreateShoppingList(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockShoppingListService)
		expectedStatus int
		expectedBody   func(*testing.T, map[string]interface{})
	}{
		{
			name: "successfully creates shopping list",
			requestBody: CreateShoppingListRequest{
				Name:        "Grocery List",
				Description: "Weekly groceries",
			},
			mockSetup: func(m *MockShoppingListService) {
				expectedList := &entities.ShoppingList{
					ID:          uuid.New(),
					Name:        "Grocery List",
					Description: "Weekly groceries",
					Items:       []entities.Item{},
				}
				m.On("CreateShoppingList", mock.Anything, "Grocery List", "Weekly groceries").Return(expectedList, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Grocery List", body["name"])
				assert.Equal(t, "Weekly groceries", body["description"])
				assert.NotNil(t, body["id"])
			},
		},
		{
			name: "creates shopping list with empty description",
			requestBody: CreateShoppingListRequest{
				Name:        "Quick List",
				Description: "",
			},
			mockSetup: func(m *MockShoppingListService) {
				expectedList := &entities.ShoppingList{
					ID:          uuid.New(),
					Name:        "Quick List",
					Description: "",
					Items:       []entities.Item{},
				}
				m.On("CreateShoppingList", mock.Anything, "Quick List", "").Return(expectedList, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Quick List", body["name"])
				assert.Equal(t, "", body["description"])
			},
		},
		{
			name:           "fails with missing name",
			requestBody:    map[string]interface{}{"description": "Test"},
			mockSetup:      func(m *MockShoppingListService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "required")
			},
		},
		{
			name:           "fails with invalid JSON",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockShoppingListService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.NotNil(t, body["error"])
			},
		},
		{
			name: "fails with invalid input error from service",
			requestBody: CreateShoppingListRequest{
				Name:        "ValidName",
				Description: "Test",
			},
			mockSetup: func(m *MockShoppingListService) {
				m.On("CreateShoppingList", mock.Anything, "ValidName", "Test").Return(nil, entities.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, entities.ErrInvalidInput.Error(), body["error"])
			},
		},
		{
			name: "fails with internal server error",
			requestBody: CreateShoppingListRequest{
				Name:        "Test List",
				Description: "Test",
			},
			mockSetup: func(m *MockShoppingListService) {
				m.On("CreateShoppingList", mock.Anything, "Test List", "Test").Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Failed to create shopping list", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockShoppingListService{}
			tt.mockSetup(mockService)

			handler := NewShoppingListHandler(mockService)
			router := setupTestRouter()
			router.POST("/lists", handler.CreateShoppingList)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/lists", bytes.NewBuffer(body))
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

func TestShoppingListHandler_GetShoppingList(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		mockSetup      func(*MockShoppingListService)
		expectedStatus int
		expectedBody   func(*testing.T, map[string]interface{})
	}{
		{
			name:   "successfully gets shopping list",
			listID: uuid.New().String(),
			mockSetup: func(m *MockShoppingListService) {
				expectedList := &entities.ShoppingList{
					ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Name:        "Test List",
					Description: "Test Description",
					Items:       []entities.Item{},
				}
				m.On("GetShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(expectedList, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Test List", body["name"])
				assert.Equal(t, "Test Description", body["description"])
			},
		},
		{
			name:           "fails with invalid UUID",
			listID:         "invalid-uuid",
			mockSetup:      func(m *MockShoppingListService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Invalid ID format", body["error"])
			},
		},
		{
			name:   "fails with not found error",
			listID: uuid.New().String(),
			mockSetup: func(m *MockShoppingListService) {
				m.On("GetShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, entities.ErrShoppingListNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Shopping list not found", body["error"])
			},
		},
		{
			name:   "fails with internal server error",
			listID: uuid.New().String(),
			mockSetup: func(m *MockShoppingListService) {
				m.On("GetShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Failed to retrieve shopping list", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockShoppingListService{}
			tt.mockSetup(mockService)

			handler := NewShoppingListHandler(mockService)
			router := setupTestRouter()
			router.GET("/lists/:id", handler.GetShoppingList)

			req := httptest.NewRequest(http.MethodGet, "/lists/"+tt.listID, nil)
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

func TestShoppingListHandler_GetAllShoppingLists(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockShoppingListService)
		expectedStatus int
		expectedBody   func(*testing.T, interface{})
	}{
		{
			name: "successfully gets all shopping lists",
			mockSetup: func(m *MockShoppingListService) {
				expectedLists := []*entities.ShoppingList{
					{
						ID:          uuid.New(),
						Name:        "List 1",
						Description: "Description 1",
						Items:       []entities.Item{},
					},
					{
						ID:          uuid.New(),
						Name:        "List 2",
						Description: "Description 2",
						Items:       []entities.Item{},
					},
				}
				m.On("GetAllShoppingLists", mock.Anything).Return(expectedLists, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body interface{}) {
				lists, ok := body.([]interface{})
				require.True(t, ok)
				assert.Len(t, lists, 2)
			},
		},
		{
			name: "successfully gets empty list",
			mockSetup: func(m *MockShoppingListService) {
				m.On("GetAllShoppingLists", mock.Anything).Return([]*entities.ShoppingList{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body interface{}) {
				lists, ok := body.([]interface{})
				require.True(t, ok)
				assert.Empty(t, lists)
			},
		},
		{
			name: "fails with internal server error",
			mockSetup: func(m *MockShoppingListService) {
				m.On("GetAllShoppingLists", mock.Anything).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body interface{}) {
				bodyMap, ok := body.(map[string]interface{})
				require.True(t, ok)
				assert.Equal(t, "Failed to retrieve shopping lists", bodyMap["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockShoppingListService{}
			tt.mockSetup(mockService)

			handler := NewShoppingListHandler(mockService)
			router := setupTestRouter()
			router.GET("/lists", handler.GetAllShoppingLists)

			req := httptest.NewRequest(http.MethodGet, "/lists", nil)
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

func TestShoppingListHandler_UpdateShoppingList(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		requestBody    interface{}
		mockSetup      func(*MockShoppingListService)
		expectedStatus int
		expectedBody   func(*testing.T, map[string]interface{})
	}{
		{
			name:   "successfully updates shopping list",
			listID: uuid.New().String(),
			requestBody: UpdateShoppingListRequest{
				Name:        "Updated List",
				Description: "Updated Description",
			},
			mockSetup: func(m *MockShoppingListService) {
				expectedList := &entities.ShoppingList{
					ID:          uuid.MustParse("123e4567-e89b-12d3-a456-426614174000"),
					Name:        "Updated List",
					Description: "Updated Description",
					Items:       []entities.Item{},
				}
				m.On("UpdateShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Updated List", "Updated Description").Return(expectedList, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Updated List", body["name"])
				assert.Equal(t, "Updated Description", body["description"])
			},
		},
		{
			name:           "fails with invalid UUID",
			listID:         "invalid-uuid",
			requestBody:    UpdateShoppingListRequest{Name: "Test", Description: "Test"},
			mockSetup:      func(m *MockShoppingListService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Invalid ID format", body["error"])
			},
		},
		{
			name:           "fails with missing name",
			listID:         uuid.New().String(),
			requestBody:    map[string]interface{}{"description": "Test"},
			mockSetup:      func(m *MockShoppingListService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Contains(t, body["error"], "required")
			},
		},
		{
			name:   "fails with not found error",
			listID: uuid.New().String(),
			requestBody: UpdateShoppingListRequest{
				Name:        "Test List",
				Description: "Test",
			},
			mockSetup: func(m *MockShoppingListService) {
				m.On("UpdateShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Test List", "Test").Return(nil, entities.ErrShoppingListNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Shopping list not found", body["error"])
			},
		},
		{
			name:   "fails with invalid input error",
			listID: uuid.New().String(),
			requestBody: UpdateShoppingListRequest{
				Name:        "ValidName",
				Description: "Test",
			},
			mockSetup: func(m *MockShoppingListService) {
				m.On("UpdateShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID"), "ValidName", "Test").Return(nil, entities.ErrInvalidInput)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, entities.ErrInvalidInput.Error(), body["error"])
			},
		},
		{
			name:   "fails with internal server error",
			listID: uuid.New().String(),
			requestBody: UpdateShoppingListRequest{
				Name:        "Test List",
				Description: "Test",
			},
			mockSetup: func(m *MockShoppingListService) {
				m.On("UpdateShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID"), "Test List", "Test").Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body map[string]interface{}) {
				assert.Equal(t, "Failed to update shopping list", body["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockShoppingListService{}
			tt.mockSetup(mockService)

			handler := NewShoppingListHandler(mockService)
			router := setupTestRouter()
			router.PUT("/lists/:id", handler.UpdateShoppingList)

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPut, "/lists/"+tt.listID, bytes.NewBuffer(body))
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

func TestShoppingListHandler_DeleteShoppingList(t *testing.T) {
	tests := []struct {
		name           string
		listID         string
		mockSetup      func(*MockShoppingListService)
		expectedStatus int
		expectedBody   func(*testing.T, []byte)
	}{
		{
			name:   "successfully deletes shopping list",
			listID: uuid.New().String(),
			mockSetup: func(m *MockShoppingListService) {
				m.On("DeleteShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
			expectedBody: func(t *testing.T, body []byte) {
				// NoContent responses typically have empty body
				assert.True(t, len(body) == 0 || string(body) == "null")
			},
		},
		{
			name:           "fails with invalid UUID",
			listID:         "invalid-uuid",
			mockSetup:      func(m *MockShoppingListService) {},
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
			listID: uuid.New().String(),
			mockSetup: func(m *MockShoppingListService) {
				m.On("DeleteShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(entities.ErrShoppingListNotFound)
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: func(t *testing.T, body []byte) {
				var responseBody map[string]interface{}
				err := json.Unmarshal(body, &responseBody)
				require.NoError(t, err)
				assert.Equal(t, "Shopping list not found", responseBody["error"])
			},
		},
		{
			name:   "fails with internal server error",
			listID: uuid.New().String(),
			mockSetup: func(m *MockShoppingListService) {
				m.On("DeleteShoppingList", mock.Anything, mock.AnythingOfType("uuid.UUID")).Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody: func(t *testing.T, body []byte) {
				var responseBody map[string]interface{}
				err := json.Unmarshal(body, &responseBody)
				require.NoError(t, err)
				assert.Equal(t, "Failed to delete shopping list", responseBody["error"])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockShoppingListService{}
			tt.mockSetup(mockService)

			handler := NewShoppingListHandler(mockService)
			router := setupTestRouter()
			router.DELETE("/lists/:id", handler.DeleteShoppingList)

			req := httptest.NewRequest(http.MethodDelete, "/lists/"+tt.listID, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			tt.expectedBody(t, w.Body.Bytes())
			mockService.AssertExpectations(t)
		})
	}
}
