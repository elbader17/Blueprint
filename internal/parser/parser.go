package parser

import (
	"encoding/json"
	"fmt"
	"regexp"

	"github.com/eduardo/blueprint/internal/domain"
)

// MarkdownParser implements domain.ParserPort
type MarkdownParser struct {
	fs domain.FileSystemPort
}

func NewMarkdownParser(fs domain.FileSystemPort) *MarkdownParser {
	return &MarkdownParser{fs: fs}
}

// Parse reads the markdown file and extracts the JSON configuration
func (p *MarkdownParser) Parse(filename string) (*domain.Config, error) {
	content, err := p.fs.ReadFile(filename)
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

	var config domain.Config
	if err := json.Unmarshal(jsonContent, &config); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Default to Firestore if no database is specified
	if config.Database.Type == "" {
		config.Database.Type = "firestore"
		if config.Database.ProjectID == "" {
			config.Database.ProjectID = config.FirestoreProjectID
		}
	}

	return &config, nil
}
