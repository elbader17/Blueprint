package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/charmbracelet/huh"
)

func runInteractiveMode() (string, error) {
	var action string

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Blueprint Generator").
				Options(
					huh.NewOption("Create new blueprint", "create"),
					huh.NewOption("Select existing blueprint", "select"),
				).
				Value(&action),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	switch action {
	case "create":
		return createNewBlueprint()
	case "select":
		return selectExistingBlueprint()
	default:
		return "", fmt.Errorf("invalid option")
	}
}

func createNewBlueprint() (string, error) {
	var (
		projectName    string
		dbType         string
		enableAuth     bool
		userCollection string
	)

	// 1. General Info & Auth
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Title("Project Name").
				Value(&projectName),
			huh.NewSelect[string]().
				Title("Database Type").
				Options(
					huh.NewOption("Firestore", "firestore"),
					huh.NewOption("PostgreSQL", "postgresql"),
					huh.NewOption("MongoDB", "mongodb"),
				).
				Value(&dbType),
		),
		huh.NewGroup(
			huh.NewConfirm().
				Title("Enable Authentication?").
				Value(&enableAuth),
			huh.NewInput().
				Title("User Collection Name").
				Value(&userCollection).
				Validate(func(s string) error {
					if enableAuth && strings.TrimSpace(s) == "" {
						return fmt.Errorf("user collection name is required")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	// 2. Models
	type Field struct {
		Name string
		Type string
	}
	type Relation struct {
		Name        string
		Type        string
		TargetModel string
	}
	type Model struct {
		Name      string
		Protected bool
		Fields    []Field
		Relations []Relation
	}
	var models []Model

	for {
		var addModel bool
		huh.NewForm(
			huh.NewGroup(
				huh.NewConfirm().
					Title("Add a model?").
					Value(&addModel),
			),
		).Run()

		if !addModel {
			break
		}

		var (
			modelName      string
			modelProtected bool
		)

		huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("Model Name").
					Value(&modelName),
				huh.NewConfirm().
					Title("Protected Route?").
					Value(&modelProtected),
			),
		).Run()

		var fields []Field
		for {
			var addField bool
			huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Add a field to %s?", modelName)).
						Value(&addField),
				),
			).Run()

			if !addField {
				break
			}

			var (
				fieldName string
				fieldType string
			)

			huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Field Name").
						Value(&fieldName),
					huh.NewSelect[string]().
						Title("Field Type").
						Options(
							huh.NewOption("String", "string"),
							huh.NewOption("Integer", "integer"),
							huh.NewOption("Float", "float"),
							huh.NewOption("Boolean", "boolean"),
							huh.NewOption("DateTime", "datetime"),
						).
						Value(&fieldType),
				),
			).Run()

			fields = append(fields, Field{Name: fieldName, Type: fieldType})
		}

		var relations []Relation
		for {
			var addRelation bool
			huh.NewForm(
				huh.NewGroup(
					huh.NewConfirm().
						Title(fmt.Sprintf("Add a relation to %s?", modelName)).
						Value(&addRelation),
				),
			).Run()

			if !addRelation {
				break
			}

			var (
				relName   string
				relType   string
				targetMod string
			)

			huh.NewForm(
				huh.NewGroup(
					huh.NewInput().
						Title("Relation Name (e.g. items)").
						Value(&relName),
					huh.NewSelect[string]().
						Title("Relation Type").
						Options(
							huh.NewOption("Has Many", "hasMany"),
							huh.NewOption("Belongs To", "belongsTo"),
							huh.NewOption("Has One", "hasOne"),
						).
						Value(&relType),
					huh.NewInput().
						Title("Target Model (e.g. order_items)").
						Value(&targetMod),
				),
			).Run()

			relations = append(relations, Relation{Name: relName, Type: relType, TargetModel: targetMod})
		}

		models = append(models, Model{
			Name:      modelName,
			Protected: modelProtected,
			Fields:    fields,
			Relations: relations,
		})
	}

	// Construct JSON
	type Config struct {
		ProjectName string `json:"project_name"`
		Database    struct {
			Type      string `json:"type"`
			ProjectID string `json:"project_id"`
		} `json:"database"`
		Auth *struct {
			Enabled        bool   `json:"enabled"`
			UserCollection string `json:"user_collection"`
		} `json:"auth,omitempty"`
		Models []map[string]interface{} `json:"models"`
	}

	config := Config{
		ProjectName: projectName,
	}
	config.Database.Type = dbType
	config.Database.ProjectID = "your-project-id"

	if enableAuth {
		config.Auth = &struct {
			Enabled        bool   `json:"enabled"`
			UserCollection string `json:"user_collection"`
		}{
			Enabled:        true,
			UserCollection: userCollection,
		}
	}

	for _, m := range models {
		fieldsMap := make(map[string]string)
		for _, f := range m.Fields {
			fieldsMap[f.Name] = f.Type
		}
		relationsMap := make(map[string]string)
		for _, r := range m.Relations {
			relationsMap[r.Name] = fmt.Sprintf("%s:%s", r.Type, r.TargetModel)
		}
		modelMap := map[string]interface{}{
			"name":      m.Name,
			"protected": m.Protected,
			"fields":    fieldsMap,
			"relations": relationsMap,
		}
		config.Models = append(config.Models, modelMap)
	}

	jsonBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return "", err
	}

	content := fmt.Sprintf("# %s Blueprint\n\n```json\n%s\n```\n", projectName, string(jsonBytes))
	filename := "blueprint.md"

	if err := os.WriteFile(filename, []byte(content), 0644); err != nil {
		return "", fmt.Errorf("failed to create blueprint file: %w", err)
	}

	fmt.Printf("Created %s\n", filename)
	return filename, nil
}

func selectExistingBlueprint() (string, error) {
	files, err := filepath.Glob("*.md")
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("no .md files found in current directory")
	}

	var options []huh.Option[string]
	for _, f := range files {
		options = append(options, huh.NewOption(f, f))
	}

	var selectedFile string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select a blueprint").
				Options(options...).
				Value(&selectedFile),
		),
	)

	if err := form.Run(); err != nil {
		return "", err
	}

	return selectedFile, nil
}

// Helper to keep bufio for now if needed, but we are using huh
func _unused() {
	_ = bufio.NewReader(os.Stdin)
	_ = strconv.Itoa(1)
}
