.PHONY: generate generate-upstream generate-go-types fmt

generate: generate-upstream generate-go-types fmt

generate-upstream:
	go run ./tools/iiifgen

generate-go-types:
	bash ./scripts/generate-go-types.sh

fmt:
	gofmt -w .
