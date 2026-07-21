# iiif-spec architecture & status

`iiif-spec` is a machine-readable artifact repository for IIIF APIs and
extensions.

Its job is not to implement a server or client. Its job is to make the IIIF
specification surface consumable by software:

1. vendor upstream machine-readable artifacts where they exist
2. derive machine-readable artifacts where upstream only provides prose or
   incomplete machine-readable inputs
3. publish stable JSON Schema and OpenAPI outputs with explicit provenance
4. make those artifacts usable by downstream projects for code generation,
   validation, and interoperability testing

This repo exists because the IIIF ecosystem has many excellent servers,
viewers, manifest builders, and validators, but no obvious shared home for the
full machine-readable contract layer across APIs.

## Naming

Current name: **`iiif-spec`**

Why keep `spec` instead of `schema` for now:

- the repo contains more than schemas
  - upstream contexts
  - validator fixtures
  - profile JSON
  - derived OpenAPI
  - provenance manifests
- `schema` would be too narrow if the repository becomes the broader
  machine-readable contract project for IIIF
- `spec` better matches the long-term goal: a software-consumable
  representation of the IIIF specification surface

If the IIIF community later wants a narrower or more official name such as
`IIIF/schema`, that is still compatible with the structure here. The internal
model should remain:

- `upstream/` for canonical machine-readable inputs
- `derived/` for maintained machine-readable outputs not published upstream
- `openapi/` for language-agnostic API contracts

## High-level shape

```text
                       ┌──────────────────────────────┐
                       │          IIIF/api            │
                       │ contexts, profiles, frames   │
                       └──────────────┬───────────────┘
                                      │
                       ┌──────────────▼───────────────┐
                       │  IIIF validator repositories  │
                       │ schemas, fixtures, examples   │
                       └──────────────┬───────────────┘
                                      │
                                      ▼
                       ┌──────────────────────────────┐
                       │     tools/iiifgen            │
                       │ pinned-SHA vendoring         │
                       └──────────────┬───────────────┘
                                      │
                      ┌───────────────┼────────────────┐
                      ▼               ▼                ▼
           ┌─────────────────┐ ┌──────────────┐ ┌──────────────┐
           │    upstream/    │ │   derived/   │ │   openapi/   │
           │ verbatim inputs │ │ local schemas│ │ API contracts│
           └─────────────────┘ └──────────────┘ └──────────────┘
                                      │
                                      ▼
                       ┌──────────────────────────────┐
                       │ downstream generators/SDKs   │
                       │ triplet, TS, Python, Java    │
                       └──────────────────────────────┘
```

## Principles

- **Normative truth stays upstream.**
  The official IIIF specs remain the final authority.
- **Every artifact has provenance.**
  Vendored files get `.source` manifests. Derived files must say what they were
  derived from.
- **Do not blur upstream and derived outputs.**
  A consumer should be able to tell instantly whether a file is canonical,
  vendored, or LibOps-maintained.
- **OpenAPI is an output, not the only source.**
  JSON Schema, JSON-LD contexts, validator fixtures, and prose all contribute.
- **Keep downstream generation local.**
  This repo should publish stable machine-readable contracts; consuming repos
  can generate language-specific code from them.

## Component status

### Done

| Path | What it does | Notes |
|---|---|---|
| `README.md` | Project overview, source-of-truth rules, and current artifact summary. | |
| `go.mod` | Minimal Go module for vendoring tooling. | |
| `Makefile` | `make generate` runs upstream vendoring and Go wire-type regeneration. | |
| `scripts/generate-go-types.sh` | Regenerates public Go wire-type packages from the vendored/derived schema sources. | This is the current Go-consumer bridge for downstream repos such as triplet. |
| `tools/iiifgen/main.go` | Vendors pinned upstream machine-readable artifacts into `upstream/` and writes `.source` manifests. | Sources currently include `IIIF/api`, `IIIF/presentation-validator`, and `IIIF/image-validator`. |
| `upstream/iiif-api/` | Vendored machine-readable files from the canonical `IIIF/api` repo. | Includes Image 3 context/profile JSON, Presentation 3 context, Auth 2 context, Search 2 context, and extension contexts. |
| `upstream/presentation-validator/` | Vendored Presentation validator schema and fixture inputs. | Includes `schema/iiif_3_0.json`, `schema/v4/*.json`, and `fixtures/3/*.json`. |
| `upstream/image-validator/` | Vendored Image validator JSON examples. | Includes `tests/json/*.json`. |
| `image/v3/gen/` | Public Go wire-type package generated from the derived Image schema. | Downstream Go consumers can import this directly. |
| `image/v3/schema/` | Public Go package exposing the derived Image schema validator. | Provides `ValidateInfoBytes`. |
| `presentation/v3/gen/` | Public Go wire-type packages generated from vendored Presentation schema files. | Split per-resource because the upstream aggregate schema is not generator-friendly. |
| `derived/image/v3/schema/info.schema.json` | Derived JSON Schema for Image API 3.0 `info.json`. | Derived from IIIF Image 3 prose, `IIIF/api` context/profile JSON, and `image-validator` examples. |
| `openapi/image/v3/openapi.yaml` | Initial OpenAPI description for IIIF Image API 3.0. | Early derived contract; currently shaped by triplet’s implemented surface. |
| `derived/extensions/text-granularity/` | Derived Text Granularity schema and provenance. | Published with a composed Go validator. |
| `examples/go/README.md` | Documents the intended downstream consumption pattern. | |

### In progress / not done

#### Image API

- tighten `derived/image/v3/schema/info.schema.json` using vendored
  `IIIF/api/source/image/3/context.json` and `level{0,1,2}.json`
- generate a broader, implementation-neutral `openapi/image/v3/openapi.yaml`
  from the derived schema plus route grammar, rather than the current
  triplet-shaped first pass
- add artifact tests that validate:
  - derived schema against `image-validator` JSON examples
  - derived OpenAPI references resolve cleanly

#### Presentation API

- derive `presentation/v3` JSON Schemas or a normalized schema layer from the
  combination of:
  - `IIIF/api/source/presentation/3/context.json`
  - `presentation-validator` schemas
  - validator fixtures
- publish:
  - `openapi/presentation/v3/openapi.yaml`
- model extension-aware annotation surfaces, especially:
  - Text Granularity
  - navPlace

#### Auth API

- derive schemas from:
  - `IIIF/api/source/auth/2/context.json`
  - Auth 2 prose spec
- publish:
  - `derived/auth/v2/schema/...`
  - `openapi/auth/v2/openapi.yaml`

#### Search API

- vendor Search 2 context is already present from `IIIF/api`
- still need:
  - derived schema work
  - `openapi/search/v2/openapi.yaml`

#### Extensions

- extend `derived/extensions/text-granularity/` beyond its Annotation property
  schema when additional normative machine-readable constraints emerge
- add extension directories as needed for:
  - navPlace
  - georef
  - registry-backed extension terms where relevant

#### Tooling

- add a second-generation derivation toolchain:
  - context-aware schema derivation
  - OpenAPI normalization/generation
  - validation checks in CI
- decide whether public generated Go packages remain in this repo long-term or
  become example outputs once language-agnostic artifacts are mature
- add a manifest/lock file summarizing vendored upstream SHAs in one place,
  instead of only per-file `.source` manifests

#### Downstream integration

- move triplet to consume artifacts from `iiif-spec` rather than owning copied
  derived inputs locally
- add downstream example consumers beyond Go
  - TypeScript
  - Python
  - Java or Kotlin

## Current source dependency model

There are three source classes:

1. **Canonical upstream machine-readable inputs**
   - `IIIF/api`
   - contexts, profile JSON, frame JSON
2. **Canonical upstream validation/community artifacts**
   - validator schemas
   - validator fixtures/examples
3. **LibOps-derived artifacts**
   - JSON Schema where upstream has no complete schema
   - OpenAPI assembled from schema + prose + route grammar

The intended derivation flow is:

```text
IIIF/api contexts + profile JSON
        +
validator schemas / fixtures
        +
prose spec constraints
        ▼
   derived JSON Schema
        ▼
      OpenAPI
        ▼
 downstream client/server generation
```

## Naming "done"

This repo is doing the right job when a downstream project can:

1. choose an IIIF API or extension
2. fetch a clearly-versioned JSON Schema and/or OpenAPI artifact
3. know whether that artifact is upstream-canonical or LibOps-derived
4. generate clients or wire types in its own language
5. validate emitted documents against the same machine-readable contract

The first proof point is `triplet`. A stronger signal will be when a second,
unrelated implementation can consume the same published artifacts without any
triplet-specific assumptions leaking through.
