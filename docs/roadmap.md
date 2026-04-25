# iiif-spec roadmap

This is the execution plan for turning `iiif-spec` from a scaffold into a
useful shared artifact repository for IIIF implementations.

The order here is deliberate. Each milestone should leave the repo more useful
to downstream consumers, even if later milestones have not landed yet.

## Milestone 1: Image v3 hardening

Goal: make the Image API surface reliable enough that downstream projects can
consume it as a real dependency for validation and client generation.

Deliverables:

- tighten `derived/image/v3/schema/info.schema.json`
  - use vendored `IIIF/api` artifacts:
    - `source/image/3/context.json`
    - `source/image/3/level0.json`
    - `source/image/3/level1.json`
    - `source/image/3/level2.json`
  - validate against vendored `image-validator` examples
- make `openapi/image/v3/openapi.yaml` implementation-neutral
  - remove triplet-specific assumptions
  - keep the URL grammar documented clearly where pure schema is insufficient
- add tests for:
  - derived schema acceptance/rejection against known examples
  - OpenAPI reference resolution

Success criteria:

- downstream code generators can consume the Image OpenAPI artifact
- emitted `info.json` can be validated against the derived schema

## Milestone 2: Presentation v3 OpenAPI

Goal: publish the first broadly useful machine-readable Presentation contract.

Deliverables:

- normalize Presentation inputs from:
  - `IIIF/api/source/presentation/3/context.json`
  - `presentation-validator` schemas
  - validator fixtures
- produce a first-pass Presentation artifact set:
  - derived schema layer where needed
  - `openapi/presentation/v3/openapi.yaml`
- model the most important document surfaces first:
  - Manifest
  - Canvas
  - AnnotationPage
  - Annotation
  - TextualBody / SpecificResource

Success criteria:

- downstream projects can generate Presentation wire/document types from the
  published artifacts
- the contract is not tied to triplet’s internal route layout

## Milestone 3: Auth v2

Goal: make the Auth API machine-readable enough for interoperable client work.

Deliverables:

- derive schema inputs from:
  - `IIIF/api/source/auth/2/context.json`
  - Auth 2 prose spec
- publish:
  - `derived/auth/v2/schema/...`
  - `openapi/auth/v2/openapi.yaml`
- capture service/resource shapes:
  - probe service
  - access service
  - access token response
  - logout service

Success criteria:

- a client implementer can generate Auth wire types and service clients from
  `iiif-spec`

## Milestone 4: Search v2

Goal: add the other common interoperability surface needed by viewers and OCR
flows.

Deliverables:

- derive Search schema layer from:
  - `IIIF/api/source/search/2/context.json`
  - prose spec
- publish:
  - `derived/search/v2/schema/...`
  - `openapi/search/v2/openapi.yaml`

Success criteria:

- search clients/results can be generated from published artifacts

## Milestone 5: Extension support

Goal: stop treating important IIIF extensions as prose-only in this repo.

Priority extensions:

- Text Granularity
- navPlace
- georef

Deliverables:

- extension-specific derived artifacts under `derived/extensions/`
- extension-aware OpenAPI/schema references where they affect core APIs

Success criteria:

- downstream projects can model common extensions without inventing their own
  one-off document types

## Milestone 6: Downstream adoption

Goal: prove `iiif-spec` is actually useful.

First downstream target:

- `triplet`

Deliverables:

- move triplet off repo-local copied spec artifacts where practical
- have triplet consume `iiif-spec` outputs as upstream machine-readable inputs
- add at least one more downstream example:
  - TypeScript
  - Python
  - Java/Kotlin

Success criteria:

- at least two independent implementations consume published `iiif-spec`
  artifacts without triplet-specific assumptions leaking through

## Repo-level tasks

These cut across all milestones:

- add CI for vendoring, schema checks, and OpenAPI validation
- add a top-level lock/manifest file summarizing vendored upstream SHAs
- document artifact versioning and release policy
- publish guidance for downstream code generation patterns

## Naming review point

Keep the repo name `iiif-spec` through Milestones 1-3.

Revisit naming only after:

- Presentation and Auth artifacts exist
- the repo clearly contains more than just schemas
- there is interest from the broader IIIF community in adoption or transfer

If that happens, evaluate whether the broader ecosystem would prefer:

- `iiif-spec`
- `iiif-schema`
- eventual migration to an IIIF community-owned repo such as `IIIF/schema`
