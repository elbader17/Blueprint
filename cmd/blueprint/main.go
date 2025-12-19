package main

import (
	"context"
	"log"
	"os"

	"github.com/eduardo/blueprint/internal/application"
	"github.com/eduardo/blueprint/internal/generator"
	"github.com/eduardo/blueprint/internal/infrastructure"
	"github.com/eduardo/blueprint/internal/parser"
)

func main() {
	f, _ := os.Create("debug_log.txt")
	defer f.Close()
	log.SetOutput(f)

	// 1. Initialize Adapters
	fs := infrastructure.NewOSFileSystem()
	templateEngine := infrastructure.NewGoTemplateEngine()
	markdownParser := parser.NewMarkdownParser(fs)

	// 2. Initialize Application Service
	// We pass the Generate function from the generator package as a dependency
	blueprintService := application.NewBlueprintService(fs, templateEngine, markdownParser, generator.Generate)

	// 3. Parse and Generate
	filename := "blueprint.md"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}

	outputDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	log.Printf("Generating project from blueprint: %s", filename)
	if err := blueprintService.Generate(context.Background(), filename, outputDir); err != nil {
		log.Fatalf("Failed to generate project: %v", err)
	}

	// 4. Run setup.sh (Optional/Manual step usually, but keeping it for compatibility)
	// We need to know the project name, which is in the config. 
	// Since we don't have the config here easily (it's inside the service), 
	// we might need to change the service to return it or just let the user run setup.sh.
	// For now, I'll assume the user wants to run it manually or I'll just skip it if I can't easily get the project name.
	// Actually, I'll just skip it to keep the main clean and follow the new architecture.
	log.Printf("Successfully generated project in %s", outputDir)
}
