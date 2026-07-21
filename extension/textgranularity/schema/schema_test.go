package schema_test

import (
	"os"
	"path/filepath"
	"testing"

	textgranularity "github.com/libops/iiif-spec/extension/textgranularity/schema"
)

func TestValidateAnnotationPageBytes(t *testing.T) {
	tests := []struct {
		name    string
		doc     string
		wantErr bool
	}{
		{
			name: "known level with compact target",
			doc: `{
				"@context":[{"example":"https://example.org/vocab#"},"http://iiif.io/api/extension/text-granularity/context.json","http://iiif.io/api/presentation/3/context.json"],
				"id":"https://example.org/pages/1",
				"type":"AnnotationPage",
				"items":[{
					"id":"https://example.org/annotations/1",
					"type":"Annotation",
					"motivation":"supplementing",
					"body":{"type":"TextualBody","value":"hello"},
					"target":"https://example.org/canvas/1#xywh=1,2,3,4",
					"textGranularity":"line"
				}]
			}`,
		},
		{
			name: "registry extension level",
			doc: `{
				"@context":["http://iiif.io/api/extension/text-granularity/context.json","http://iiif.io/api/presentation/3/context.json"],
				"id":"https://example.org/pages/1",
				"type":"AnnotationPage",
				"items":[{"id":"https://example.org/annotations/1","type":"Annotation","textGranularity":"token"}]
			}`,
		},
		{
			name: "empty string is allowed by the normative type constraint",
			doc: `{
				"@context":["http://iiif.io/api/extension/text-granularity/context.json","http://iiif.io/api/presentation/3/context.json"],
				"id":"https://example.org/pages/1",
				"type":"AnnotationPage",
				"items":[{"id":"https://example.org/annotations/1","type":"Annotation","textGranularity":""}]
			}`,
		},
		{
			name: "page without extension",
			doc: `{
				"@context":"http://iiif.io/api/presentation/3/context.json",
				"id":"https://example.org/pages/1",
				"type":"AnnotationPage",
				"items":[]
			}`,
		},
		{
			name: "non string value",
			doc: `{
				"@context":["http://iiif.io/api/extension/text-granularity/context.json","http://iiif.io/api/presentation/3/context.json"],
				"id":"https://example.org/pages/1",
				"type":"AnnotationPage",
				"items":[{"id":"https://example.org/annotations/1","type":"Annotation","textGranularity":["line"]}]
			}`,
			wantErr: true,
		},
		{
			name: "missing extension context",
			doc: `{
				"@context":"http://iiif.io/api/presentation/3/context.json",
				"id":"https://example.org/pages/1",
				"type":"AnnotationPage",
				"items":[{"id":"https://example.org/annotations/1","type":"Annotation","textGranularity":"line"}]
			}`,
			wantErr: true,
		},
		{
			name: "Presentation context is not final",
			doc: `{
				"@context":["http://iiif.io/api/presentation/3/context.json","http://iiif.io/api/extension/text-granularity/context.json"],
				"id":"https://example.org/pages/1",
				"type":"AnnotationPage",
				"items":[{"id":"https://example.org/annotations/1","type":"Annotation","textGranularity":"line"}]
			}`,
			wantErr: true,
		},
		{
			name: "invalid entry after Text Granularity context",
			doc: `{
				"@context":["http://iiif.io/api/extension/text-granularity/context.json",42,"http://iiif.io/api/presentation/3/context.json"],
				"id":"https://example.org/pages/1",
				"type":"AnnotationPage",
				"items":[{"id":"https://example.org/annotations/1","type":"Annotation","textGranularity":"line"}]
			}`,
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := textgranularity.ValidateAnnotationPageBytes([]byte(test.doc))
			if test.wantErr && err == nil {
				t.Fatal("invalid Text Granularity page accepted")
			}
			if !test.wantErr && err != nil {
				t.Fatalf("valid Text Granularity page rejected: %v", err)
			}
		})
	}
}

func TestValidateStandaloneAnnotationAndVocabulary(t *testing.T) {
	doc := []byte(`{
		"@context":["http://iiif.io/api/extension/text-granularity/context.json","http://iiif.io/api/presentation/3/context.json"],
		"id":"https://example.org/annotations/1",
		"type":"Annotation",
		"target":"https://example.org/canvas/1",
		"textGranularity":"word"
	}`)
	if err := textgranularity.ValidateAnnotationBytes(doc); err != nil {
		t.Fatalf("valid standalone Annotation rejected: %v", err)
	}
	if !textgranularity.IsKnownLevel("word") {
		t.Fatal("word was not recognized as a built-in level")
	}
	if textgranularity.IsKnownLevel(" word ") {
		t.Fatal("whitespace-padded value reported as a built-in level")
	}
	if textgranularity.IsKnownLevel("token") {
		t.Fatal("registry-defined value reported as built-in")
	}
}

func TestPublishedSchemaMatchesDerivedArtifact(t *testing.T) {
	published, err := os.ReadFile("annotation.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	derived, err := os.ReadFile(filepath.Join("..", "..", "..", "derived", "extensions", "text-granularity", "schema", "annotation.schema.json"))
	if err != nil {
		t.Fatal(err)
	}
	if string(published) != string(derived) {
		t.Fatal("published Text Granularity schema differs from derived artifact")
	}
}
