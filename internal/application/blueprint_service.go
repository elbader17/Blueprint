package application

import (
	"context"

	"github.com/eduardo/blueprint/internal/domain"
)

// BlueprintService implements domain.BlueprintServicePort
type BlueprintService struct {
	fs       domain.FileSystemPort
	template domain.TemplatePort
	parser   domain.ParserPort
	// We will add the generator logic here or call a generator adapter
	generateFunc func(config *domain.Config, outputDir string, fs domain.FileSystemPort, template domain.TemplatePort) error
}

func NewBlueprintService(fs domain.FileSystemPort, template domain.TemplatePort, parser domain.ParserPort, generateFunc func(*domain.Config, string, domain.FileSystemPort, domain.TemplatePort) error) *BlueprintService {
	return &BlueprintService{
		fs:           fs,
		template:     template,
		parser:       parser,
		generateFunc: generateFunc,
	}
}

func (s *BlueprintService) Generate(ctx context.Context, blueprintPath, outputDir string) error {
	config, err := s.parser.Parse(blueprintPath)
	if err != nil {
		return err
	}

	s.enrichConfig(config)

	return s.generateFunc(config, outputDir, s.fs, s.template)
}

func (s *BlueprintService) enrichConfig(config *domain.Config) {
	s.enrichAuth(config)
	s.enrichPayments(config)
}

func (s *BlueprintService) enrichAuth(config *domain.Config) {
	if config.Auth == nil || !config.Auth.Enabled {
		return
	}

	if config.Auth.Provider == "" {
		config.Auth.Provider = "firebase"
	}

	if config.Auth.UserCollection == "" {
		config.Auth.UserCollection = "users"
	}

	if !s.hasModel(config, config.Auth.UserCollection) {
		config.Models = append(config.Models, domain.Model{
			Name:      config.Auth.UserCollection,
			Protected: true,
			Fields:    map[string]string{},
			Relations: map[string]string{},
		})
	}

	// Ensure required auth fields are present
	for i, m := range config.Models {
		if m.Name == config.Auth.UserCollection {
			if m.Fields == nil {
				config.Models[i].Fields = make(map[string]string)
			}
			if _, ok := m.Fields["email"]; !ok {
				config.Models[i].Fields["email"] = "string"
			}
			if _, ok := m.Fields["name"]; !ok {
				config.Models[i].Fields["name"] = "string"
			}
			if _, ok := m.Fields["picture"]; !ok {
				config.Models[i].Fields["picture"] = "string"
			}
			if _, ok := m.Fields["role_id"]; !ok {
				config.Models[i].Fields["role_id"] = "string"
			}
			if _, ok := m.Fields["created_at"]; !ok {
				config.Models[i].Fields["created_at"] = "datetime"
			}
			if _, ok := m.Fields["updated_at"]; !ok {
				config.Models[i].Fields["updated_at"] = "datetime"
			}

			if config.Auth.Provider == "firebase" {
				if _, ok := m.Fields["uid"]; !ok {
					config.Models[i].Fields["uid"] = "string"
				}
			} else {
				if _, ok := m.Fields["password"]; !ok {
					config.Models[i].Fields["password"] = "string"
				}
			}
			break
		}
	}
}

func (s *BlueprintService) enrichPayments(config *domain.Config) {
	if config.Payments == nil || !config.Payments.Enabled {
		return
	}

	if config.Payments.TransactionsColl == "" {
		config.Payments.TransactionsColl = "transactions"
	}

	if !s.hasModel(config, config.Payments.TransactionsColl) {
		config.Models = append(config.Models, domain.Model{
			Name:      config.Payments.TransactionsColl,
			Protected: true,
			Fields: map[string]string{
				"amount":     "float",
				"status":     "string",
				"provider":   "string",
				"payload":    "text",
				"created_at": "datetime",
			},
			Relations: map[string]string{},
		})
	}
}

func (s *BlueprintService) hasModel(config *domain.Config, name string) bool {
	for _, model := range config.Models {
		if model.Name == name {
			return true
		}
	}
	return false
}
