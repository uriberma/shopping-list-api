package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
	"github.com/uriberma/go-shopping-list-api/internal/domain/repositories"
	"gorm.io/gorm"
)

// PostgresItemRepository implements the ItemRepository interface
type PostgresItemRepository struct {
	db *gorm.DB
}

// NewPostgresItemRepository creates a new PostgreSQL item repository
func NewPostgresItemRepository(db *gorm.DB) repositories.ItemRepository {
	return &PostgresItemRepository{db: db}
}

// Create creates a new item
func (r *PostgresItemRepository) Create(ctx context.Context, item *entities.Item) error {
	return r.db.WithContext(ctx).Create(item).Error
}

// GetByID retrieves an item by ID
func (r *PostgresItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.Item, error) {
	var item entities.Item
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrItemNotFound
		}
		return nil, err
	}
	return &item, nil
}

// GetByShoppingListID retrieves all items for a shopping list
func (r *PostgresItemRepository) GetByShoppingListID(
	ctx context.Context,
	shoppingListID uuid.UUID,
) ([]*entities.Item, error) {
	var items []*entities.Item
	err := r.db.WithContext(ctx).Where("shopping_list_id = ?", shoppingListID).Find(&items).Error
	return items, err
}

// Update updates an existing item
func (r *PostgresItemRepository) Update(ctx context.Context, item *entities.Item) error {
	return r.db.WithContext(ctx).Save(item).Error
}

// Delete deletes an item
func (r *PostgresItemRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entities.Item{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entities.ErrItemNotFound
	}
	return nil
}
