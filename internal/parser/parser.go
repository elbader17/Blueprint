package parser

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
)

// Config represents the top-level structure of the blueprint JSON
type Config struct {
	ProjectName        string  `json:"project_name"`
	FirestoreProjectID string  `json:"firestore_project_id"`
	Auth               *Auth   `json:"auth,omitempty"`
	Models             []Model `json:"models"`
}

// Auth configures the authentication module
type Auth struct {
	Enabled        bool   `json:"enabled"`
	UserCollection string `json:"user_collection"`
}

// Model represents a data model definition
type Model struct {
	Name      string            `json:"name"`
	Protected bool              `json:"protected"`
	Fields    map[string]string `json:"fields"`
	Relations map[string]string `json:"relations"`
}

// ParseBlueprint reads the markdown file and extracts the JSON configuration
func ParseBlueprint(filename string) (*Config, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	// Regex to find the JSON block between ```json and ```
	re := regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")
	matches := re.FindSubmatch(content)

	if len(matches) < 2 {
		return nil, fmt.Errorf("no JSON block found in %s", filename)
	}

	jsonContent := matches[1]

	var config Config
	if err := json.Unmarshal(jsonContent, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Auto-create user model if auth is enabled and model is missing
	if config.Auth != nil && config.Auth.Enabled {
		userCollection := config.Auth.UserCollection
		if userCollection == "" {
			userCollection = "users"
			config.Auth.UserCollection = userCollection
		}

		hasUserModel := false
		for _, model := range config.Models {
			if model.Name == userCollection {
				hasUserModel = true
				break
			}
		}

		if !hasUserModel {
			userModel := Model{
				Name:      userCollection,
				Protected: true,
				Fields: map[string]string{
					"uid":        "string",
					"email":      "string",
					"name":       "string",
					"picture":    "string",
					"roleId":     "string",
					"settingsId": "string",
					"created_at": "datetime",
					"updated_at": "datetime",
				},
				Relations: map[string]string{},
			}
			config.Models = append(config.Models, userModel)
		}
	}

	return &config, nil
}
