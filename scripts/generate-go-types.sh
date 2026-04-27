#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
GENERATOR="go run github.com/atombender/go-jsonschema@v0.23.0 --only-models --tags json"
PRESENTATION_SCHEMA_DIR="$(mktemp -d)"

cleanup() {
  rm -rf "$PRESENTATION_SCHEMA_DIR"
}
trap cleanup EXIT

install_schema() {
  local source_path="$1"
  local output_path="$2"

  mkdir -p "$(dirname "$output_path")"
  cp "$source_path" "$output_path"
}

run_gen() {
  local package_name="$1"
  local schema_path="$2"
  local output_path="$3"
  local err_path

  mkdir -p "$(dirname "$output_path")"
  err_path="$(mktemp)"
  if ! (cd "$ROOT_DIR" && PATH="/usr/local/go/bin:$PATH" ${GENERATOR} -p "$package_name" -o "$output_path" "$schema_path" 2>"$err_path"); then
    cat "$err_path" >&2
    rm -f "$err_path"
    return 1
  fi
  awk '
    /^go-jsonschema: Warning: Field "(id|type)" maps to a field by the same name declared in the same struct; it will be declared as (Id|Type)_2$/ { next }
    /^go-jsonschema: Warning: Object type with no properties has required fields; skipping validation code for them since we don'\''t know their types$/ { next }
    { print > "/dev/stderr" }
  ' "$err_path"
  rm -f "$err_path"
}

prepare_presentation_schemas() {
  cp -R "$ROOT_DIR/upstream/presentation-validator/schema/v4/." "$PRESENTATION_SCHEMA_DIR/"

  # The upstream v4 schemas use title-cased refs for a few lowercase files.
  # Normalize refs in a temporary copy so local file resolution is stable on
  # case-sensitive filesystems while leaving vendored artifacts untouched.
  perl -0pi -e 's/"Agent\.json"/"agent.json"/g; s/"Audiance\.json"/"audiance.json"/g; s/"Metadata\.json"/"metadata.json"/g; s/"Service\.json"/"service.json"/g' \
    "$PRESENTATION_SCHEMA_DIR"/*.json
}

prepare_presentation_schemas

run_gen gen \
  derived/image/v3/schema/info.schema.json \
  image/v3/gen/info.gen.go

run_gen manifest \
  "$PRESENTATION_SCHEMA_DIR/Manifest.json" \
  presentation/v3/gen/manifest/manifest.gen.go

run_gen canvas \
  "$PRESENTATION_SCHEMA_DIR/Canvas.json" \
  presentation/v3/gen/canvas/canvas.gen.go
perl -0pi -e 's/ContainerJsonAccompanyingCanvas/AccompanyingCanvasJson/g' \
  "$ROOT_DIR/presentation/v3/gen/canvas/canvas.gen.go"

run_gen annotation \
  "$PRESENTATION_SCHEMA_DIR/AnnotationPage.json" \
  presentation/v3/gen/annotation/annotation.gen.go

run_gen textualbody \
  "$PRESENTATION_SCHEMA_DIR/TextualBody.json" \
  presentation/v3/gen/textualbody/textualbody.gen.go

run_gen specificresource \
  "$PRESENTATION_SCHEMA_DIR/SpecificResource.json" \
  presentation/v3/gen/specificresource/specificresource.gen.go

install_schema \
  "$ROOT_DIR/upstream/presentation-validator/schema/iiif_3_0.json" \
  "$ROOT_DIR/presentation/v3/schema/schemas/iiif_3_0.json"
