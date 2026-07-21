package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	presentationschema "github.com/libops/iiif-spec/presentation/v3/schema"
	"github.com/santhosh-tekuri/jsonschema/v6"
)

const (
	// Context is the canonical JSON-LD context for the Text Granularity
	// Extension.
	Context = "http://iiif.io/api/extension/text-granularity/context.json"

	schemaURI = "embedded://iiif-spec/extension/text-granularity/annotation.schema.json"
)

// Level is one of the text granularity values defined by the IIIF context.
// The extension also recommends values from the IIIF Extension Registry, so
// validation accepts any string and IsKnownLevel is available to applications
// that intentionally constrain their own profile to the built-in vocabulary.
type Level string

const (
	LevelPage      Level = "page"
	LevelBlock     Level = "block"
	LevelParagraph Level = "paragraph"
	LevelLine      Level = "line"
	LevelWord      Level = "word"
	LevelGlyph     Level = "glyph"
)

var knownLevels = map[Level]struct{}{
	LevelPage: {}, LevelBlock: {}, LevelParagraph: {},
	LevelLine: {}, LevelWord: {}, LevelGlyph: {},
}

var (
	//go:embed annotation.schema.json
	annotationSchemaBytes []byte

	compileOnce sync.Once
	compiled    *jsonschema.Schema
	compileErr  error
)

// IsKnownLevel reports whether value is one of the six levels defined by the
// IIIF Text Granularity context.
func IsKnownLevel(value string) bool {
	_, ok := knownLevels[Level(value)]
	return ok
}

// ValidateAnnotationBytes validates a standalone Presentation 3 Annotation
// and, when textGranularity is present, its extension value and JSON-LD
// context. An Annotation without the extension remains a valid input.
func ValidateAnnotationBytes(doc []byte) error {
	if err := presentationschema.ValidateExtensibleAnnotationBytes(doc); err != nil {
		return err
	}
	var annotation map[string]any
	if err := json.Unmarshal(doc, &annotation); err != nil {
		return fmt.Errorf("decode text granularity annotation: %w", err)
	}
	usesExtension, err := validateAnnotation(annotation)
	if err != nil {
		return err
	}
	if usesExtension {
		if err := validateContexts(annotation["@context"]); err != nil {
			return fmt.Errorf("validate text granularity annotation context: %w", err)
		}
	}
	return nil
}

// ValidateAnnotationPageBytes validates a Presentation 3 AnnotationPage and
// each embedded Text Granularity Annotation. Extension terms inherit the
// page's top-level JSON-LD context, as required for a dereferenceable page.
func ValidateAnnotationPageBytes(doc []byte) error {
	if err := presentationschema.ValidateExtensibleAnnotationPageBytes(doc); err != nil {
		return err
	}
	var page map[string]any
	if err := json.Unmarshal(doc, &page); err != nil {
		return fmt.Errorf("decode text granularity annotation page: %w", err)
	}
	items, _ := page["items"].([]any)
	usesExtension := false
	for index, value := range items {
		annotation, ok := value.(map[string]any)
		if !ok {
			return fmt.Errorf("validate text granularity annotation %d: expected an object", index)
		}
		uses, err := validateAnnotation(annotation)
		if err != nil {
			return fmt.Errorf("validate text granularity annotation %d: %w", index, err)
		}
		usesExtension = usesExtension || uses
	}
	if usesExtension {
		if err := validateContexts(page["@context"]); err != nil {
			return fmt.Errorf("validate text granularity annotation page context: %w", err)
		}
	}
	return nil
}

func validateAnnotation(annotation map[string]any) (bool, error) {
	if _, ok := annotation["textGranularity"]; !ok {
		return false, nil
	}
	s, err := compiledSchema()
	if err != nil {
		return true, err
	}
	if err := s.Validate(annotation); err != nil {
		return true, fmt.Errorf("validate textGranularity: %w", err)
	}
	return true, nil
}

func validateContexts(raw any) error {
	switch value := raw.(type) {
	case string:
		return fmt.Errorf("Text Granularity context is required before the Presentation 3 context")
	case []any:
		if len(value) == 0 {
			return fmt.Errorf("@context must not be empty")
		}
		last, ok := value[len(value)-1].(string)
		if !ok || last != presentationschema.Context {
			return fmt.Errorf("Presentation 3 context must be the final @context entry")
		}
		found := false
		for _, entry := range value[:len(value)-1] {
			switch context := entry.(type) {
			case string:
				if context == Context {
					found = true
				}
			case map[string]any:
				// JSON-LD permits inline context definitions before the
				// Presentation context. They do not replace this extension's
				// published context URI.
			default:
				return fmt.Errorf("@context entries must be URIs or objects")
			}
		}
		if found {
			return nil
		}
		return fmt.Errorf("Text Granularity context is required before the Presentation 3 context")
	default:
		return fmt.Errorf("@context is required")
	}
}

func compiledSchema() (*jsonschema.Schema, error) {
	compileOnce.Do(func() {
		compiler := jsonschema.NewCompiler()
		compiler.DefaultDraft(jsonschema.Draft2020)
		var schemaDocument any
		if err := json.Unmarshal(annotationSchemaBytes, &schemaDocument); err != nil {
			compileErr = fmt.Errorf("decode text granularity schema: %w", err)
			return
		}
		if err := compiler.AddResource(schemaURI, schemaDocument); err != nil {
			compileErr = fmt.Errorf("add text granularity schema: %w", err)
			return
		}
		compiled, compileErr = compiler.Compile(schemaURI)
		if compileErr != nil {
			compileErr = fmt.Errorf("compile text granularity schema: %w", compileErr)
		}
	})
	return compiled, compileErr
}
