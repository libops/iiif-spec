#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GENERATOR="go run github.com/atombender/go-jsonschema@v0.23.0 --only-models --tags json"

run_gen() {
  local package_name="$1"
  local schema_path="$2"
  local output_path="$3"

  mkdir -p "$(dirname "$output_path")"
  (cd "$ROOT_DIR" && PATH="/usr/local/go/bin:$PATH" ${GENERATOR} -p "$package_name" -o "$output_path" "$schema_path")
}

run_gen gen \
  derived/image/v3/schema/info.schema.json \
  image/v3/gen/info.gen.go

run_gen manifest \
  upstream/presentation-validator/schema/v4/Manifest.json \
  presentation/v3/gen/manifest/manifest.gen.go

run_gen canvas \
  upstream/presentation-validator/schema/v4/Canvas.json \
  presentation/v3/gen/canvas/canvas.gen.go
perl -0pi -e 's/ContainerJsonAccompanyingCanvas/AccompanyingCanvasJson/g' \
  "$ROOT_DIR/presentation/v3/gen/canvas/canvas.gen.go"

run_gen annotation \
  upstream/presentation-validator/schema/v4/AnnotationPage.json \
  presentation/v3/gen/annotation/annotation.gen.go

run_gen textualbody \
  upstream/presentation-validator/schema/v4/TextualBody.json \
  presentation/v3/gen/textualbody/textualbody.gen.go

run_gen specificresource \
  upstream/presentation-validator/schema/v4/SpecificResource.json \
  presentation/v3/gen/specificresource/specificresource.gen.go
