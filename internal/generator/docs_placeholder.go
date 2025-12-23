package generator

import (
	"path/filepath"
	"strings"

	"github.com/eduardo/blueprint/internal/domain"
)

func generateDocsPlaceholder(projectPath string, config *domain.Config, fs domain.FileSystemPort) error {
	const docsTemplate = `package docs

import "github.com/swaggo/swag"

var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "localhost:8080",
	BasePath:         "/",
	Schemes:          []string{},
	Title:            "{{.ProjectName}} API",
	Description:      "This is a sample server.",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

var docTemplate = ` + "`" + `{
    "swagger": "2.0",
    "info": {
        "description": "{{.ProjectName}} API",
        "title": "{{.ProjectName}} API",
        "contact": {},
        "version": "1.0"
    },
    "host": "localhost:8080",
    "basePath": "/",
    "paths": {}
}` + "`" + `

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
`
    // We can't use the template engine here easily because it's in the same package, 
    // but we can just use string replacement or a simple write since it's a placeholder.
    // Actually, let's just write a static file for now, or use the template if we can access it.
    // The Generate function has access to the template engine.
    
    // Let's just write the file directly to avoid complex template logic for now.
    // We need to replace {{.ProjectName}} manually.
    content := strings.ReplaceAll(docsTemplate, "{{.ProjectName}}", config.ProjectName)
    
	if err := fs.MkdirAll(filepath.Join(projectPath, "docs")); err != nil {
		return err
	}
	return fs.WriteFile(filepath.Join(projectPath, "docs", "docs.go"), []byte(content))
}
