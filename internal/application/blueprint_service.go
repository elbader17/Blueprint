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

	if config.Auth.UserCollection == "" {
		config.Auth.UserCollection = "users"
	}

	if !s.hasModel(config, config.Auth.UserCollection) {
		config.Models = append(config.Models, domain.Model{
			Name:      config.Auth.UserCollection,
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
		})
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
