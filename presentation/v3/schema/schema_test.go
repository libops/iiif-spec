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

func validateByDocumentType(doc []byte) error {
	var probe struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(doc, &probe); err != nil {
		return err
	}
	switch probe.Type {
	case "Manifest":
		return schema.ValidateManifestBytes(doc)
	case "Canvas":
		return schema.ValidateCanvasBytes(doc)
	case "AnnotationPage":
		return schema.ValidateAnnotationPageBytes(doc)
	case "Annotation":
		return schema.ValidateAnnotationBytes(doc)
	default:
		return schema.ValidateBytes(doc)
	}
}
