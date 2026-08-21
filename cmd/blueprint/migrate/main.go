package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/eduardo/blueprint/internal/infrastructure"
	"github.com/eduardo/blueprint/internal/migration"
	"github.com/eduardo/blueprint/internal/parser"
)

func main() {
	f, _ := os.Create("debug_log.txt")
	defer f.Close()
	log.SetOutput(f)

	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		runMigrations()
	case "down":
		rollbackMigrations()
	case "status":
		getMigrationStatus()
	case "generate":
		generateMigrations()
	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Blueprint Migration CLI")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  blueprint migrate up          Run pending migrations")
	fmt.Println("  blueprint migrate down [N]    Rollback N migrations (default: 1)")
	fmt.Println("  blueprint migrate status      Show migration status")
	fmt.Println("  blueprint migrate generate   Generate migration files from blueprint")
	fmt.Println("")
	fmt.Println("Environment variables:")
	fmt.Println("  DATABASE_URL    Database connection URL")
	fmt.Println("  MIGRATIONS_DIR  Directory for migration files (default: ./migrations)")
}

func getDBURL() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		fmt.Println("Error: DATABASE_URL environment variable is required")
		os.Exit(1)
	}
	return url
}

func getMigrationsDir() string {
	dir := os.Getenv("MIGRATIONS_DIR")
	if dir == "" {
		return "./migrations"
	}
	return dir
}

func getDatabaseType() string {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return "postgresql"
	}
	if len(url) > 10 {
		if url[:10] == "mongodb://" {
			return "mongodb"
		}
	}
	return "postgresql"
}

func runMigrations() {
	dbURL := getDBURL()
	migrationsDir := getMigrationsDir()
	dbType := getDatabaseType()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var migrator migration.Migrator
	switch dbType {
	case "mongodb":
		migrator = migration.NewMongoMigrator(dbURL, migrationsDir)
	default:
		migrator = migration.NewPostgresMigrator(dbURL, migrationsDir)
	}

	if err := migrator.RunMigrations(ctx); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	fmt.Println("Migrations completed successfully")
}

func rollbackMigrations() {
	dbURL := getDBURL()
	migrationsDir := getMigrationsDir()
	dbType := getDatabaseType()

	steps := 1
	if len(os.Args) > 2 {
		if s, err := strconv.Atoi(os.Args[2]); err == nil && s > 0 {
			steps = s
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var migrator migration.Migrator
	switch dbType {
	case "mongodb":
		migrator = migration.NewMongoMigrator(dbURL, migrationsDir)
	default:
		migrator = migration.NewPostgresMigrator(dbURL, migrationsDir)
	}

	if err := migrator.Rollback(ctx, steps); err != nil {
		log.Fatalf("Failed to rollback migrations: %v", err)
	}

	fmt.Printf("Rolled back %d migration(s) successfully\n", steps)
}

func getMigrationStatus() {
	dbURL := getDBURL()
	migrationsDir := getMigrationsDir()
	dbType := getDatabaseType()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var migrator migration.Migrator
	switch dbType {
	case "mongodb":
		migrator = migration.NewMongoMigrator(dbURL, migrationsDir)
	default:
		migrator = migration.NewPostgresMigrator(dbURL, migrationsDir)
	}

	migrations, err := migrator.GetAppliedMigrations(ctx)
	if err != nil {
		log.Fatalf("Failed to get migration status: %v", err)
	}

	if len(migrations) == 0 {
		fmt.Println("No migrations have been applied yet")
		return
	}

	fmt.Println("Applied migrations:")
	for _, m := range migrations {
		appliedAt := ""
		if m.AppliedAt != nil {
			appliedAt = m.AppliedAt.Format("2006-01-02 15:04:05")
		}
		fmt.Printf("  %s - %s (applied: %s)\n", m.Version, m.Name, appliedAt)
	}
}

func generateMigrations() {
	if len(os.Args) < 3 {
		fmt.Println("Error: blueprint file path is required")
		fmt.Println("Usage: blueprint migrate generate <blueprint.md>")
		os.Exit(1)
	}

	blueprintPath := os.Args[2]
	outputDir := "./migrations"
	if len(os.Args) > 3 {
		outputDir = os.Args[3]
	}

	fs := infrastructure.NewOSFileSystem()
	markdownParser := parser.NewMarkdownParser(fs)

	config, err := markdownParser.Parse(blueprintPath)
	if err != nil {
		log.Fatalf("Failed to parse blueprint: %v", err)
	}

	var migrator migration.Migrator
	dbType := getDatabaseType()
	dbURL := os.Getenv("DATABASE_URL")

	switch dbType {
	case "mongodb":
		migrator = migration.NewMongoMigrator(dbURL, outputDir)
	default:
		migrator = migration.NewPostgresMigrator(dbURL, outputDir)
	}

	if err := migrator.GenerateMigrationFiles(config, outputDir, fs); err != nil {
		log.Fatalf("Failed to generate migrations: %v", err)
	}

	fmt.Printf("Generated migration files in %s\n", outputDir)
}
