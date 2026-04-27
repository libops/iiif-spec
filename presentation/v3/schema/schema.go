package schema

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaURI = "embedded://iiif-spec/presentation/v3/schema/iiif_3_0.json"

//go:embed schemas/iiif_3_0.json
var schemaFiles embed.FS

var (
	compileOnce sync.Once
	compiled    *jsonschema.Schema
	compileErr  error
)

// ValidateBytes validates any top-level Presentation API document supported by
// the vendored upstream aggregate schema.
func ValidateBytes(doc []byte) error {
	s, err := compiledSchema()
	if err != nil {
		return err
	}
	var value any
	if err := json.Unmarshal(doc, &value); err != nil {
		return fmt.Errorf("decode presentation document: %w", err)
	}
	if err := s.Validate(value); err != nil {
		return fmt.Errorf("validate presentation document: %w", err)
	}
	return nil
}

// ValidateManifestBytes validates a Presentation API Manifest document against
// the vendored upstream validator schema.
func ValidateManifestBytes(doc []byte) error {
	return validateBytes("Manifest", "manifest", doc)
}

// ValidateCanvasBytes validates a Presentation API Canvas document against the
// vendored upstream validator schema.
func ValidateCanvasBytes(doc []byte) error {
	return validateBytes("Canvas", "canvas", doc)
}

// ValidateAnnotationPageBytes validates a Presentation API AnnotationPage
// document against the vendored upstream validator schema.
func ValidateAnnotationPageBytes(doc []byte) error {
	return validateBytes("AnnotationPage", "annotation page", doc)
}

// ValidateAnnotationBytes validates a Presentation API Annotation document
// against the vendored upstream validator schema.
func ValidateAnnotationBytes(doc []byte) error {
	return validateBytes("Annotation", "annotation", doc)
}

func validateBytes(schemaName, label string, doc []byte) error {
	if err := requireType(doc, schemaName); err != nil {
		return err
	}
	s, err := compiledSchema()
	if err != nil {
		return err
	}
	var value any
	if err := json.Unmarshal(doc, &value); err != nil {
		return fmt.Errorf("decode %s: %w", label, err)
	}
	if err := s.Validate(value); err != nil {
		return fmt.Errorf("validate %s: %w", label, err)
	}
	return nil
}

func compiledSchema() (*jsonschema.Schema, error) {
	compileOnce.Do(func() {
		c := jsonschema.NewCompiler()
		c.DefaultDraft(jsonschema.Draft7)

		data, err := schemaFiles.ReadFile("schemas/iiif_3_0.json")
		if err != nil {
			compileErr = fmt.Errorf("read presentation schema: %w", err)
			return
		}
		var schemaDoc any
		if err := json.Unmarshal(data, &schemaDoc); err != nil {
			compileErr = fmt.Errorf("decode presentation schema: %w", err)
			return
		}
		if err := c.AddResource(schemaURI, schemaDoc); err != nil {
			compileErr = fmt.Errorf("add presentation schema: %w", err)
			return
		}
		compiled, compileErr = c.Compile(schemaURI)
		if compileErr != nil {
			compileErr = fmt.Errorf("compile presentation schema: %w", compileErr)
		}
	})
	if compileErr != nil {
		return nil, compileErr
	}
	return compiled, nil
}

func requireType(doc []byte, want string) error {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(doc, &probe); err != nil {
		return fmt.Errorf("decode presentation document type: %w", err)
	}
	if probe.Type != want {
		return fmt.Errorf("presentation document type is %q, want %q", probe.Type, want)
	}
	return nil
}
