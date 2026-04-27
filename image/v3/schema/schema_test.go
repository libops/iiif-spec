package schema_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/libops/iiif-spec/image/v3/schema"
)

func TestValidateInfoFixtures(t *testing.T) {
	fixturesDir := filepath.Join("..", "..", "..", "upstream", "image-validator", "tests", "json")

	valid := []string{
		"info-3.0.json",
		"info-3.0-logo.json",
		"info-3.0-service.json",
		"info-3.0-service-label.json",
	}
	for _, name := range valid {
		t.Run(name, func(t *testing.T) {
			doc, err := os.ReadFile(filepath.Join(fixturesDir, name))
			if err != nil {
				t.Fatal(err)
			}
			if err := schema.ValidateInfoBytes(doc); err != nil {
				t.Fatalf("valid fixture rejected: %v", err)
			}
		})
	}

	invalid := []string{
		"info-2.0.json",
		"info-3.0-service-badlabel.json",
	}
	for _, name := range invalid {
		t.Run(name, func(t *testing.T) {
			doc, err := os.ReadFile(filepath.Join(fixturesDir, name))
			if err != nil {
				t.Fatal(err)
			}
			if err := schema.ValidateInfoBytes(doc); err == nil {
				t.Fatal("invalid fixture accepted")
			}
		})
	}
}
