# Go Example

Downstream Go projects should treat this repository as an artifact source, not
as a runtime dependency.

Recommended flow:

1. vendor or fetch the desired JSON Schema / OpenAPI document from `iiif-spec`
2. generate wire types or clients in your own repo
3. keep handwritten application logic separate from generated code
