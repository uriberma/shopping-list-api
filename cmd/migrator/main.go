package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func main() {
	var (
		action         = flag.String("action", "", "Migration action: up, down, version, force, drop")
		databaseURL    = flag.String("database-url", "", "Database URL (optional, will use env vars if not provided)")
		migrationsPath = flag.String("migrations-path", "./cmd/migrator/migrations", "Path to migrations directory")
		forceVersion   = flag.Int("force-version", -1, "Version to force migration to (used with force action)")
	)
	flag.Parse()

	// Get database URL from flag or environment variables
	dbURL := getDatabaseURL(*databaseURL)
	if dbURL == "" {
		log.Fatal("Database URL is required. Set DB_* environment variables or use -database-url flag")
	}

	// Get migrations path
	migrPath := getMigrationsPath(*migrationsPath)

	// Create migrator config
	config := Config{
		DatabaseURL:    dbURL,
		MigrationsPath: migrPath,
	}

	// Create migrator
	migrator, err := NewMigrator(config)
	if err != nil {
		log.Fatalf("Failed to create migrator: %v", err)
	}
	defer func() {
		if err := migrator.Close(); err != nil {
			log.Printf("Failed to close migrator: %v", err)
		}
	}()

	// Execute action
	switch *action {
	case "up":
		if err := migrator.Up(); err != nil {
			log.Fatalf("Failed to run migrations: %v", err)
		}
	case "down":
		if err := migrator.Down(); err != nil {
			log.Fatalf("Failed to rollback migration: %v", err)
		}
	case "version":
		version, dirty, err := migrator.Version()
		if err != nil {
			log.Fatalf("Failed to get migration version: %v", err)
		}
		if version == 0 {
			fmt.Println("No migrations have been applied")
		} else {
			fmt.Printf("Current migration version: %d", version)
			if dirty {
				fmt.Print(" (dirty)")
			}
			fmt.Println()
		}
	case "force":
		if *forceVersion < 0 {
			log.Fatal("Force version must be specified with -force-version flag")
		}
		if err := migrator.Force(*forceVersion); err != nil {
			log.Fatalf("Failed to force migration version: %v", err)
		}
	case "drop":
		fmt.Print("Are you sure you want to drop all database tables? (y/N): ")
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			fmt.Println("Operation canceled")
			return
		}
		if err := migrator.Drop(); err != nil {
			log.Fatalf("Failed to drop database: %v", err)
		}
	default:
		log.Fatalf("Unknown action: %s. Available actions: up, down, version, force, drop", *action)
	}
}

// getDatabaseURL constructs database URL from flag or environment variables
func getDatabaseURL(flagURL string) string {
	if flagURL != "" {
		return flagURL
	}

	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	password := getEnv("DB_PASSWORD", "password")
	dbname := getEnv("DB_NAME", "shopping_list_db")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}

// getMigrationsPath returns the migrations directory path
func getMigrationsPath(flagPath string) string {
	if flagPath != "" {
		return flagPath
	}

	// Default to migrations directory relative to project root
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// Look for migrations directory
	migrationsPath := filepath.Join(wd, "migrations")
	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// Try parent directory (in case we're in cmd/migrator)
		migrationsPath = filepath.Join(wd, "..", "..", "migrations")
		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			log.Fatalf("Migrations directory not found. Use -migrations-path flag to specify location")
		}
	}

	return migrationsPath
}

// getEnv gets an environment variable with a fallback value
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
