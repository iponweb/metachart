# CLAUDE.md

## Project Overview

**Metachart** is a Go CLI tool that generates Helm Charts automatically from Kubernetes Resource JSON Schemas. It reads a single config file (`config/config.yaml`) and produces `values.schema.json` and Helm template files (`templates/generated/_settings.tpl`).

**Core workflow**:
```
config/config.yaml → metachart gen → values.schema.json + Helm templates
```

## Commands

```bash
# Build
./build/build.sh          # produces bin/metachart

# Test
./build/test.sh           # go test ./... with coverage

# Lint
./build/lint.sh           # gocyclo + golint + ineffassign + go vet

# Typical chart workflow
metachart init -r <chart-dir>   # initialize empty chart structure
metachart gen  -r <chart-dir>   # generate schema + templates

# Run generated chart
helm template -f values.yaml <chart-dir>
```

## Repository Layout

```
cmd/
  main.go                   # root CLI; version command
  app/
    command-init.go         # metachart init
    command-gen.go          # metachart gen (orchestrator)
    schema.go               # core schema generation logic
    templates.go            # generates templates/generated/_settings.tpl
    docs.go                 # generates docs/resources.md
    flags.go                # --root / -r flag
pkg/
  chart/
    config.go               # chart config structs + YAML I/O
    schema.go               # JSON Schema type definitions
    data.go                 # embedded init/gen template files
    autodiscover.go         # Autodiscover() — fetches discovery + definitions,
                            # filters with include/exclude selectors,
                            # emits ConversionRules + ResourceDefinitions
  helpers/
    helpers.go              # slice utilities (Index, Contains, Unique)
    file.go                 # FilePath supporting http/https/file://gitlab-api:// schemes
    gitlab.go               # GitLab blob URL parsing + API fetching
docs/
  generation.md             # complete gen pipeline documentation
  rendering.md              # Helm template rendering pipeline
  preprocessing.md          # preprocessor function reference
  config.md                 # config reference (partially outdated)
  gotpl.md                  # Go template JSON communication patterns
  quickstart.md             # quickstart guide
  values.md                 # values file reference
examples/
  demo/                     # minimal working example (Deployment + Service + ConfigMap)
  gateway-api/              # gateway-api CRDs via autodiscover (v1.5.1)
```

## Key Design Concepts

### Config File (inside a chart)
A single `config/config.yaml` with Kubernetes-object structure (`apiVersion / kind / metadata / spec`):

```yaml
apiVersion: metachart.iponweb.net/v1alpha1
kind: MetachartConfig
metadata:
  name: my-chart
spec:
  schema:
    definitions: [...]   # explicit _definitions.json URLs (supplement to autodiscover)
    rules: [...]         # manual ConversionRules
  autodiscover: [...]    # AutodiscoverSource entries
  resources: {}          # manual ResourceDefinition entries
```

`config/values.schema.custom.json` — custom JSON Schema additions merged into the final schema (always loaded last; can override any upstream definition).

### Autodiscover
Each `AutodiscoverSource` fetches a discovery JSON (Kubernetes `APIGroupList` or `APIResourceList`) and a `_definitions.json`, then emits `ConversionRules` and `ResourceDefinitions` for every resource that passes the `include`/`exclude` filters.

**Include rule enrichment** — each `include` selector may carry:
- `properties` — merged on top of the autodiscovered defaults (`enabled`, `metadata`)
- `disallowed` — appended to the autodiscovered defaults (`status`, `kind`, `apiVersion`)
- `related` — `[]ResourceSelector`; resolved cross-source after all sources are processed

**Merge precedence** (important):
- Autodiscovered `ConversionRules` are **prepended**; manual `spec.schema.rules` stay last and **win**
- Autodiscovered `ResourceDefinitions` are merged into `spec.resources`; manual entries **win**
- Autodiscovered definition URLs are prepended to `spec.schema.definitions`

**Reference name transformation**: source definition key (e.g. `io.k8s.api.apps.v1.Deployment`) is prefixed with `metachart.api.` to produce the ConversionRule target.

**Version sorting**: plural-only selectors resolve to the latest Kubernetes version — GA > beta > alpha, higher number within a tier wins.

### Generated Outputs
- `values.schema.json` — Helm values validation schema (descriptions stripped)
- `values.schema.full.json` — full schema including all referenced definitions with descriptions
- `templates/generated/_settings.tpl` — kind settings helper used by the rendering engine
- `docs/resources.md` — resource kind reference table
- `templates/_metachart.tpl` + `templates/resources.yaml` — static runtime files (identical across all charts, embedded in the binary)

### Rendering Pipeline (see `docs/rendering.md`)
1. Discover kinds → Discover resources → Build metadata
2. Apply `defaults` (deep merge; slices concatenate, not replace)
3. Preprocess (custom Go template functions)
4. Deep Render (template all string fields)
5. Render YAML

### Template Helpers (see `docs/preprocessing.md`)
- `metachart.fullname` — resource name
- `metachart.selectorLabels` — selector labels
- `metachart.resourceMeta` — full metadata block
- `metachart.setDefaults` — merge defaults into resource

### FilePath Schemes (`pkg/helpers/file.go`)
Supports fetching JSON Schema from: `http://`, `https://`, `file://`, `gitlab-api://` (uses `METACHART_GITLAB_API_TOKEN` + `METACHART_GITLAB_API_TOKEN_TYPE` env vars; token types: `private` (default), `oauth`, `job`)

## Dependencies

| Package | Purpose |
|---------|---------|
| `github.com/spf13/cobra` | CLI framework |
| `github.com/spf13/pflag` | Flag parsing |
| `github.com/xanzy/go-gitlab` | GitLab API client |
| `github.com/barkimedes/go-deepcopy` | Deep copy for defaults merging |
| `sigs.k8s.io/yaml` | YAML marshaling |

## Code Conventions

- All source files carry Apache 2.0 license header
- Linting threshold: cyclomatic complexity ≤ 20 (`gocyclo -over 20`)
- Tests use table-driven patterns
- JSON is used for inter-template communication in Go templates (see `docs/gotpl.md`)
- Config YAML uses block sequence format (`- item`) not JSON-style inline arrays

## Example Charts

| Chart | Description |
|-------|-------------|
| `examples/demo/` | Minimal working example: Deployment + Service + ConfigMap via autodiscover |
| `examples/gateway-api/` | Gateway API CRDs via autodiscover (v1.5.1) |

```bash
helm template -f examples/demo/values.example.yaml examples/demo/
helm template -f examples/gateway-api/values.example.yaml examples/gateway-api/
```

## Release

Releases are automated via GoReleaser (`.goreleaser.yaml`):
- Builds for Linux, Windows, Darwin (amd64 + arm64)
- Publishes to Homebrew tap `iponweb/tap`
