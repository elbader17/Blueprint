package tests

import (
	"testing"

	"github.com/eduardo/blueprint/internal/domain"
	"github.com/eduardo/blueprint/internal/parser"
)

type mockFileSystem struct {
	files map[string][]byte
}

func (m *mockFileSystem) ReadFile(path string) ([]byte, error) {
	if content, ok := m.files[path]; ok {
		return content, nil
	}
	return nil, &mockError{msg: "file not found"}
}

func (m *mockFileSystem) MkdirAll(path string) error {
	return nil
}

func (m *mockFileSystem) WriteFile(path string, data []byte) error {
	m.files[path] = data
	return nil
}

func (m *mockFileSystem) CopyFile(src, dst string) error {
	return nil
}

func (m *mockFileSystem) Chmod(path string, mode uint32) error {
	return nil
}

func (m *mockFileSystem) RemoveAll(path string) error {
	return nil
}

type mockError struct {
	msg string
}

func (e *mockError) Error() string {
	return e.msg
}

func TestParser_ValidJSON(t *testing.T) {
	fs := &mockFileSystem{
		files: map[string][]byte{
			"blueprint.md": []byte("# My Blueprint\n\n```json\n{\n  \"project_name\": \"TestAPI\",\n  \"database\": {\n    \"type\": \"firestore\",\n    \"project_id\": \"test-project\"\n  },\n  \"models\": [\n    {\n      \"name\": \"users\",\n      \"protected\": false,\n      \"fields\": {\n        \"email\": \"string\",\n        \"name\": \"string\"\n      }\n    }\n  ]\n}\n```\n"),
		},
	}

	p := parser.NewMarkdownParser(fs)
	cfg, err := p.Parse("blueprint.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.ProjectName != "TestAPI" {
		t.Errorf("expected project name 'TestAPI', got %s", cfg.ProjectName)
	}
	if cfg.Database.Type != "firestore" {
		t.Errorf("expected database type 'firestore', got %s", cfg.Database.Type)
	}
	if len(cfg.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(cfg.Models))
	}
	if cfg.Models[0].Name != "users" {
		t.Errorf("expected model name 'users', got %s", cfg.Models[0].Name)
	}
}

func TestParser_DefaultFirestore(t *testing.T) {
	fs := &mockFileSystem{
		files: map[string][]byte{
			"blueprint.md": []byte("# Test\n\n```json\n{\n  \"project_name\": \"TestAPI\",\n  \"models\": []\n}\n```\n"),
		},
	}

	p := parser.NewMarkdownParser(fs)
	cfg, err := p.Parse("blueprint.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Database.Type != "firestore" {
		t.Errorf("expected default database type 'firestore', got %s", cfg.Database.Type)
	}
}

func TestParser_NoJSONBlock(t *testing.T) {
	fs := &mockFileSystem{
		files: map[string][]byte{
			"blueprint.md": []byte("# Test\n\nNo JSON here\n"),
		},
	}

	p := parser.NewMarkdownParser(fs)
	_, err := p.Parse("blueprint.md")
	if err == nil {
		t.Fatal("expected error for missing JSON block")
	}
}

func TestParser_InvalidJSON(t *testing.T) {
	fs := &mockFileSystem{
		files: map[string][]byte{
			"blueprint.md": []byte("# Test\n\n```json\n{\n  invalid json\n}\n```\n"),
		},
	}

	p := parser.NewMarkdownParser(fs)
	_, err := p.Parse("blueprint.md")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParser_FileNotFound(t *testing.T) {
	fs := &mockFileSystem{
		files: map[string][]byte{},
	}

	p := parser.NewMarkdownParser(fs)
	_, err := p.Parse("nonexistent.md")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestParser_WithAuth(t *testing.T) {
	fs := &mockFileSystem{
		files: map[string][]byte{
			"blueprint.md": []byte("# Test\n\n```json\n{\n  \"project_name\": \"AuthAPI\",\n  \"database\": {\n    \"type\": \"postgresql\",\n    \"url\": \"postgres://localhost/test\"\n  },\n  \"auth\": {\n    \"enabled\": true,\n    \"provider\": \"jwt\",\n    \"user_collection\": \"users\"\n  },\n  \"models\": []\n}\n```\n"),
		},
	}

	p := parser.NewMarkdownParser(fs)
	cfg, err := p.Parse("blueprint.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Auth == nil || !cfg.Auth.Enabled {
		t.Fatal("expected auth to be enabled")
	}
	if cfg.Auth.Provider != "jwt" {
		t.Errorf("expected auth provider 'jwt', got %s", cfg.Auth.Provider)
	}
}

func TestParser_WithPayments(t *testing.T) {
	fs := &mockFileSystem{
		files: map[string][]byte{
			"blueprint.md": []byte("# Test\n\n```json\n{\n  \"project_name\": \"ShopAPI\",\n  \"database\": {\n    \"type\": \"mongodb\",\n    \"url\": \"mongodb://localhost:27017\"\n  },\n  \"payments\": {\n    \"enabled\": true,\n    \"provider\": \"stripe\",\n    \"transactions_collection\": \"transactions\"\n  },\n  \"models\": []\n}\n```\n"),
		},
	}

	p := parser.NewMarkdownParser(fs)
	cfg, err := p.Parse("blueprint.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.Payments == nil || !cfg.Payments.Enabled {
		t.Fatal("expected payments to be enabled")
	}
	if cfg.Payments.Provider != "stripe" {
		t.Errorf("expected payment provider 'stripe', got %s", cfg.Payments.Provider)
	}
}

func TestParser_ModelRelations(t *testing.T) {
	fs := &mockFileSystem{
		files: map[string][]byte{
			"blueprint.md": []byte("# Test\n\n```json\n{\n  \"project_name\": \"SocialAPI\",\n  \"database\": {\n    \"type\": \"firestore\",\n    \"project_id\": \"social-app\"\n  },\n  \"models\": [\n    {\n      \"name\": \"posts\",\n      \"protected\": true,\n      \"fields\": {\n        \"title\": \"string\",\n        \"content\": \"text\"\n      },\n      \"relations\": {\n        \"author\": \"belongsTo:users\",\n        \"comments\": \"hasMany:comments\"\n      }\n    }\n  ]\n}\n```\n"),
		},
	}

	p := parser.NewMarkdownParser(fs)
	cfg, err := p.Parse("blueprint.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(cfg.Models) != 1 {
		t.Fatalf("expected 1 model, got %d", len(cfg.Models))
	}
	model := cfg.Models[0]
	if !model.Protected {
		t.Error("expected model to be protected")
	}
	if len(model.Relations) != 2 {
		t.Errorf("expected 2 relations, got %d", len(model.Relations))
	}
	if model.Relations["author"] != "belongsTo:users" {
		t.Errorf("expected author relation 'belongsTo:users', got %s", model.Relations["author"])
	}
}

func TestParser_AllFieldTypes(t *testing.T) {
	fs := &mockFileSystem{
		files: map[string][]byte{
			"blueprint.md": []byte("# Test\n\n```json\n{\n  \"project_name\": \"TypesAPI\",\n  \"database\": {\n    \"type\": \"firestore\",\n    \"project_id\": \"types-test\"\n  },\n  \"models\": [\n    {\n      \"name\": \"products\",\n      \"protected\": false,\n      \"fields\": {\n        \"name\": \"string\",\n        \"description\": \"text\",\n        \"price\": \"float\",\n        \"quantity\": \"integer\",\n        \"in_stock\": \"boolean\",\n        \"created_at\": \"datetime\"\n      }\n    }\n  ]\n}\n```\n"),
		},
	}

	p := parser.NewMarkdownParser(fs)
	cfg, err := p.Parse("blueprint.md")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	model := cfg.Models[0]
	expectedFields := map[string]string{
		"name":        "string",
		"description": "text",
		"price":       "float",
		"quantity":    "integer",
		"in_stock":    "boolean",
		"created_at":  "datetime",
	}
	for field, expectedType := range expectedFields {
		if actualType, ok := model.Fields[field]; !ok || actualType != expectedType {
			t.Errorf("expected field %s to be type %s, got %s", field, expectedType, actualType)
		}
	}
}

func TestDomainModels(t *testing.T) {
	cfg := &domain.Config{
		ProjectName: "TestProject",
		Database: domain.Database{
			Type:      "firestore",
			ProjectID: "test-id",
		},
		Auth: &domain.Auth{
			Enabled:        true,
			Provider:       "jwt",
			UserCollection: "users",
		},
		Payments: &domain.Payments{
			Enabled:          true,
			Provider:         "mercadopago",
			TransactionsColl: "transactions",
		},
		Pagination: &domain.Pagination{
			DefaultLimit: 20,
		},
		Models: []domain.Model{
			{
				Name:      "posts",
				Protected: true,
				Fields: map[string]string{
					"title": "string",
				},
				Relations: map[string]string{
					"author": "belongsTo:users",
				},
			},
		},
	}

	if cfg.ProjectName != "TestProject" {
		t.Errorf("expected TestProject, got %s", cfg.ProjectName)
	}
	if cfg.Auth.Provider != "jwt" {
		t.Errorf("expected jwt, got %s", cfg.Auth.Provider)
	}
	if cfg.Payments.Provider != "mercadopago" {
		t.Errorf("expected mercadopago, got %s", cfg.Payments.Provider)
	}
	if cfg.Pagination.DefaultLimit != 20 {
		t.Errorf("expected 20, got %d", cfg.Pagination.DefaultLimit)
	}
	if len(cfg.Models) != 1 {
		t.Errorf("expected 1 model, got %d", len(cfg.Models))
	}
}
