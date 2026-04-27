.PHONY: generate generate-upstream generate-go-types fmt lint test test-artifacts test-race

generate: generate-upstream generate-go-types fmt

generate-upstream:
	go run ./tools/iiifgen

generate-go-types:
	bash ./scripts/generate-go-types.sh

fmt:
	gofmt -w .

lint:
	go vet ./...

test: test-artifacts

test-artifacts:
	go test ./...

test-race:
	go test -race ./...
