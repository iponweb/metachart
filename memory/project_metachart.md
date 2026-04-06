---
name: metachart project overview
description: Key design decisions, architecture, and implementation state of the metachart tool
type: project
---

Metachart is a Go CLI that generates Helm Charts from Kubernetes API JSON Schemas.

**Why:** Eliminates manual maintenance of values.schema.json and _settings.tpl as Kubernetes API evolves.

**How to apply:** When making changes, understand that Autodiscover prepends rules (manual rules win by being last), and ResourceSelector.Related is []ResourceSelector (not map).

## Key architecture decisions

- Autodiscover rules are PREPENDED to spec.schema.rules; manual rules at the end WIN (run last in GenSchema, overwrite autodiscovered definitions)
- ResourceDefinitions from autodiscover are MERGED into spec.resources; manual entries always win
- Related selectors on include rules resolve AFTER all sources processed (two-pass, cross-source works)
- External CRDs not in any autodiscover source must use manual ConversionRules + resources

## Reference name transformation

Source key: `io.k8s.api.apps.v1.Deployment`
Target: `metachart.api.io.k8s.api.apps.v1.Deployment`
Rule: prefix `metachart.api.` to the source definition key from the definitions index

## Examples

- `examples/demo/` — minimal: Deployment + Service + ConfigMap via autodiscover (k8s v1.35.3)
- `examples/gateway-api/` — gateway-api CRDs via autodiscover (v1.5.1)

## Documentation

- `docs/generation.md` — complete gen pipeline docs (written 2026-03-29)
- `docs/rendering.md` — Helm template rendering pipeline
- `docs/preprocessing.md` — preprocessor function reference
