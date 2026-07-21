package schema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/libops/iiif-spec/presentation/v3/schema"
)

func TestValidatePresentationFixtures(t *testing.T) {
	fixturesDir := filepath.Join("..", "..", "..", "upstream", "presentation-validator", "fixtures", "3")
	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" || strings.HasPrefix(entry.Name(), "broken_") || presentationFixtureIsKnownIssue(entry.Name()) {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			doc, err := os.ReadFile(filepath.Join(fixturesDir, name))
			if err != nil {
				t.Fatal(err)
			}
			err = validateByDocumentType(doc)
			if err != nil {
				t.Fatalf("valid fixture rejected: %v", err)
			}
		})
	}
}

func presentationFixtureIsKnownIssue(name string) bool {
	switch name {
	case "collection_of_canvases.json",
		"non_cc_license.json",
		"old_format_label.json",
		"rights_lang_issues.json":
		return true
	default:
		return false
	}
}

func TestRejectsBrokenPresentationFixtures(t *testing.T) {
	fixturesDir := filepath.Join("..", "..", "..", "upstream", "presentation-validator", "fixtures", "3")
	entries, err := os.ReadDir(fixturesDir)
	if err != nil {
		t.Fatal(err)
	}

	checked := 0
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" || !strings.HasPrefix(entry.Name(), "broken_") {
			continue
		}
		name := entry.Name()
		t.Run(name, func(t *testing.T) {
			doc, err := os.ReadFile(filepath.Join(fixturesDir, name))
			if err != nil {
				t.Fatal(err)
			}
			err = validateByDocumentType(doc)
			if err == nil {
				t.Fatal("broken fixture accepted")
			}
		})
		checked++
	}
	if checked == 0 {
		t.Fatal("no broken presentation fixtures found")
	}
}

func TestValidateCanvasBytesUsesCanvasClassSchema(t *testing.T) {
	valid := []byte(`{
		"@context":"http://iiif.io/api/presentation/3/context.json",
		"id":"https://example.org/canvas/1",
		"type":"Canvas",
		"height":1000,
		"width":800,
		"items":[]
	}`)
	if err := schema.ValidateCanvasBytes(valid); err != nil {
		t.Fatalf("valid standalone Canvas rejected: %v", err)
	}

	invalid := []byte(`{
		"@context":"http://iiif.io/api/presentation/3/context.json",
		"id":"https://example.org/canvas/1",
		"type":"Canvas",
		"height":0,
		"width":800
	}`)
	if err := schema.ValidateCanvasBytes(invalid); err == nil {
		t.Fatal("Canvas with an invalid dimension accepted")
	}

	missingContext := []byte(`{
		"id":"https://example.org/canvas/1",
		"type":"Canvas",
		"height":1000,
		"width":800,
		"items":[]
	}`)
	if err := schema.ValidateCanvasBytes(missingContext); err == nil {
		t.Fatal("standalone Canvas without the required @context accepted")
	}
}

func TestNamedTopLevelValidators(t *testing.T) {
	collection := []byte(`{
		"@context":"http://iiif.io/api/presentation/3/context.json",
		"id":"https://example.org/collections/1",
		"type":"Collection",
		"label":{"en":["Example"]},
		"items":[]
	}`)
	if err := schema.ValidateCollectionBytes(collection); err != nil {
		t.Fatalf("valid standalone Collection rejected: %v", err)
	}

	annotationCollection := []byte(`{
		"@context":"http://iiif.io/api/presentation/3/context.json",
		"id":"https://example.org/annotation-collections/1",
		"type":"AnnotationCollection"
	}`)
	if err := schema.ValidateAnnotationCollectionBytes(annotationCollection); err != nil {
		t.Fatalf("valid standalone AnnotationCollection rejected: %v", err)
	}
}

func TestValidateExtensibleResources(t *testing.T) {
	extensiblePage := []byte(`{
		"@context":["https://example.org/extension/context.json","http://iiif.io/api/presentation/3/context.json"],
		"id":"https://example.org/pages/1",
		"type":"AnnotationPage",
		"items":[],
		"example:custom":7
	}`)
	if err := schema.ValidateAnnotationPageBytes(extensiblePage); err == nil {
		t.Fatal("strict validator accepted an AnnotationPage extension property")
	}
	if err := schema.ValidateExtensibleAnnotationPageBytes(extensiblePage); err != nil {
		t.Fatalf("extension-aware validator rejected a legal extension property: %v", err)
	}

	extensibleManifest := []byte(`{
		"@context":["https://example.org/extension/context.json","http://iiif.io/api/presentation/3/context.json"],
		"id":"https://example.org/manifests/1",
		"type":"Manifest",
		"label":{"en":["Example"]},
		"items":[{
			"id":"https://example.org/canvases/1",
			"type":"Canvas",
			"height":1000,
			"width":800,
			"items":[]
		}],
		"example:custom":true
	}`)
	if err := schema.ValidateManifestBytes(extensibleManifest); err == nil {
		t.Fatal("strict validator accepted a Manifest extension property")
	}
	if err := schema.ValidateExtensibleManifestBytes(extensibleManifest); err != nil {
		t.Fatalf("extension-aware validator rejected a legal Manifest extension property: %v", err)
	}

	inlineContextCanvas := []byte(`{
		"@context":[{"example":"https://example.org/extension#"},"http://iiif.io/api/presentation/3/context.json"],
		"id":"https://example.org/canvases/1",
		"type":"Canvas",
		"height":1000,
		"width":800,
		"items":[],
		"example:custom":true
	}`)
	if err := schema.ValidateCanvasBytes(inlineContextCanvas); err == nil {
		t.Fatal("strict validator accepted an inline JSON-LD context object")
	}
	if err := schema.ValidateExtensibleCanvasBytes(inlineContextCanvas); err != nil {
		t.Fatalf("extension-aware validator rejected an inline JSON-LD context object: %v", err)
	}

	invalidCore := []byte(`{
		"@context":["https://example.org/extension/context.json","http://iiif.io/api/presentation/3/context.json"],
		"id":"not-an-http-uri",
		"type":"AnnotationPage",
		"items":[],
		"example:custom":7
	}`)
	if err := schema.ValidateExtensibleAnnotationPageBytes(invalidCore); err == nil {
		t.Fatal("extension-aware validator accepted an invalid core id")
	}

	wrongContextOrder := []byte(`{
		"@context":["http://iiif.io/api/presentation/3/context.json","https://example.org/extension/context.json"],
		"id":"https://example.org/pages/1",
		"type":"AnnotationPage",
		"items":[],
		"example:custom":7
	}`)
	if err := schema.ValidateExtensibleAnnotationPageBytes(wrongContextOrder); err == nil {
		t.Fatal("extension-aware validator accepted a non-final Presentation context")
	}

	invalidContextValue := []byte(`{
		"@context":[42,"http://iiif.io/api/presentation/3/context.json"],
		"id":"https://example.org/pages/1",
		"type":"AnnotationPage",
		"items":[]
	}`)
	if err := schema.ValidateExtensibleAnnotationPageBytes(invalidContextValue); err == nil {
		t.Fatal("extension-aware validator accepted a non-string/non-object context entry")
	}

	duplicatePresentationContext := []byte(`{
		"@context":["http://iiif.io/api/presentation/3/context.json","http://iiif.io/api/presentation/3/context.json"],
		"id":"https://example.org/pages/1",
		"type":"AnnotationPage",
		"items":[]
	}`)
	if err := schema.ValidateExtensibleAnnotationPageBytes(duplicatePresentationContext); err == nil {
		t.Fatal("extension-aware validator accepted a repeated Presentation context")
	}

	topLevelGraph := []byte(`{
		"@context":"http://iiif.io/api/presentation/3/context.json",
		"id":"https://example.org/pages/1",
		"type":"AnnotationPage",
		"items":[],
		"@graph":[]
	}`)
	if err := schema.ValidateExtensibleAnnotationPageBytes(topLevelGraph); err == nil {
		t.Fatal("extension-aware validator accepted a top-level @graph")
	}

	embeddedContext := []byte(`{
		"@context":"http://iiif.io/api/presentation/3/context.json",
		"id":"https://example.org/pages/1",
		"type":"AnnotationPage",
		"items":[{
			"@context":"http://iiif.io/api/presentation/3/context.json",
			"id":"https://example.org/annotations/1",
			"type":"Annotation"
		}]
	}`)
	if err := schema.ValidateExtensibleAnnotationPageBytes(embeddedContext); err == nil {
		t.Fatal("extension-aware validator accepted @context on an embedded resource")
	}
}

func validateByDocumentType(doc []byte) error {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(doc, &probe); err != nil {
		return err
	}
	switch probe.Type {
	case "Collection":
		return schema.ValidateCollectionBytes(doc)
	case "Manifest":
		return schema.ValidateManifestBytes(doc)
	case "Canvas":
		return schema.ValidateCanvasBytes(doc)
	case "AnnotationCollection":
		return schema.ValidateAnnotationCollectionBytes(doc)
	case "AnnotationPage":
		return schema.ValidateAnnotationPageBytes(doc)
	case "Annotation":
		return schema.ValidateAnnotationBytes(doc)
	default:
		return schema.ValidateBytes(doc)
	}
}
