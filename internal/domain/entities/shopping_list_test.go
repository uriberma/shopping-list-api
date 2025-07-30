package entities

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewShoppingList(t *testing.T) {
	tests := []struct {
		name        string
		listName    string
		description string
		wantName    string
		wantDesc    string
	}{
		{
			name:        "creates shopping list with name and description",
			listName:    "Grocery List",
			description: "Weekly grocery shopping",
			wantName:    "Grocery List",
			wantDesc:    "Weekly grocery shopping",
		},
		{
			name:        "creates shopping list with empty description",
			listName:    "Quick List",
			description: "",
			wantName:    "Quick List",
			wantDesc:    "",
		},
		{
			name:        "creates shopping list with special characters",
			listName:    "José's List 🛒",
			description: "Special chars & symbols!",
			wantName:    "José's List 🛒",
			wantDesc:    "Special chars & symbols!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			list := NewShoppingList(tt.listName, tt.description)

			assert.NotNil(t, list)
			assert.NotEqual(t, uuid.Nil, list.ID)
			assert.Equal(t, tt.wantName, list.Name)
			assert.Equal(t, tt.wantDesc, list.Description)
			assert.NotNil(t, list.Items)
			assert.Empty(t, list.Items)
			// Note: CreatedAt and UpdatedAt are set by GORM, so they'll be zero in unit tests
			assert.True(t, list.CreatedAt.IsZero())
			assert.True(t, list.UpdatedAt.IsZero())
		})
	}
}

func TestShoppingList_AddItem(t *testing.T) {
	list := NewShoppingList("Test List", "Test Description")
	item1 := NewItem("Milk", 2)
	item2 := NewItem("Bread", 1)

	// Test adding first item
	list.AddItem(item1)
	assert.Len(t, list.Items, 1)
	assert.Equal(t, list.ID, list.Items[0].ShoppingListID)
	assert.Equal(t, "Milk", list.Items[0].Name)
	assert.Equal(t, 2, list.Items[0].Quantity)

	// Test adding second item
	list.AddItem(item2)
	assert.Len(t, list.Items, 2)
	assert.Equal(t, list.ID, list.Items[1].ShoppingListID)
	assert.Equal(t, "Bread", list.Items[1].Name)
	assert.Equal(t, 1, list.Items[1].Quantity)
}

func TestShoppingList_RemoveItem(t *testing.T) {
	list := NewShoppingList("Test List", "Test Description")
	item1 := NewItem("Milk", 2)
	item2 := NewItem("Bread", 1)
	item3 := NewItem("Eggs", 12)

	// Add items
	list.AddItem(item1)
	list.AddItem(item2)
	list.AddItem(item3)
	require.Len(t, list.Items, 3)

	// Test removing middle item
	list.RemoveItem(item2.ID)
	assert.Len(t, list.Items, 2)
	assert.Equal(t, "Milk", list.Items[0].Name)
	assert.Equal(t, "Eggs", list.Items[1].Name)

	// Test removing first item
	list.RemoveItem(item1.ID)
	assert.Len(t, list.Items, 1)
	assert.Equal(t, "Eggs", list.Items[0].Name)

	// Test removing last item
	list.RemoveItem(item3.ID)
	assert.Empty(t, list.Items)

	// Test removing non-existent item (should not panic)
	nonExistentID := uuid.New()
	list.RemoveItem(nonExistentID)
	assert.Empty(t, list.Items)
}

func TestShoppingList_GetItem(t *testing.T) {
	list := NewShoppingList("Test List", "Test Description")
	item1 := NewItem("Milk", 2)
	item2 := NewItem("Bread", 1)

	list.AddItem(item1)
	list.AddItem(item2)

	// Test getting existing item
	foundItem := list.GetItem(item1.ID)
	assert.NotNil(t, foundItem)
	assert.Equal(t, item1.ID, foundItem.ID)
	assert.Equal(t, "Milk", foundItem.Name)
	assert.Equal(t, 2, foundItem.Quantity)

	// Test getting another existing item
	foundItem2 := list.GetItem(item2.ID)
	assert.NotNil(t, foundItem2)
	assert.Equal(t, item2.ID, foundItem2.ID)
	assert.Equal(t, "Bread", foundItem2.Name)

	// Test getting non-existent item
	nonExistentID := uuid.New()
	foundItem3 := list.GetItem(nonExistentID)
	assert.Nil(t, foundItem3)
}

func TestShoppingList_UpdateItem(t *testing.T) {
	list := NewShoppingList("Test List", "Test Description")
	item := NewItem("Milk", 2)
	list.AddItem(item)

	tests := []struct {
		name         string
		itemID       uuid.UUID
		newName      string
		newQuantity  int
		newCompleted bool
		wantErr      bool
		expectedErr  error
	}{
		{
			name:         "successfully updates existing item",
			itemID:       item.ID,
			newName:      "Whole Milk",
			newQuantity:  3,
			newCompleted: true,
			wantErr:      false,
		},
		{
			name:         "updates item with different values",
			itemID:       item.ID,
			newName:      "Skim Milk",
			newQuantity:  1,
			newCompleted: false,
			wantErr:      false,
		},
		{
			name:         "fails to update non-existent item",
			itemID:       uuid.New(),
			newName:      "Non-existent",
			newQuantity:  1,
			newCompleted: false,
			wantErr:      true,
			expectedErr:  ErrItemNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := list.UpdateItem(tt.itemID, tt.newName, tt.newQuantity, tt.newCompleted)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
				updatedItem := list.GetItem(tt.itemID)
				assert.NotNil(t, updatedItem)
				assert.Equal(t, tt.newName, updatedItem.Name)
				assert.Equal(t, tt.newQuantity, updatedItem.Quantity)
				assert.Equal(t, tt.newCompleted, updatedItem.Completed)
			}
		})
	}
}

func TestShoppingList_Integration(t *testing.T) {
	// Test a complete workflow
	list := NewShoppingList("Weekly Groceries", "Shopping for the week")

	// Add multiple items
	milk := NewItem("Milk", 2)
	bread := NewItem("Bread", 1)
	eggs := NewItem("Eggs", 12)

	list.AddItem(milk)
	list.AddItem(bread)
	list.AddItem(eggs)

	assert.Len(t, list.Items, 3)

	// Update an item
	err := list.UpdateItem(milk.ID, "Organic Milk", 3, true)
	assert.NoError(t, err)

	updatedMilk := list.GetItem(milk.ID)
	assert.Equal(t, "Organic Milk", updatedMilk.Name)
	assert.Equal(t, 3, updatedMilk.Quantity)
	assert.True(t, updatedMilk.Completed)

	// Remove an item
	list.RemoveItem(bread.ID)
	assert.Len(t, list.Items, 2)
	assert.Nil(t, list.GetItem(bread.ID))

	// Verify remaining items
	assert.NotNil(t, list.GetItem(milk.ID))
	assert.NotNil(t, list.GetItem(eggs.ID))
}
