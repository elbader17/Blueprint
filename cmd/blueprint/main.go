package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eduardo/blueprint/internal/generator"
	"github.com/eduardo/blueprint/internal/parser"
)

func main() {
	f, _ := os.Create("debug_log.txt")
	defer f.Close()
	log.SetOutput(f)

	// 1. Parse the blueprint
	filename := "blueprint.md"
	if len(os.Args) > 1 {
		filename = os.Args[1]
	}
	log.Printf("Step 1/3: Parsing blueprint: %s", filename)

	config, err := parser.ParseBlueprint(filename)
	if err != nil {
		log.Fatalf("Failed to parse blueprint: %v", err)
	}
	log.Printf("Parsed config: %+v", config)

	// 2. Generate the project
	// We'll generate it in the current directory for now, creating a folder with the project name
	outputDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	log.Printf("Step 2/3: Generating project '%s'...", config.ProjectName)
	log.Printf("Output directory: %s", outputDir)
	if err := generator.Generate(config, outputDir); err != nil {
		log.Fatalf("Failed to generate project: %v", err)
	}

	log.Printf("Successfully generated project in %s/%s", outputDir, config.ProjectName)

	// 3. Run setup.sh
	projectPath := filepath.Join(outputDir, config.ProjectName)
	log.Printf("Step 3/3: Running setup script in %s...", projectPath)
	log.Printf("This will install dependencies, generate docs, and start the server.")

	cmd := exec.Command("./setup.sh")
	cmd.Dir = projectPath
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Printf("Warning: Failed to run setup.sh: %v", err)
	}
}
