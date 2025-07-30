package persistence

import (
	"context"

	"github.com/google/uuid"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
	"github.com/uriberma/go-shopping-list-api/internal/domain/repositories"
	"gorm.io/gorm"
)

// PostgresShoppingListRepository implements the ShoppingListRepository interface
type PostgresShoppingListRepository struct {
	db *gorm.DB
}

// NewPostgresShoppingListRepository creates a new PostgreSQL shopping list repository
func NewPostgresShoppingListRepository(db *gorm.DB) repositories.ShoppingListRepository {
	return &PostgresShoppingListRepository{db: db}
}

// Create creates a new shopping list
func (r *PostgresShoppingListRepository) Create(ctx context.Context, list *entities.ShoppingList) error {
	return r.db.WithContext(ctx).Create(list).Error
}

// GetByID retrieves a shopping list by ID
func (r *PostgresShoppingListRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.ShoppingList, error) {
	var list entities.ShoppingList
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&list).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, entities.ErrShoppingListNotFound
		}
		return nil, err
	}
	return &list, nil
}

// GetAll retrieves all shopping lists
func (r *PostgresShoppingListRepository) GetAll(ctx context.Context) ([]*entities.ShoppingList, error) {
	var lists []*entities.ShoppingList
	err := r.db.WithContext(ctx).Find(&lists).Error
	return lists, err
}

// Update updates an existing shopping list
func (r *PostgresShoppingListRepository) Update(ctx context.Context, list *entities.ShoppingList) error {
	return r.db.WithContext(ctx).Save(list).Error
}

// Delete deletes a shopping list
func (r *PostgresShoppingListRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result := r.db.WithContext(ctx).Where("id = ?", id).Delete(&entities.ShoppingList{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return entities.ErrShoppingListNotFound
	}
	return nil
}
