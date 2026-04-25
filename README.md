# iiif-spec

Machine-readable IIIF specification artifacts maintained by LibOps.

This repository exists to fill a practical gap in the IIIF ecosystem:

- upstream IIIF specs are authoritative, but not consistently published as
  complete machine-readable API contracts
- some APIs have JSON Schema inputs upstream
- some APIs only have validator fixtures or prose
- client and server implementers still need stable artifacts for code
  generation, validation, and interoperability testing

`iiif-spec` publishes those artifacts with explicit provenance:

- `upstream/`
  - verbatim or vendored machine-readable artifacts from upstream IIIF
    community repos, pinned by commit SHA
  - currently includes both `IIIF/api` contexts/profile JSON and validator
    repo artifacts
- `derived/`
  - LibOps-maintained schemas where upstream does not publish a complete
    machine-readable contract
- `openapi/`
  - OpenAPI documents derived from upstream and/or derived schema inputs
- `examples/`
  - implementation examples showing how downstream projects can consume the
    published artifacts

## Scope

Initial focus:

- IIIF Image API 3.0
- IIIF Presentation API 3.0
- IIIF Authorization Flow API 2.0
- IIIF Content Search API 2.0
- selected extensions:
  - Text Granularity

Planned follow-ons:

- Content State API
- Change Discovery API

## Source Of Truth Rules

This repo is strict about provenance:

1. Normative truth remains the official IIIF specs.
2. If upstream publishes machine-readable artifacts, we vendor them into
   `upstream/` and keep their source manifests.
3. If upstream does not publish a complete machine-readable artifact, we create
   a derived one in `derived/` and say so explicitly.
4. Every derived artifact should document:
   - what upstream prose/spec it was derived from
   - what validator fixtures or examples it was checked against
   - what assumptions or compromises were added

## Current Artifacts

- `derived/image/v3/schema/info.schema.json`
  - derived JSON Schema for IIIF Image 3.0 `info.json`
- `openapi/image/v3/openapi.yaml`
  - initial OpenAPI description for implemented Image API routes
- `upstream/iiif-api/source/...`
  - vendored JSON-LD contexts and related machine-readable artifacts from the
    canonical `IIIF/api` repository

## Tooling

- `go run ./tools/iiifgen`
  - vendors pinned upstream artifacts into `upstream/`
- `make generate`
  - refreshes upstream vendored artifacts and generated Go packages
- `bash ./scripts/generate-go-types.sh`
  - regenerates public Go wire-type packages from the vendored/derived schema
    inputs

Project planning docs:

- `docs/architecture.md`
  - project architecture, status, and naming rationale
- `docs/roadmap.md`
  - milestone plan and execution order

This initial scaffold does not yet generate the full OpenAPI suite for all IIIF
APIs. It establishes the repo structure and the vendoring pipeline so those
artifacts can move here out of downstream implementation repos such as
`triplet`.

## Downstream Consumption

The intended workflow for downstream projects:

1. consume schemas or OpenAPI documents from `iiif-spec`
2. generate language-specific clients or wire types locally
3. keep business logic separate from generated code

For Go consumers, `iiif-spec` also publishes importable generated/wrapper
packages such as:

- `github.com/libops/iiif-spec/image/v3/gen`
- `github.com/libops/iiif-spec/image/v3/schema`
- `github.com/libops/iiif-spec/presentation/v3/gen/...`

Triplet is the first downstream consumer target for this repository.
