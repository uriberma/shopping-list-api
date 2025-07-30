package entities

import (
	"time"

	"github.com/google/uuid"
)

// Item represents an item in a shopping list
type Item struct {
	ID             uuid.UUID `json:"id" gorm:"type:uuid;primary_key"`
	ShoppingListID uuid.UUID `json:"shopping_list_id" gorm:"type:uuid;not null"`
	Name           string    `json:"name" gorm:"not null"`
	Quantity       int       `json:"quantity" gorm:"default:1"`
	Completed      bool      `json:"completed" gorm:"default:false"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// NewItem creates a new item
func NewItem(name string, quantity int) *Item {
	return &Item{
		ID:        uuid.New(),
		Name:      name,
		Quantity:  quantity,
		Completed: false,
	}
}

// MarkCompleted marks the item as completed
func (i *Item) MarkCompleted() {
	i.Completed = true
}

// MarkIncomplete marks the item as incomplete
func (i *Item) MarkIncomplete() {
	i.Completed = false
}

// UpdateQuantity updates the item quantity
func (i *Item) UpdateQuantity(quantity int) {
	i.Quantity = quantity
}
