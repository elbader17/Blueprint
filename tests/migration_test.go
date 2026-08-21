package tests

import (
	"os"
	"strings"
	"testing"

	"github.com/eduardo/blueprint/internal/domain"
	"github.com/eduardo/blueprint/internal/migration"
)

type mockMigrationFileSystem struct {
	files map[string][]byte
	dirs  map[string]bool
}

func (m *mockMigrationFileSystem) ReadFile(path string) ([]byte, error) {
	if content, ok := m.files[path]; ok {
		return content, nil
	}
	return nil, &mockError{msg: "file not found"}
}

func (m *mockMigrationFileSystem) MkdirAll(path string) error {
	m.dirs[path] = true
	return nil
}

func (m *mockMigrationFileSystem) WriteFile(path string, data []byte) error {
	m.files[path] = data
	return nil
}

func (m *mockMigrationFileSystem) CopyFile(src, dst string) error {
	return nil
}

func (m *mockMigrationFileSystem) Chmod(path string, mode uint32) error {
	return nil
}

func (m *mockMigrationFileSystem) RemoveAll(path string) error {
	m.files = make(map[string][]byte)
	m.dirs = make(map[string]bool)
	return nil
}

func TestPostgresMigrator_GenerateMigrationFiles(t *testing.T) {
	fs := &mockMigrationFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}

	config := &domain.Config{
		ProjectName: "TestAPI",
		Database: domain.Database{
			Type: "postgresql",
			URL:  "postgres://localhost/test",
		},
		Models: []domain.Model{
			{
				Name:      "users",
				Protected: true,
				Fields: map[string]string{
					"email": "string",
					"name":  "string",
				},
				Relations: map[string]string{
					"posts": "hasMany:posts",
				},
			},
			{
				Name:      "posts",
				Protected: true,
				Fields: map[string]string{
					"title": "string",
					"count": "integer",
				},
				Relations: map[string]string{
					"author_id": "belongsTo:users",
				},
			},
		},
	}

	migrator := migration.NewPostgresMigrator("postgres://localhost/test", "/tmp/migrations")
	err := migrator.GenerateMigrationFiles(config, "/tmp/migrations", fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(fs.files) == 0 {
		t.Error("expected migration files to be created")
	}

	hasUp := false
	hasDown := false
	for path := range fs.files {
		if strings.Contains(path, ".up.sql") {
			hasUp = true
		}
		if strings.Contains(path, ".down.sql") {
			hasDown = true
		}
	}

	if !hasUp {
		t.Error("expected .up.sql migration file to be created")
	}
	if !hasDown {
		t.Error("expected .down.sql migration file to be created")
	}
}

func TestPostgresMigrator_GenerateMigrationFiles_CreatesSchemaMigrationsTable(t *testing.T) {
	fs := &mockMigrationFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}

	config := &domain.Config{
		ProjectName: "TestAPI",
		Database: domain.Database{
			Type: "postgresql",
			URL:  "postgres://localhost/test",
		},
		Models: []domain.Model{
			{
				Name:      "users",
				Protected: true,
				Fields: map[string]string{
					"email": "string",
				},
				Relations: map[string]string{},
			},
		},
	}

	migrator := migration.NewPostgresMigrator("postgres://localhost/test", "/tmp/migrations")
	err := migrator.GenerateMigrationFiles(config, "/tmp/migrations", fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for path, content := range fs.files {
		if strings.Contains(path, ".up.sql") {
			if !contains(string(content), "schema_migrations") {
				t.Error("expected up migration to contain schema_migrations table creation")
			}
			if !contains(string(content), "CREATE TABLE IF NOT EXISTS schema_migrations") {
				t.Error("expected schema_migrations table creation SQL")
			}
		}
	}
}

func TestMongoMigrator_GenerateMigrationFiles(t *testing.T) {
	fs := &mockMigrationFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}

	config := &domain.Config{
		ProjectName: "TestAPI",
		Database: domain.Database{
			Type: "mongodb",
			URL:  "mongodb://localhost:27017",
		},
		Models: []domain.Model{
			{
				Name:      "users",
				Protected: true,
				Fields: map[string]string{
					"email": "string",
					"name":  "string",
				},
				Relations: map[string]string{
					"posts": "hasMany:posts",
				},
			},
			{
				Name:      "posts",
				Protected: true,
				Fields: map[string]string{
					"title": "string",
				},
				Relations: map[string]string{
					"author_id": "belongsTo:users",
				},
			},
		},
	}

	migrator := migration.NewMongoMigrator("mongodb://localhost:27017", "/tmp/migrations")
	err := migrator.GenerateMigrationFiles(config, "/tmp/migrations", fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(fs.files) == 0 {
		t.Error("expected migration files to be created")
	}

	hasUp := false
	hasDown := false
	for path := range fs.files {
		if strings.Contains(path, ".up.js") {
			hasUp = true
		}
		if strings.Contains(path, ".down.js") {
			hasDown = true
		}
	}

	if !hasUp {
		t.Error("expected .up.js migration file to be created")
	}
	if !hasDown {
		t.Error("expected .down.js migration file to be created")
	}
}

func TestMongoMigrator_GenerateMigrationFiles_CreatesSchemaMigrationsCollection(t *testing.T) {
	fs := &mockMigrationFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}

	config := &domain.Config{
		ProjectName: "TestAPI",
		Database: domain.Database{
			Type: "mongodb",
			URL:  "mongodb://localhost:27017",
		},
		Models: []domain.Model{
			{
				Name:      "users",
				Protected: true,
				Fields: map[string]string{
					"email": "string",
				},
				Relations: map[string]string{},
			},
		},
	}

	migrator := migration.NewMongoMigrator("mongodb://localhost:27017", "/tmp/migrations")
	err := migrator.GenerateMigrationFiles(config, "/tmp/migrations", fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for path, content := range fs.files {
		if strings.Contains(path, ".up.js") {
			if !contains(string(content), "schema_migrations") {
				t.Error("expected up migration to contain schema_migrations collection creation")
			}
			if !contains(string(content), "createCollection") {
				t.Error("expected createCollection statements")
			}
		}
	}
}

func TestPostgresMigrator_GenerateMigrationFiles_AllFieldTypes(t *testing.T) {
	fs := &mockMigrationFileSystem{
		files: make(map[string][]byte),
		dirs:  make(map[string]bool),
	}

	config := &domain.Config{
		ProjectName: "TestAPI",
		Database: domain.Database{
			Type: "postgresql",
			URL:  "postgres://localhost/test",
		},
		Models: []domain.Model{
			{
				Name:      "products",
				Protected: false,
				Fields: map[string]string{
					"name":        "string",
					"description": "text",
					"price":      "float",
					"quantity":   "integer",
					"in_stock":   "boolean",
					"created_at": "datetime",
				},
				Relations: map[string]string{},
			},
		},
	}

	migrator := migration.NewPostgresMigrator("postgres://localhost/test", "/tmp/migrations")
	err := migrator.GenerateMigrationFiles(config, "/tmp/migrations", fs)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for path, content := range fs.files {
		if strings.Contains(path, ".up.sql") {
			if !contains(string(content), "TEXT") {
				t.Error("expected TEXT type for string/text fields")
			}
			if !contains(string(content), "INTEGER") {
				t.Error("expected INTEGER type for integer fields")
			}
			if !contains(string(content), "DOUBLE PRECISION") {
				t.Error("expected DOUBLE PRECISION type for float fields")
			}
			if !contains(string(content), "BOOLEAN") {
				t.Error("expected BOOLEAN type for boolean fields")
			}
			if !contains(string(content), "TIMESTAMP") {
				t.Error("expected TIMESTAMP type for datetime fields")
			}
		}
	}
}

func TestPostgresMigrator_RunMigrations(t *testing.T) {
	migrator := migration.NewPostgresMigrator("postgres://localhost/test", "/tmp/migrations")
	err := migrator.RunMigrations(nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestPostgresMigrator_Rollback(t *testing.T) {
	migrator := migration.NewPostgresMigrator("postgres://localhost/test", "/tmp/migrations")
	err := migrator.Rollback(nil, 1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestPostgresMigrator_GetAppliedMigrations(t *testing.T) {
	migrator := migration.NewPostgresMigrator("postgres://localhost/test", "/tmp/migrations")
	migrations, err := migrator.GetAppliedMigrations(nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if migrations == nil {
		t.Error("expected migrations to be returned")
	}
}

func TestMongoMigrator_RunMigrations(t *testing.T) {
	migrator := migration.NewMongoMigrator("mongodb://localhost:27017", "/tmp/migrations")
	err := migrator.RunMigrations(nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMongoMigrator_Rollback(t *testing.T) {
	migrator := migration.NewMongoMigrator("mongodb://localhost:27017", "/tmp/migrations")
	err := migrator.Rollback(nil, 1)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestMongoMigrator_GetAppliedMigrations(t *testing.T) {
	migrator := migration.NewMongoMigrator("mongodb://localhost:27017", "/tmp/migrations")
	migrations, err := migrator.GetAppliedMigrations(nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if migrations == nil {
		t.Error("expected migrations to be returned")
	}
}

func TestNewPostgresMigrator(t *testing.T) {
	migrator := migration.NewPostgresMigrator("postgres://localhost/test", "/tmp/migrations")
	if migrator == nil {
		t.Error("expected migrator to be created")
	}
}

func TestNewMongoMigrator(t *testing.T) {
	migrator := migration.NewMongoMigrator("mongodb://localhost:27017", "/tmp/migrations")
	if migrator == nil {
		t.Error("expected migrator to be created")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
