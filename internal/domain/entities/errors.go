package entities

import "errors"

var (
	ErrShoppingListNotFound = errors.New("shopping list not found")
	ErrItemNotFound         = errors.New("item not found")
	ErrInvalidInput         = errors.New("invalid input")
	ErrDuplicateItem        = errors.New("item already exists")
)
