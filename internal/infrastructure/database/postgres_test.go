package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/uriberma/go-shopping-list-api/internal/domain/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestNewPostgresConnection(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "invalid config should return error",
			config: Config{
				Host:     "invalid-host",
				Port:     "invalid-port",
				User:     "invalid-user",
				Password: "invalid-password",
				DBName:   "invalid-db",
				SSLMode:  "disable",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewPostgresConnection(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAutoMigrate(t *testing.T) {
	// Use SQLite in-memory database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	err = AutoMigrate(db)
	assert.NoError(t, err)

	// Verify that tables were created
	assert.True(t, db.Migrator().HasTable(&entities.ShoppingList{}))
	assert.True(t, db.Migrator().HasTable(&entities.Item{}))

	// Verify that we can create records (basic schema validation)
	testList := &entities.ShoppingList{
		Name:        "Test List",
		Description: "Test Description",
	}
	err = db.Create(testList).Error
	assert.NoError(t, err)

	testItem := &entities.Item{
		ShoppingListID: testList.ID,
		Name:           "Test Item",
		Quantity:       1,
		Completed:      false,
	}
	err = db.Create(testItem).Error
	assert.NoError(t, err)
}

func TestAutoMigrate_InvalidDB(t *testing.T) {
	// Test with a nil database (should handle gracefully)
	var db *gorm.DB

	// This should panic, so we need to recover from it
	defer func() {
		if r := recover(); r != nil {
			// Expected panic due to nil database
			assert.NotNil(t, r)
		}
	}()

	err := AutoMigrate(db)
	// If we reach here without panic, it should be an error
	if err == nil {
		t.Error("Expected error or panic with nil database")
	}
}

func TestConfig_Validation(t *testing.T) {
	tests := []struct {
		name   string
		config Config
		valid  bool
	}{
		{
			name: "valid config",
			config: Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "password",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			valid: true,
		},
		{
			name: "empty host",
			config: Config{
				Host:     "",
				Port:     "5432",
				User:     "postgres",
				Password: "password",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			valid: false,
		},
		{
			name: "empty port",
			config: Config{
				Host:     "localhost",
				Port:     "",
				User:     "postgres",
				Password: "password",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			valid: false,
		},
		{
			name: "empty user",
			config: Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "",
				Password: "password",
				DBName:   "testdb",
				SSLMode:  "disable",
			},
			valid: false,
		},
		{
			name: "empty database name",
			config: Config{
				Host:     "localhost",
				Port:     "5432",
				User:     "postgres",
				Password: "password",
				DBName:   "",
				SSLMode:  "disable",
			},
			valid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test that the config fields are properly set
			assert.Equal(t, tt.valid, isValidConfig(tt.config))
		})
	}
}

// Helper function to validate config (this could be added to the actual Config struct)
func isValidConfig(config Config) bool {
	return config.Host != "" &&
		config.Port != "" &&
		config.User != "" &&
		config.DBName != ""
}
