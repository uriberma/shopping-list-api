// Package entities contains the core domain entities for the shopping list application.
package entities

import (
	"time"

	"github.com/google/uuid"
)

// ShoppingList represents the main aggregate root
type ShoppingList struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	Name        string    `json:"name" gorm:"not null"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	Items       []Item    `json:"items" gorm:"foreignKey:ShoppingListID;constraint:OnDelete:CASCADE"`
}

// NewShoppingList creates a new shopping list
func NewShoppingList(name, description string) *ShoppingList {
	return &ShoppingList{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		Items:       make([]Item, 0),
	}
}

// AddItem adds an item to the shopping list
func (sl *ShoppingList) AddItem(item *Item) {
	item.ShoppingListID = sl.ID
	sl.Items = append(sl.Items, *item)
}

// RemoveItem removes an item from the shopping list
func (sl *ShoppingList) RemoveItem(itemID uuid.UUID) {
	for i, item := range sl.Items {
		if item.ID == itemID {
			sl.Items = append(sl.Items[:i], sl.Items[i+1:]...)
			break
		}
	}
}

// GetItem returns an item by ID
func (sl *ShoppingList) GetItem(itemID uuid.UUID) *Item {
	for _, item := range sl.Items {
		if item.ID == itemID {
			return &item
		}
	}
	return nil
}

// UpdateItem updates an existing item
func (sl *ShoppingList) UpdateItem(itemID uuid.UUID, name string, quantity int, completed bool) error {
	for i := range sl.Items {
		if sl.Items[i].ID == itemID {
			sl.Items[i].Name = name
			sl.Items[i].Quantity = quantity
			sl.Items[i].Completed = completed
			return nil
		}
	}
	return ErrItemNotFound
}
