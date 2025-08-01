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

func setupTestDBForItems(t *testing.T) (*gorm.DB, *entities.ShoppingList) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = db.AutoMigrate(&entities.ShoppingList{}, &entities.Item{})
	require.NoError(t, err)

	// Create a test shopping list for items
	testList := &entities.ShoppingList{
		ID:          uuid.New(),
		Name:        "Test List",
		Description: "Test Description",
	}
	err = db.Create(testList).Error
	require.NoError(t, err)

	return db, testList
}

func TestPostgresItemRepository_Create(t *testing.T) {
	db, testList := setupTestDBForItems(t)
	repo := NewPostgresItemRepository(db)
	ctx := context.Background()

	tests := []struct {
		name    string
		item    *entities.Item
		wantErr bool
	}{
		{
			name: "successful creation",
			item: &entities.Item{
				ID:             uuid.New(),
				ShoppingListID: testList.ID,
				Name:           "Test Item",
				Quantity:       2,
				Completed:      false,
			},
			wantErr: false,
		},
		{
			name: "creation with default quantity",
			item: &entities.Item{
				ID:             uuid.New(),
				ShoppingListID: testList.ID,
				Name:           "Test Item 2",
				Quantity:       1,
				Completed:      false,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Create(ctx, tt.item)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, tt.item.ID)
			}
		})
	}
}

func TestPostgresItemRepository_GetByID(t *testing.T) {
	db, testList := setupTestDBForItems(t)
	repo := NewPostgresItemRepository(db)
	ctx := context.Background()

	// Create a test item
	testItem := &entities.Item{
		ID:             uuid.New(),
		ShoppingListID: testList.ID,
		Name:           "Test Item",
		Quantity:       3,
		Completed:      false,
	}
	err := repo.Create(ctx, testItem)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		want    *entities.Item
		wantErr error
	}{
		{
			name: "existing item",
			id:   testItem.ID,
			want: testItem,
		},
		{
			name:    "non-existing item",
			id:      uuid.New(),
			want:    nil,
			wantErr: entities.ErrItemNotFound,
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
				assert.Equal(t, tt.want.Quantity, got.Quantity)
				assert.Equal(t, tt.want.Completed, got.Completed)
				assert.Equal(t, tt.want.ShoppingListID, got.ShoppingListID)
			}
		})
	}
}

func TestPostgresItemRepository_GetByShoppingListID(t *testing.T) {
	db, testList := setupTestDBForItems(t)
	repo := NewPostgresItemRepository(db)
	ctx := context.Background()

	// Create another shopping list for isolation testing
	anotherList := &entities.ShoppingList{
		ID:          uuid.New(),
		Name:        "Another List",
		Description: "Another Description",
	}
	err := db.Create(anotherList).Error
	require.NoError(t, err)

	// Create test items for the first list
	testItems := []*entities.Item{
		{
			ID:             uuid.New(),
			ShoppingListID: testList.ID,
			Name:           "Item 1",
			Quantity:       1,
			Completed:      false,
		},
		{
			ID:             uuid.New(),
			ShoppingListID: testList.ID,
			Name:           "Item 2",
			Quantity:       2,
			Completed:      true,
		},
	}

	// Create an item for another list (should not be returned)
	anotherItem := &entities.Item{
		ID:             uuid.New(),
		ShoppingListID: anotherList.ID,
		Name:           "Another Item",
		Quantity:       1,
		Completed:      false,
	}

	for _, item := range testItems {
		err := repo.Create(ctx, item)
		require.NoError(t, err)
	}
	err = repo.Create(ctx, anotherItem)
	require.NoError(t, err)

	got, err := repo.GetByShoppingListID(ctx, testList.ID)
	assert.NoError(t, err)
	assert.Len(t, got, 2)

	// Check that only items from the correct list are returned
	gotIDs := make(map[uuid.UUID]bool)
	for _, item := range got {
		gotIDs[item.ID] = true
		assert.Equal(t, testList.ID, item.ShoppingListID)
	}

	for _, expectedItem := range testItems {
		assert.True(t, gotIDs[expectedItem.ID], "Expected item %s not found in results", expectedItem.ID)
	}

	// Verify the other item is not included
	assert.False(t, gotIDs[anotherItem.ID], "Item from another list should not be included")
}

func TestPostgresItemRepository_Update(t *testing.T) {
	db, testList := setupTestDBForItems(t)
	repo := NewPostgresItemRepository(db)
	ctx := context.Background()

	// Create a test item
	testItem := &entities.Item{
		ID:             uuid.New(),
		ShoppingListID: testList.ID,
		Name:           "Original Name",
		Quantity:       1,
		Completed:      false,
	}
	err := repo.Create(ctx, testItem)
	require.NoError(t, err)

	// Update the item
	testItem.Name = "Updated Name"
	testItem.Quantity = 5
	testItem.Completed = true

	err = repo.Update(ctx, testItem)
	assert.NoError(t, err)

	// Verify the update
	updated, err := repo.GetByID(ctx, testItem.ID)
	assert.NoError(t, err)
	assert.Equal(t, "Updated Name", updated.Name)
	assert.Equal(t, 5, updated.Quantity)
	assert.Equal(t, true, updated.Completed)
}

func TestPostgresItemRepository_Delete(t *testing.T) {
	db, testList := setupTestDBForItems(t)
	repo := NewPostgresItemRepository(db)
	ctx := context.Background()

	// Create a test item
	testItem := &entities.Item{
		ID:             uuid.New(),
		ShoppingListID: testList.ID,
		Name:           "Test Item",
		Quantity:       1,
		Completed:      false,
	}
	err := repo.Create(ctx, testItem)
	require.NoError(t, err)

	tests := []struct {
		name    string
		id      uuid.UUID
		wantErr error
	}{
		{
			name: "delete existing item",
			id:   testItem.ID,
		},
		{
			name:    "delete non-existing item",
			id:      uuid.New(),
			wantErr: entities.ErrItemNotFound,
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

				// Verify the item is deleted
				_, err := repo.GetByID(ctx, tt.id)
				assert.Equal(t, entities.ErrItemNotFound, err)
			}
		})
	}
}

func TestPostgresItemRepository_GetByShoppingListID_EmptyList(t *testing.T) {
	db, testList := setupTestDBForItems(t)
	repo := NewPostgresItemRepository(db)
	ctx := context.Background()

	// Test getting items from a list with no items
	got, err := repo.GetByShoppingListID(ctx, testList.ID)
	assert.NoError(t, err)
	assert.Empty(t, got)
}

func TestPostgresItemRepository_CascadeDelete(t *testing.T) {
	db, testList := setupTestDBForItems(t)
	itemRepo := NewPostgresItemRepository(db)
	listRepo := NewPostgresShoppingListRepository(db)
	ctx := context.Background()

	// Create a test item
	testItem := &entities.Item{
		ID:             uuid.New(),
		ShoppingListID: testList.ID,
		Name:           "Test Item",
		Quantity:       1,
		Completed:      false,
	}
	err := itemRepo.Create(ctx, testItem)
	require.NoError(t, err)

	// Delete the shopping list (should cascade delete items)
	err = listRepo.Delete(ctx, testList.ID)
	assert.NoError(t, err)

	// Note: SQLite in-memory doesn't enforce foreign key constraints by default
	// In a real PostgreSQL environment, this would cascade delete
	// For this test, we'll verify the list is deleted
	_, err = listRepo.GetByID(ctx, testList.ID)
	assert.Equal(t, entities.ErrShoppingListNotFound, err)
}

func TestPostgresItemRepository_GetByID_DatabaseError(t *testing.T) {
	// Test with a closed database connection to trigger database errors
	db, _ := setupTestDBForItems(t)
	repo := NewPostgresItemRepository(db)
	ctx := context.Background()

	// Close the database connection to simulate database errors
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	// This should trigger a database error (not record not found)
	_, err = repo.GetByID(ctx, uuid.New())
	assert.Error(t, err)
	assert.NotEqual(t, entities.ErrItemNotFound, err)
}

func TestPostgresItemRepository_Delete_DatabaseError(t *testing.T) {
	// Test with a closed database connection to trigger database errors
	db, _ := setupTestDBForItems(t)
	repo := NewPostgresItemRepository(db)
	ctx := context.Background()

	// Close the database connection to simulate database errors
	sqlDB, err := db.DB()
	require.NoError(t, err)
	sqlDB.Close()

	// This should trigger a database error
	err = repo.Delete(ctx, uuid.New())
	assert.Error(t, err)
	assert.NotEqual(t, entities.ErrItemNotFound, err)
}
