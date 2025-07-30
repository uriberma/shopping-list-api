package entities

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewItem(t *testing.T) {
	tests := []struct {
		name         string
		itemName     string
		quantity     int
		wantName     string
		wantQuantity int
	}{
		{
			name:         "creates item with positive quantity",
			itemName:     "Milk",
			quantity:     2,
			wantName:     "Milk",
			wantQuantity: 2,
		},
		{
			name:         "creates item with quantity 1",
			itemName:     "Bread",
			quantity:     1,
			wantName:     "Bread",
			wantQuantity: 1,
		},
		{
			name:         "creates item with zero quantity",
			itemName:     "Sugar",
			quantity:     0,
			wantName:     "Sugar",
			wantQuantity: 0,
		},
		{
			name:         "creates item with special characters",
			itemName:     "José's Coffee ☕",
			quantity:     3,
			wantName:     "José's Coffee ☕",
			wantQuantity: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item := NewItem(tt.itemName, tt.quantity)

			assert.NotNil(t, item)
			assert.NotEqual(t, uuid.Nil, item.ID)
			assert.Equal(t, uuid.Nil, item.ShoppingListID) // Should be nil until added to a list
			assert.Equal(t, tt.wantName, item.Name)
			assert.Equal(t, tt.wantQuantity, item.Quantity)
			assert.False(t, item.Completed) // Should default to false
			// Note: CreatedAt and UpdatedAt are set by GORM, so they'll be zero in unit tests
			assert.True(t, item.CreatedAt.IsZero())
			assert.True(t, item.UpdatedAt.IsZero())
		})
	}
}

func TestItem_MarkCompleted(t *testing.T) {
	item := NewItem("Test Item", 1)

	// Initially should be incomplete
	assert.False(t, item.Completed)

	// Mark as completed
	item.MarkCompleted()
	assert.True(t, item.Completed)

	// Mark as completed again (should remain true)
	item.MarkCompleted()
	assert.True(t, item.Completed)
}

func TestItem_MarkIncomplete(t *testing.T) {
	item := NewItem("Test Item", 1)

	// Mark as completed first
	item.MarkCompleted()
	assert.True(t, item.Completed)

	// Mark as incomplete
	item.MarkIncomplete()
	assert.False(t, item.Completed)

	// Mark as incomplete again (should remain false)
	item.MarkIncomplete()
	assert.False(t, item.Completed)
}

func TestItem_UpdateQuantity(t *testing.T) {
	item := NewItem("Test Item", 1)

	tests := []struct {
		name         string
		newQuantity  int
		wantQuantity int
	}{
		{
			name:         "updates to positive quantity",
			newQuantity:  5,
			wantQuantity: 5,
		},
		{
			name:         "updates to zero quantity",
			newQuantity:  0,
			wantQuantity: 0,
		},
		{
			name:         "updates to negative quantity",
			newQuantity:  -1,
			wantQuantity: -1,
		},
		{
			name:         "updates to large quantity",
			newQuantity:  1000,
			wantQuantity: 1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			item.UpdateQuantity(tt.newQuantity)
			assert.Equal(t, tt.wantQuantity, item.Quantity)
		})
	}
}

func TestItem_CompletionToggle(t *testing.T) {
	item := NewItem("Test Item", 1)

	// Test multiple toggles
	assert.False(t, item.Completed)

	item.MarkCompleted()
	assert.True(t, item.Completed)

	item.MarkIncomplete()
	assert.False(t, item.Completed)

	item.MarkCompleted()
	assert.True(t, item.Completed)

	item.MarkIncomplete()
	assert.False(t, item.Completed)
}

func TestItem_Integration(t *testing.T) {
	// Test a complete workflow with an item
	item := NewItem("Organic Milk", 2)

	// Verify initial state
	assert.Equal(t, "Organic Milk", item.Name)
	assert.Equal(t, 2, item.Quantity)
	assert.False(t, item.Completed)

	// Update quantity
	item.UpdateQuantity(3)
	assert.Equal(t, 3, item.Quantity)

	// Mark as completed
	item.MarkCompleted()
	assert.True(t, item.Completed)

	// Update quantity while completed
	item.UpdateQuantity(1)
	assert.Equal(t, 1, item.Quantity)
	assert.True(t, item.Completed) // Should remain completed

	// Mark as incomplete
	item.MarkIncomplete()
	assert.False(t, item.Completed)
	assert.Equal(t, 1, item.Quantity) // Quantity should remain unchanged
}

func TestItem_UniqueIDs(t *testing.T) {
	// Test that each item gets a unique ID
	item1 := NewItem("Item 1", 1)
	item2 := NewItem("Item 2", 2)
	item3 := NewItem("Item 3", 3)

	assert.NotEqual(t, item1.ID, item2.ID)
	assert.NotEqual(t, item1.ID, item3.ID)
	assert.NotEqual(t, item2.ID, item3.ID)

	// All IDs should be valid UUIDs
	assert.NotEqual(t, uuid.Nil, item1.ID)
	assert.NotEqual(t, uuid.Nil, item2.ID)
	assert.NotEqual(t, uuid.Nil, item3.ID)
}
