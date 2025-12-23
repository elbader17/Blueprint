package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eduardo/blueprint/internal/infrastructure"
	"github.com/eduardo/blueprint/internal/parser"
)

func main() {
	fs := infrastructure.NewOSFileSystem()
	p := parser.NewMarkdownParser(fs)

	config, err := p.Parse("test_relations.md")
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	fmt.Printf("Project Name: %s\n", config.ProjectName)
}
