package persistence

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.ShoppingList{}, &entities.Item{})
	require.NoError(t, err)

	return db
}

func TestPostgresShoppingListRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresShoppingListRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		list    *entities.ShoppingList
		wantErr bool
	}{
		{
			name: "successful creation",
			list: &entities.ShoppingList{
				ID:          uuid.New(),
				Name:        "Test List",
				Description: "Test Description",
			},
			wantErr: false,
		},
		{
			name: "creation with empty name should succeed", // GORM doesn't enforce this at DB level
			list: &entities.ShoppingList{
				ID:          uuid.New(),
				Name:        "",
				Description: "Test Description",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.list)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.list.ID)
			}
		})
	}
}

func TestPostgresShoppingListRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresShoppingListRepository(db)
	ctx := context.Background()

	// Create a test shopping list
	testList := &entities.ShoppingList{
		ID:          uuid.New(),
		Name:        "Test List",
		Description: "Test Description",
	}
	err := repo.Create(ctx, testList)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *entities.ShoppingList
		wantErr error
	}{
		{
			name: "existing list",
			id:   testList.ID,
			want: testList,
		},
		{
			name:    "non-existing list",
			id:      uuid.New(),
			want:    nil,
			wantErr: entities.ErrShoppingListNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.GetByID(ctx, tt.id)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				assert.Equal(t, tt.want.ID, got.ID)
				assert.Equal(t, tt.want.Name, got.Name)
				assert.Equal(t, tt.want.Description, got.Description)
			}
		})
	}
}

func TestPostgresShoppingListRepository_GetAll(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresShoppingListRepository(db)
	ctx := context.Background()

	// Create test shopping lists
	testLists := []*entities.ShoppingList{
		{
			ID:          uuid.New(),
			Name:        "List 1",
			Description: "Description 1",
		},
		{
			ID:          uuid.New(),
			Name:        "List 2",
			Description: "Description 2",
		},
	}

	for _, list := range testLists {
		err := repo.Create(ctx, list)
		require.NoError(t, err)
	}

	got, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, got, 2)

	// Check that all created lists are returned
	gotIDs := make(map[uuid.UUID]bool)
	for _, list := range got {
		gotIDs[list.ID] = true
	}

	for _, expectedList := range testLists {
		assert.True(t, gotIDs[expectedList.ID], "Expected list %s not found in results", expectedList.ID)
	}
}

func TestPostgresShoppingListRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresShoppingListRepository(db)
	ctx := context.Background()

	// Create a test shopping list
	testList := &entities.ShoppingList{
		ID:          uuid.New(),
		Name:        "Original Name",
		Description: "Original Description",
	}
	err := repo.Create(ctx, testList)
	require.NoError(t, err)

	// Update the list
	testList.Name = "Updated Name"
	testList.Description = "Updated Description"

	err = repo.Update(ctx, testList)
	assert.NoError(t, err)

	// Verify the update
	updated, err := repo.GetByID(ctx, testList.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, "Updated Description", updated.Description)
}

func TestPostgresShoppingListRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewPostgresShoppingListRepository(db)
	ctx := context.Background()

	// Create a test shopping list
	testList := &entities.ShoppingList{
		ID:          uuid.New(),
		Name:        "Test List",
		Description: "Test Description",
	}
	err := repo.Create(ctx, testList)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name: "delete existing list",
			id:   testList.ID,
		},
		{
			name:    "delete non-existing list",
			id:      uuid.New(),
			wantErr: entities.ErrShoppingListNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Delete(ctx, tt.id)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr, err)
			} else {
				assert.NoError(t, err)

				// Verify the list is deleted
				_, err := repo.GetByID(ctx, tt.id)
				assert.Equal(t, entities.ErrShoppingListNotFound, err)
			}
		})
	}
}

func TestPostgresShoppingListRepository_GetByID_DatabaseError(t *testing.T) {
	// Test with a closed database connection to trigger database errors
	db := setupTestDB(t)
	repo := NewPostgresShoppingListRepository(db)
	ctx := context.Background()

	// Close the database connection to simulate database errors
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	// This should trigger a database error (not record not found)
	_, err = repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.NotEqual(t, entities.ErrShoppingListNotFound, err)
}

func TestPostgresShoppingListRepository_Delete_DatabaseError(t *testing.T) {
	// Test with a closed database connection to trigger database errors
	db := setupTestDB(t)
	repo := NewPostgresShoppingListRepository(db)
	ctx := context.Background()

	// Close the database connection to simulate database errors
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	// This should trigger a database error
	err = repo.Delete(ctx, uuid.New())
	assert.Error(t, err)
	assert.NotEqual(t, entities.ErrShoppingListNotFound, err)
}
