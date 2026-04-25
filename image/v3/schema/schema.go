package schema

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/santhosh-tekuri/jsonschema/v6"
)

const infoSchemaURI = "embedded://iiif-spec/image/v3/info.schema.json"

var (
	//go:embed info.schema.json
	infoSchemaBytes []byte

	infoOnce     sync.Once
	infoCompiled *jsonschema.Schema
	infoErr      error
)

// ValidateInfoBytes validates an Image API info.json payload against the
// iiif-spec maintained derived schema.
func ValidateInfoBytes(doc []byte) error {
	s, err := compiledInfoSchema()
	if err != nil {
		return err
	}
	var value any
	if err := json.Unmarshal(doc, &value); err != nil {
		return fmt.Errorf("decode info.json: %w", err)
	}
	if err := s.Validate(value); err != nil {
		return fmt.Errorf("validate info.json: %w", err)
	}
	return nil
}

func compiledInfoSchema() (*jsonschema.Schema, error) {
	infoOnce.Do(func() {
		c := jsonschema.NewCompiler()
		c.DefaultDraft(jsonschema.Draft2020)

		var schemaDoc any
		if err := json.Unmarshal(infoSchemaBytes, &schemaDoc); err != nil {
			infoErr = fmt.Errorf("decode info schema: %w", err)
			return
		}
		if err := c.AddResource(infoSchemaURI, schemaDoc); err != nil {
			infoErr = fmt.Errorf("add info schema: %w", err)
			return
		}
		infoCompiled, infoErr = c.Compile(infoSchemaURI)
		if infoErr != nil {
			infoErr = fmt.Errorf("compile info schema: %w", infoErr)
		}
	})
	return infoCompiled, infoErr
}
