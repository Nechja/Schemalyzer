package schema

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	
	"github.com/nechja/schemalyzer/pkg/models"
	"gopkg.in/yaml.v3"
)

type Loader struct{}

func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) LoadFromFile(path string) (*models.Schema, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()
	
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return l.LoadFromJSON(file)
	case ".yaml", ".yml":
		return l.LoadFromYAML(file)
	default:
		return nil, fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func (l *Loader) LoadFromJSON(reader io.Reader) (*models.Schema, error) {
	decoder := json.NewDecoder(reader)
	schema := &models.Schema{}
	
	if err := decoder.Decode(schema); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}
	
	return schema, nil
}

func (l *Loader) LoadFromYAML(reader io.Reader) (*models.Schema, error) {
	decoder := yaml.NewDecoder(reader)
	schema := &models.Schema{}
	
	if err := decoder.Decode(schema); err != nil {
		return nil, fmt.Errorf("failed to decode YAML: %w", err)
	}
	
	return schema, nil
}

func (l *Loader) SaveToFile(schema *models.Schema, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	ext := filepath.Ext(path)
	switch ext {
	case ".json":
		return l.SaveToJSON(schema, file)
	case ".yaml", ".yml":
		return l.SaveToYAML(schema, file)
	default:
		return fmt.Errorf("unsupported file extension: %s", ext)
	}
}

func (l *Loader) SaveToJSON(schema *models.Schema, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	
	if err := encoder.Encode(schema); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}
	
	return nil
}

func (l *Loader) SaveToYAML(schema *models.Schema, writer io.Writer) error {
	encoder := yaml.NewEncoder(writer)
	encoder.SetIndent(2) // Use 2 spaces for indentation

	if err := encoder.Encode(schema); err != nil {
		return fmt.Errorf("failed to encode YAML: %w", err)
	}

	return nil
}