package domain

import "context"

// FileSystemPort defines the interface for file and directory operations
type FileSystemPort interface {
	MkdirAll(path string) error
	WriteFile(path string, data []byte) error
	ReadFile(path string) ([]byte, error)
	CopyFile(src, dst string) error
	Chmod(path string, mode uint32) error
	RemoveAll(path string) error
}

// TemplatePort defines the interface for rendering templates
type TemplatePort interface {
	Render(name, tmpl string, data interface{}) ([]byte, error)
}

// ParserPort defines the interface for parsing the blueprint
type ParserPort interface {
	Parse(filename string) (*Config, error)
}

// BlueprintServicePort defines the interface for the core generation logic
type BlueprintServicePort interface {
	Generate(ctx context.Context, config *Config, outputDir string) error
}
