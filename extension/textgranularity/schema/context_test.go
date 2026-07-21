package schema

import (
	"testing"

	presentationschema "github.com/libops/iiif-spec/presentation/v3/schema"
)

func TestValidateContextsChecksEveryEntry(t *testing.T) {
	err := validateContexts([]any{
		Context,
		float64(42),
		presentationschema.Context,
	})
	if err == nil {
		t.Fatal("invalid entry after the Text Granularity context accepted")
	}
}
