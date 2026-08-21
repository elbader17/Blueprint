package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/eduardo/blueprint/internal/application"
	"github.com/eduardo/blueprint/internal/generator"
	"github.com/eduardo/blueprint/internal/infrastructure"
	"github.com/eduardo/blueprint/internal/parser"
)

func main() {
	f, _ := os.Create("debug_log.txt")
	defer f.Close()
	log.SetOutput(f)

	if len(os.Args) > 1 && os.Args[1] == "migrate" {
		runMigrateCommand()
		return
	}

	runGenerateCommand()
}

func runMigrateCommand() {
	execPath, err := os.Executable()
	if err != nil {
		log.Fatalf("Failed to get executable path: %v", err)
	}

	migratePath := filepath.Join(filepath.Dir(execPath), "blueprint", "migrate")
	if _, err := os.Stat(migratePath); os.IsNotExist(err) {
		migratePath, err = lookMigrateBinary()
		if err != nil {
			log.Fatalf("Migrate command not found. Please build with: go build -o blueprint ./cmd/blueprint && go build -o blueprint/migrate ./cmd/blueprint/migrate")
		}
	}

	cmd := exec.Command(migratePath, os.Args[2:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Run(); err != nil {
		log.Fatalf("Migrate command failed: %v", err)
	}
}

func lookMigrateBinary() (string, error) {
	paths := []string{
		"./blueprint/migrate",
		"../blueprint/migrate",
		"cmd/blueprint/migrate",
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return filepath.Abs(p)
		}
	}

	return "", fmt.Errorf("migrate binary not found")
}

func runGenerateCommand() {
	fs := infrastructure.NewOSFileSystem()
	templateEngine := infrastructure.NewGoTemplateEngine()
	markdownParser := parser.NewMarkdownParser(fs)

	blueprintService := application.NewBlueprintService(fs, templateEngine, markdownParser, generator.Generate)

	var filename string
	var outputDir string
	var err error

	if len(os.Args) > 1 {
		filename = os.Args[1]
		if len(os.Args) > 2 {
			outputDir = os.Args[2]
		} else {
			outputDir, err = os.Getwd()
			if err != nil {
				log.Fatalf("Failed to get current directory: %v", err)
			}
		}
	} else {
		filename, err = runInteractiveMode()
		if err != nil {
			log.Fatalf("Interactive mode failed: %v", err)
		}
		outputDir, err = os.Getwd()
		if err != nil {
			log.Fatalf("Failed to get current directory: %v", err)
		}
	}

	log.Printf("Generating project from blueprint: %s", filename)
	if err := blueprintService.Generate(context.Background(), filename, outputDir); err != nil {
		log.Fatalf("Failed to generate project: %v", err)
	}

	log.Printf("Successfully generated project in %s", outputDir)
}
