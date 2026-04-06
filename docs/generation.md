# metachart gen — Generation Pipeline

This document describes the complete process that `metachart gen` executes to
produce `values.schema.json` and Helm template files from a chart config.

---

## Pipeline Overview

```
config/config.yaml
  │
  ├─ 1. Config load
  │
  ├─ 2. Autodiscover
  │       ├─ Fetch discovery JSON (APIGroupList or APIResourceList)
  │       ├─ Fetch definitions JSON, build GVK index
  │       ├─ Filter resources with include/exclude selectors
  │       ├─ Emit ConversionRules + ResourceDefinitions
  │       └─ Resolve cross-source `related` selectors (pass 2)
  │
  ├─ 3. Cleanup templates/generated/
  │
  ├─ 4. GenTemplates  →  templates/generated/_settings.tpl
  │
  ├─ 5. GenSchema     →  values.schema.json
  │                   →  values.schema.full.json
  │
  ├─ 6. GenDocs       →  docs/resources.md
  │
  └─ 7. WriteGen      →  templates/_metachart.tpl
                      →  templates/resources.yaml
```

---

## 1. Config Load

`metachart gen` reads `config/config.yaml`, which is a single YAML file that
follows the Kubernetes object convention (`apiVersion / kind / metadata / spec`).
The top-level spec contains three sections:

| Field | Purpose |
|---|---|
| `spec.schema.definitions` | Explicit paths / URLs of `_definitions.json` files |
| `spec.schema.rules` | Manual `ConversionRule` entries |
| `spec.autodiscover` | Zero or more `AutodiscoverSource` entries |
| `spec.resources` | Manual `ResourceDefinition` entries |

After `Autodiscover()` runs (step 2), the in-memory config is augmented:
discovered rules are **prepended** to `spec.schema.rules` and discovered
resources are **merged** into `spec.resources` (manual entries win).

---

## 2. Autodiscover

Each `AutodiscoverSource` is processed in two passes.

### 2a. URL Template Rendering

If an `AutodiscoverSource` has a non-empty `version` field, the strings
`discovery` and `definitions` are treated as Go `text/template` templates where
`{{ .version }}` is substituted before the URLs are fetched.  This lets you pin
a single version in one place:

```yaml
version: "v1.35.3"
discovery: "https://…/{{ .version }}/discovery/apis.json"
definitions: "https://…/{{ .version }}/json-schema/source/_definitions.json"
```

### 2b. Discovery Format Detection

The fetched discovery document can be in one of two Kubernetes formats:

| Format | When used | `kind` field |
|---|---|---|
| `APIGroupList` | Named API groups (`apis.json`) | absent / other |
| `APIResourceList` | Core API group (`api__v1.json`) | `"APIResourceList"` |

For `APIGroupList`, each group version is listed in `groups[*].versions`.
Metachart derives the per-group-version URL from the `apis.json` URL by replacing
the filename:

```
…/discovery/apis.json  +  "apps/v1"
  →  …/discovery/apis__apps__v1.json
```

### 2c. Definition Index

The `_definitions.json` file is indexed by `(group, version, kind)` using the
`x-kubernetes-group-version-kind` annotation present on each root definition.
This index maps a GVK to the definition key string, e.g.:

```
{group: "apps", version: "v1", kind: "Deployment"}
  →  "io.k8s.api.apps.v1.Deployment"
```

### 2d. Reference Name Transformation

Every autodiscovered resource definition key is prefixed with `metachart.api.`
to produce the ConversionRule target:

```
io.k8s.api.apps.v1.Deployment
  →  metachart.api.io.k8s.api.apps.v1.Deployment
```

This prefix namespaces all metachart-managed definitions inside the final JSON
Schema, keeping them separate from the upstream Kubernetes definitions.

The same transformation applies to external CRDs:

```
io.external-secrets.apis.externalsecrets.v1.ExternalSecret
  →  metachart.api.io.external-secrets.apis.externalsecrets.v1.ExternalSecret
```

Manual `ConversionRule` entries in `spec.schema.rules` must also use this
`metachart.api.*` target naming convention.

### 2e. Include / Exclude Filtering

A resource is included when it matches at least one `include` selector (or
`include` is empty) **and** does not match any `exclude` selector.

Three selector modes are available, determined by which fields are set:

| Fields set | Match condition |
|---|---|
| `apiVersion` + `kind` | apiVersion glob-matches **and** at least one kind element glob-matches (empty `kind` = all kinds in that apiVersion) |
| `apiVersion` + `plural` | apiVersion glob-matches **and** at least one plural element glob-matches |
| `plural` only | plural glob-matches **and** this is the latest Kubernetes version for that resource |

The `"*"` wildcard matches any sequence of characters in any field.

**Plural-only mode and version selection:**

When no `apiVersion` is given, metachart picks the *latest* stable version of
the resource by comparing version strings:

- Tier: GA (`v1`) > beta (`v1beta2`) > alpha (`v1alpha1`)
- Within a tier: higher number wins (`v2` > `v1`, `v1beta2` > `v1beta1`)

### 2f. Include Rule Enrichment

When a resource matches an include selector, that selector's optional fields
are applied on top of the autodiscovery defaults:

**Default ConversionRule properties** (added to every autodiscovered resource):
```yaml
properties:
  enabled: metachart.interface.boolean
  metadata: metachart.api.meta.v1.ObjectMeta
```

**Default disallowed fields** (removed from every autodiscovered resource):
```
status, kind, apiVersion
```

**`properties` on an include selector** merges additional property references
into the rule, with selector values taking precedence:

```yaml
- apiVersion: "apps/v1"
  kind:
    - Deployment
  properties:
    spec: metachart.api.io.k8s.api.apps.v1.DeploymentSpec   # added / overrides
```

**`disallowed` on an include selector** appends additional field names to the
default list (the list is extended, never replaced):

```yaml
- apiVersion: "v1"
  kind:
    - Pod
  disallowed:
    - ephemeralContainers    # appended to [status, kind, apiVersion]
```

**`related` on an include selector** schedules a cross-source related
resolution (see 2g below):

```yaml
- apiVersion: "apps/v1"
  kind:
    - Deployment
  related:
    - plural:
        - services
        - poddisruptionbudgets
```

### 2g. Cross-Source Related Resolution (Pass 2)

`related` selectors on include rules are resolved **after** all autodiscover
sources have been processed.  This allows a resource in one source to reference
resources discovered by a different source.

Resolution iterates every entry in the combined resource map (explicit
`spec.resources` + all autodiscovered resources, explicit entries win).  For
each resource that matches the related selector, its plural name becomes the
key and its `jsonSchemaRef` becomes the value in the ConversionRule's `related`
map:

```yaml
related:
  services: metachart.api.io.k8s.api.core.v1.Service
  poddisruptionbudgets: metachart.api.io.k8s.api.policy.v1.PodDisruptionBudget
```

A single `related` selector may list multiple plurals; all matching resources
are collected (no early exit):

```yaml
related:
  - plural:
      - services
      - poddisruptionbudgets   # both resolved into the related map
```

### 2h. Merge Into Config

After both passes:

- Discovered ConversionRules are **prepended** to `spec.schema.rules`.
  Because GenSchema processes rules in order and each rule overwrites the same
  target key, rules appearing **later** win.  Manual rules defined in the config
  stay at the end and therefore take precedence over autodiscovered rules for
  the same target.

- Discovered `ResourceDefinition`s are merged into `spec.resources`.  Manual
  entries with the same key win (they are never overwritten).

- Discovered definition file URLs are collected deduplicated into a final list
  and **prepended** to `spec.schema.definitions`.  The collection order is
  controlled by the `definitionsLast` field on each `AutodiscoverSource`:
  sources with `definitionsLast: false` (default) appear first in source order,
  sources with `definitionsLast: true` are appended after.  Because
  `GenSchema` merges definitions with last-write-wins, sources marked
  `definitionsLast: true` take precedence over all other autodiscover sources.

  Use this for Kubernetes core definitions to prevent CRD bundles (which often
  bundle stale copies of core Kubernetes types) from shadowing the authoritative
  definitions:

  ```yaml
  - version: "v1.35.3"
    definitionsLast: true
    discovery: "https://…/discovery/apis.json"
    definitions: "https://…/json-schema/source/_definitions.json"
  ```

---

## 3. Schema Generation (`GenSchema`)

### 3a. Load Definitions

All `spec.schema.definitions` files are loaded and merged into a flat
`definitions` map.  The custom file `config/values.schema.custom.json` is
always loaded last, allowing it to override anything from upstream sources.

### 3b. Process ConversionRules

Each `ConversionRule` is processed in order.  The result is stored in
`schema.definitions[rule.Target]`, overwriting any previous entry.  This is
how manual rules override autodiscovered ones for the same target.

**Rule processing steps:**

1. **Source lookup**: deep-copy the source definition from the loaded
   definitions map, or create an empty object if `source` is null.

2. **Disallowed** (`rule.Disallowed`): remove listed keys from `properties`
   and from the `required` array.

3. **Allowed** (`rule.Allowed`): keep *only* listed keys in `properties` and
   `required` (takes effect after Disallowed).

4. **Properties** (`rule.Properties`): for each `key: schemaRef` entry, add
   `{ "$ref": "#/definitions/<schemaRef>" }` to `properties`.  Existing keys
   are overwritten.

5. **Required** (`rule.Required`): append the listed field names to the
   existing `required` array.

6. **Related** (`rule.Related`): build a nested `related` property — an object
   with one key per entry where each key maps to a `patternProperties` root-key
   schema (FQDN-keyed map of the related resource type).

### 3c. Root-Key Properties

For every resource with `root: true`, a `patternProperties` entry is added to
the top-level schema `properties`:

```json
"deployments": {
  "type": "object",
  "patternProperties": {
    "^[a-z][0-9a-z]*(-[0-9a-z]+)*$": { "$ref": "#/definitions/metachart.api.io.k8s.api.apps.v1.Deployment" }
  },
  "additionalProperties": false
}
```

### 3d. Settings Schema

For every resource with `root: true` or `defaults: true`, a corresponding
entry is added under `settings.properties`:

- `root: true` → `settings.<kind>.disabled` (boolean flag to disable all resources of that kind)
- `defaults: true` → `settings.<kind>.defaults` (reference to the resource schema, used as a defaults template)

### 3e. Checksums Definition

The `metachart.interface.checksums` definition is built dynamically: every
`root: true` resource gets a key in `checksums.properties` mapping to a
`checksumEntryList` schema.

### 3f. Unused Definition Pruning

`CollectUsedDefinitions` performs a transitive closure over all `$ref`
pointers reachable from `schema.properties`.  Any definition not reachable is
removed from the final schema.

### 3g. Output Files

| File | Content |
|---|---|
| `values.schema.full.json` | All definitions including description fields |
| `values.schema.json` | Same structure with all `description` fields stripped (smaller, used by Helm) |

---

## 4. Template Generation (`GenTemplates`)

Produces `templates/generated/_settings.tpl`, which defines the
`metachart.settings` helper used by the rendering engine.

For each resource with `template: true`, a `KindSettings` entry is emitted:

```yaml
deployments:
  apiVersion: apps/v1
  kindCamelCase: Deployment
  preprocess: true    # true when templates/preprocess/_deployments.tpl exists
```

Preprocessor detection: `templates/preprocess/_<kind>.tpl` is checked via
`os.Stat`; if the file exists, `preprocess: true` is set.

---

## 5. Docs Generation (`GenDocs`)

Produces `docs/resources.md` — a Markdown table listing every resource kind
and its `apiVersion`, `jsonSchemaRef`, and flags (`template`, `root`,
`defaults`).

---

## 6. Static File Generation (`WriteGen`)

Writes the two runtime files that are identical across all metachart charts:

| File | Purpose |
|---|---|
| `templates/_metachart.tpl` | Core Go template helpers (rendering engine) |
| `templates/resources.yaml` | Entry point: `{{- include "metachart.renderAll" $ }}` |

These files are embedded in the `metachart` binary and are safe to overwrite on
every `metachart gen` run.

---

## Manual vs Autodiscovered Rules

Since autodiscovered rules are prepended and manual rules stay last,
**manual rules in `spec.schema.rules` always override autodiscovered rules
for the same target**.  This means you can:

- Use autodiscover for the broad set of resources.
- Override only the resources that need extra customisation (e.g. adding
  `related` pointing to external CRDs not covered by any autodiscover source).

Example: let autodiscover handle all `apps/v1` resources, then override
`Deployment` with a manual rule to add `podmonitors` to its `related` map.

---

## FilePath Schemes

Definition and discovery URLs support these schemes:

| Scheme | Notes |
|---|---|
| `https://` / `http://` | Standard HTTP fetch |
| `file://` | Local filesystem |
| `gitlab-api://` | GitLab Blob API; requires `METACHART_GITLAB_API_TOKEN` env var |

**GitLab token types** — set `METACHART_GITLAB_API_TOKEN_TYPE` to control how
the token is sent:

| Value | Header sent | When to use |
|---|---|---|
| `private` (default) | `PRIVATE-TOKEN` | Personal access token |
| `oauth` | `Authorization: Bearer` | OAuth / OpenID token |
| `job` | `JOB-TOKEN` | GitLab CI job token |
