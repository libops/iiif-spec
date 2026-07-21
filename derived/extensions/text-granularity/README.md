# Text Granularity Extension

This directory contains derived machine-readable artifacts for the IIIF Text
Granularity extension:

https://iiif.io/api/extension/text-granularity/

The extension is prose-only today for most implementation purposes. The schema
under `schema/annotation.schema.json` is therefore LibOps-derived rather than
upstream-canonical. It models the extension's normative requirement that
`textGranularity` be a single string. It deliberately does not use an enum or
impose a non-empty constraint: the six terms in the published JSON-LD context
and values from the extension registry are recommendations rather than
MUST-level constraints.

The importable Go validator at
`github.com/libops/iiif-spec/extension/textgranularity/schema` composes this
artifact with the Presentation 3 validators and verifies JSON-LD context order.
The upstream Presentation context, Text Granularity context, prose contract,
and validator fixtures remain the provenance sources.
