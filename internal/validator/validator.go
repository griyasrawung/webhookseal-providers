package validator

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/xeipuuv/gojsonschema"
	"gopkg.in/yaml.v3"
)

func ValidateProviderSpec(schemaPath, specPath string) error {
	schemaAbs, err := filepath.Abs(schemaPath)
	if err != nil {
		return fmt.Errorf("resolve schema path: %w", err)
	}
	data, err := os.ReadFile(specPath)
	if err != nil {
		return fmt.Errorf("read spec: %w", err)
	}
	var generic map[string]any
	if err := yaml.Unmarshal(data, &generic); err != nil {
		return fmt.Errorf("parse yaml: %w", err)
	}
	jsonBytes, err := json.Marshal(generic)
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	result, err := gojsonschema.Validate(
		gojsonschema.NewReferenceLoader("file:///"+filepath.ToSlash(schemaAbs)),
		gojsonschema.NewBytesLoader(jsonBytes),
	)
	if err != nil {
		return fmt.Errorf("validate schema: %w", err)
	}
	if !result.Valid() {
		return fmt.Errorf("schema validation failed: %v", result.Errors())
	}
	return nil
}