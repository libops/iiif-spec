package schema

import (
	"embed"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const schemaURI = "embedded://iiif-spec/presentation/v3/schema/iiif_3_0.json"

const extensibleSchemaURI = "embedded://iiif-spec/presentation/v3/schema/iiif_3_0_extensible.json"

// Context is the canonical IIIF Presentation API 3 JSON-LD context URI.
const Context = "http://iiif.io/api/presentation/3/context.json"

//go:embed schemas/iiif_3_0.json
var schemaFiles embed.FS

var (
	compileOnce sync.Once
	compiled    map[string]*jsonschema.Schema
	compileErr  error

	extensibleCompileOnce sync.Once
	extensibleCompiled    map[string]*jsonschema.Schema
	extensibleCompileErr  error
)

var schemaFragments = map[string]string{
	"":                     "",
	"Collection":           "#/classes/collection",
	"Manifest":             "#/classes/manifest",
	"Canvas":               "#/classes/canvas",
	"Range":                "#/classes/range",
	"AnnotationCollection": "#/classes/annotationCollection",
	"AnnotationPage":       "#/classes/annotationPage",
	"Annotation":           "#/classes/annotation",
}

// ValidateBytes validates any top-level Presentation API document supported by
// the vendored upstream aggregate schema.
func ValidateBytes(doc []byte) error {
	s, err := compiledSchema("")
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
	if err := validateTopLevelContext(value, false); err != nil {
		return fmt.Errorf("validate presentation document context: %w", err)
	}
	return nil
}

// ValidateCollectionBytes validates a Presentation API Collection document
// against the vendored upstream validator schema.
func ValidateCollectionBytes(doc []byte) error {
	return validateBytes("Collection", "collection", doc)
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

// ValidateRangeBytes validates a Presentation API Range document against the
// vendored upstream validator schema.
func ValidateRangeBytes(doc []byte) error {
	return validateBytes("Range", "range", doc)
}

// ValidateAnnotationCollectionBytes validates a Presentation API
// AnnotationCollection document against the vendored upstream validator
// schema.
func ValidateAnnotationCollectionBytes(doc []byte) error {
	return validateBytes("AnnotationCollection", "annotation collection", doc)
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

// ValidateExtensibleBytes validates any supported top-level Presentation API
// document with the IIIF extension points enabled. Unlike the vendored
// community validator schema, this mode permits extension properties on the
// otherwise closed Manifest and AnnotationPage objects. It still validates
// all core properties and requires the Presentation 3 context to be the final
// top-level @context entry.
func ValidateExtensibleBytes(doc []byte) error {
	return validateExtensibleBytes("", "presentation document", doc)
}

// ValidateExtensibleCollectionBytes validates an extension-aware Collection.
func ValidateExtensibleCollectionBytes(doc []byte) error {
	return validateExtensibleBytes("Collection", "collection", doc)
}

// ValidateExtensibleManifestBytes validates an extension-aware Manifest.
func ValidateExtensibleManifestBytes(doc []byte) error {
	return validateExtensibleBytes("Manifest", "manifest", doc)
}

// ValidateExtensibleCanvasBytes validates an extension-aware standalone
// Canvas.
func ValidateExtensibleCanvasBytes(doc []byte) error {
	return validateExtensibleBytes("Canvas", "canvas", doc)
}

// ValidateExtensibleRangeBytes validates an extension-aware standalone Range.
func ValidateExtensibleRangeBytes(doc []byte) error {
	return validateExtensibleBytes("Range", "range", doc)
}

// ValidateExtensibleAnnotationCollectionBytes validates an extension-aware
// AnnotationCollection.
func ValidateExtensibleAnnotationCollectionBytes(doc []byte) error {
	return validateExtensibleBytes("AnnotationCollection", "annotation collection", doc)
}

// ValidateExtensibleAnnotationPageBytes validates an extension-aware
// AnnotationPage.
func ValidateExtensibleAnnotationPageBytes(doc []byte) error {
	return validateExtensibleBytes("AnnotationPage", "annotation page", doc)
}

// ValidateExtensibleAnnotationBytes validates an extension-aware standalone
// Annotation.
func ValidateExtensibleAnnotationBytes(doc []byte) error {
	return validateExtensibleBytes("Annotation", "annotation", doc)
}

func validateBytes(schemaName, label string, doc []byte) error {
	if err := requireType(doc, schemaName); err != nil {
		return err
	}
	s, err := compiledSchema(schemaName)
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
	if err := validateTopLevelContext(value, false); err != nil {
		return fmt.Errorf("validate %s context: %w", label, err)
	}
	return nil
}

func validateExtensibleBytes(schemaName, label string, doc []byte) error {
	if schemaName != "" {
		if err := requireType(doc, schemaName); err != nil {
			return err
		}
	}
	s, err := compiledExtensibleSchema(schemaName)
	if err != nil {
		return err
	}
	var value any
	if err := json.Unmarshal(doc, &value); err != nil {
		return fmt.Errorf("decode %s: %w", label, err)
	}
	if err := s.Validate(value); err != nil {
		return fmt.Errorf("validate extensible %s: %w", label, err)
	}
	if err := validateTopLevelContext(value, true); err != nil {
		return fmt.Errorf("validate extensible %s context: %w", label, err)
	}
	return nil
}

func compiledSchema(schemaName string) (*jsonschema.Schema, error) {
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
		compiled, compileErr = compileSchemaSet(c, schemaURI)
	})
	if compileErr != nil {
		return nil, compileErr
	}
	s, ok := compiled[schemaName]
	if !ok {
		return nil, fmt.Errorf("presentation schema %q is not supported", schemaName)
	}
	return s, nil
}

func compiledExtensibleSchema(schemaName string) (*jsonschema.Schema, error) {
	extensibleCompileOnce.Do(func() {
		c := jsonschema.NewCompiler()
		c.DefaultDraft(jsonschema.Draft7)

		data, err := schemaFiles.ReadFile("schemas/iiif_3_0.json")
		if err != nil {
			extensibleCompileErr = fmt.Errorf("read presentation schema: %w", err)
			return
		}
		var schemaDoc any
		if err := json.Unmarshal(data, &schemaDoc); err != nil {
			extensibleCompileErr = fmt.Errorf("decode presentation schema: %w", err)
			return
		}
		if err := enableResourceExtensions(schemaDoc); err != nil {
			extensibleCompileErr = err
			return
		}
		if err := c.AddResource(extensibleSchemaURI, schemaDoc); err != nil {
			extensibleCompileErr = fmt.Errorf("add extensible presentation schema: %w", err)
			return
		}
		extensibleCompiled, extensibleCompileErr = compileSchemaSet(c, extensibleSchemaURI)
	})
	if extensibleCompileErr != nil {
		return nil, extensibleCompileErr
	}
	s, ok := extensibleCompiled[schemaName]
	if !ok {
		return nil, fmt.Errorf("presentation schema %q is not supported", schemaName)
	}
	return s, nil
}

func compileSchemaSet(compiler *jsonschema.Compiler, baseURI string) (map[string]*jsonschema.Schema, error) {
	schemas := make(map[string]*jsonschema.Schema, len(schemaFragments))
	for name, fragment := range schemaFragments {
		compiledSchema, err := compiler.Compile(baseURI + fragment)
		if err != nil {
			return nil, fmt.Errorf("compile presentation %s schema: %w", schemaLabel(name), err)
		}
		schemas[name] = compiledSchema
	}
	return schemas, nil
}

func enableResourceExtensions(value any) error {
	document, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("derive extensible presentation schema: root is not an object")
	}
	classes, ok := document["classes"].(map[string]any)
	if !ok {
		return fmt.Errorf("derive extensible presentation schema: classes are missing")
	}
	for _, name := range []string{"manifest", "annotationPage"} {
		class, ok := classes[name].(map[string]any)
		if !ok {
			return fmt.Errorf("derive extensible presentation schema: class %s is missing", name)
		}
		allOf, ok := class["allOf"].([]any)
		if !ok || len(allOf) < 2 {
			return fmt.Errorf("derive extensible presentation schema: class %s shape changed", name)
		}
		resource, ok := allOf[1].(map[string]any)
		if !ok {
			return fmt.Errorf("derive extensible presentation schema: class %s resource shape changed", name)
		}
		resource["additionalProperties"] = true
	}
	if err := enableObjectContextEntries(document); err != nil {
		return err
	}
	return nil
}

func enableObjectContextEntries(value any) error {
	switch node := value.(type) {
	case []any:
		for _, child := range node {
			if err := enableObjectContextEntries(child); err != nil {
				return err
			}
		}
	case map[string]any:
		if properties, ok := node["properties"].(map[string]any); ok {
			if contextSchema, ok := properties["@context"].(map[string]any); ok {
				if err := allowObjectContextItems(contextSchema); err != nil {
					return err
				}
			}
		}
		for _, child := range node {
			if err := enableObjectContextEntries(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func allowObjectContextItems(contextSchema map[string]any) error {
	choices, ok := contextSchema["oneOf"].([]any)
	if !ok {
		return nil
	}
	for _, choice := range choices {
		arraySchema, ok := choice.(map[string]any)
		if !ok || arraySchema["type"] != "array" {
			continue
		}
		itemSchema, ok := arraySchema["items"].(map[string]any)
		if !ok {
			return fmt.Errorf("derive extensible presentation schema: @context array item shape changed")
		}
		arraySchema["items"] = map[string]any{
			"oneOf": []any{itemSchema, map[string]any{"type": "object"}},
		}
		return nil
	}
	return nil
}

func validateTopLevelContext(value any, allowObjects bool) error {
	document, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("presentation document must be an object")
	}
	if _, ok := document["@graph"]; ok {
		return fmt.Errorf("@graph is not permitted at the top level")
	}
	for key, child := range document {
		if key == "@context" {
			// Scoped context definitions can themselves contain @context.
			continue
		}
		if hasEmbeddedContext(child) {
			return fmt.Errorf("@context is only permitted on the top-level resource")
		}
	}
	raw, ok := document["@context"]
	if !ok {
		return fmt.Errorf("@context is required")
	}
	switch context := raw.(type) {
	case string:
		if context != Context {
			return fmt.Errorf("@context must be %q", Context)
		}
	case []any:
		if len(context) == 0 {
			return fmt.Errorf("@context must not be empty")
		}
		presentationContexts := 0
		for _, entry := range context {
			switch entry := entry.(type) {
			case string:
				if entry == Context {
					presentationContexts++
					continue
				}
				parsed, err := url.Parse(entry)
				if err != nil || !parsed.IsAbs() || (parsed.Scheme != "http" && parsed.Scheme != "https") {
					return fmt.Errorf("@context string entries must be HTTP(S) URIs")
				}
			case map[string]any:
				if !allowObjects {
					return fmt.Errorf("@context entries must be HTTP(S) URIs")
				}
			default:
				if allowObjects {
					return fmt.Errorf("@context entries must be HTTP(S) URIs or objects")
				}
				return fmt.Errorf("@context entries must be HTTP(S) URIs")
			}
		}
		if presentationContexts != 1 {
			return fmt.Errorf("Presentation 3 context must occur exactly once")
		}
		last, ok := context[len(context)-1].(string)
		if !ok || last != Context {
			return fmt.Errorf("Presentation 3 context must be the final @context entry")
		}
	default:
		return fmt.Errorf("@context must be a URI or array")
	}
	return nil
}

func hasEmbeddedContext(value any) bool {
	switch value := value.(type) {
	case []any:
		for _, child := range value {
			if hasEmbeddedContext(child) {
				return true
			}
		}
	case map[string]any:
		if _, ok := value["@context"]; ok {
			return true
		}
		for _, child := range value {
			if hasEmbeddedContext(child) {
				return true
			}
		}
	}
	return false
}

func schemaLabel(name string) string {
	if name == "" {
		return "root"
	}
	return name
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
