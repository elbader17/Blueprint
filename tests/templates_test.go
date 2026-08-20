package tests

import (
	"testing"

	"github.com/eduardo/blueprint/internal/infrastructure"
)

func TestGoTemplateEngine_Render(t *testing.T) {
	engine := infrastructure.NewGoTemplateEngine()

	tests := []struct {
		name      string
		tmpl      string
		data      interface{}
		expected  string
		shouldErr bool
	}{
		{
			name:     "simple title case",
			tmpl:     `{{.Name | title}}`,
			data:     struct{ Name string }{Name: "hello"},
			expected: "Hello",
		},
		{
			name:     "lowercase",
			tmpl:     `{{.Name | lower}}`,
			data:     struct{ Name string }{Name: "HELLO"},
			expected: "hello",
		},
		{
			name:     "pascal case single",
			tmpl:     `{{.Name | pascal}}`,
			data:     struct{ Name string }{Name: "hello"},
			expected: "Hello",
		},
		{
			name:     "pascal case with underscore",
			tmpl:     `{{.Name | pascal}}`,
			data:     struct{ Name string }{Name: "hello_world"},
			expected: "HelloWorld",
		},
		{
			name:     "add function",
			tmpl:     `{{add 2 3}}`,
			data:     nil,
			expected: "5",
		},
		{
			name:     "hasPrefix true",
			tmpl:     `{{hasPrefix "hello world" "hello"}}`,
			data:     nil,
			expected: "true",
		},
		{
			name:     "hasPrefix false",
			tmpl:     `{{hasPrefix "goodbye" "hello"}}`,
			data:     nil,
			expected: "false",
		},
		{
			name:     "conditional with eq",
			tmpl:     `{{if eq .Type "firestore"}}FIRESTORE{{else}}OTHER{{end}}`,
			data:     struct{ Type string }{Type: "firestore"},
			expected: "FIRESTORE",
		},
		{
			name:     "conditional with eq else",
			tmpl:     `{{if eq .Type "postgresql"}}PG{{else}}OTHER{{end}}`,
			data:     struct{ Type string }{Type: "mongodb"},
			expected: "OTHER",
		},
		{
			name:     "range over slice",
			tmpl:     `{{range .Items}}{{.}} {{end}}`,
			data:     struct{ Items []string }{Items: []string{"a", "b", "c"}},
			expected: "a b c ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render("test", tt.tmpl, tt.data)
			if tt.shouldErr && err == nil {
				t.Error("expected error but got none")
			}
			if !tt.shouldErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(result) != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, string(result))
			}
		})
	}
}

func TestGoTemplateEngine_InvalidTemplate(t *testing.T) {
	engine := infrastructure.NewGoTemplateEngine()
	_, err := engine.Render("test", `{{.Invalid}`, nil)
	if err == nil {
		t.Error("expected error for invalid template")
	}
}

func TestGoTemplateEngine_FuncMap(t *testing.T) {
	engine := infrastructure.NewGoTemplateEngine()

	result, err := engine.Render("test", `{{"hello" | title}}`, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(result) != "Hello" {
		t.Errorf("expected Hello, got %s", string(result))
	}

	result2, err := engine.Render("test", `{{"" | title}}`, nil)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(result2) != "" {
		t.Errorf("expected empty string, got %s", string(result2))
	}
}
