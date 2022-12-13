# Quickstart

Make an empty directory:

```shell
mkdir mychart
cd mychart
```

Create an empty chart:

```shell
metachart init
```

Add required schema definitions to the `config/schema.yaml` file:

```yaml
.defaults:
  disallowed: &defaultDisallowed
    - status
    - kind
    - apiVersion
  properties: &defaultProperties
    enabled: metachart.interface.boolean
    metadata: metachart.api.meta.v1.ObjectMeta
  rootKey: &defaultRootKey
    disallowed: *defaultDisallowed
    properties: *defaultProperties

definitions:
  - https://raw.githubusercontent.com/iponweb/schemas/main/json-schemas/external-secrets/v0.6.1-strict/_definitions.json

rules:
  #: Common
  #:
  #: meta.v1.ObjectMeta
  - target: metachart.api.meta.v1.ObjectMeta
    source: io.k8s.apimachinery.pkg.apis.meta.v1.ObjectMeta
    allowed:
      - annotations
      - labels
      - finalizers
      - namespace
      - name
    properties:
      checksums: metachart.interface.checksums

  #: rbac.authorization.k8s.io/v1
  - target: metachart.api.io.k8s.api.rbac.v1.Role
    source: io.k8s.api.rbac.v1.Role
    <<: *defaultRootKey
```

Add resource definition to the `config/resources.yaml` file:

```yaml
resources:
  #: rbac.authorization.k8s.io/v1
  roles:
    apiVersion: rbac.authorization.k8s.io/v1
    kind: Role
    jsonSchemaRef: metachart.api.io.k8s.api.rbac.v1.Role
```

Generate chart schema and templates:

```shell
metachart gen
```

Generate must be executed after any configuration file or preprocessor change.

Add generated schema to your favourite IDE. Examples:
- [IntelliJ IDEA](https://www.jetbrains.com/help/idea/json.html#ws_json_schema_add_custom)
- [Visual Studio Code](https://code.visualstudio.com/docs/languages/json#_json-schemas-and-settings)
- [Sublime Text](https://github.com/sublimelsp/LSP-json)

Create the `values.overrides.yaml` file

```yaml
roles:
  main:
    rules:
      - apiGroups:
          - ""
        verbs:
          - "*"
        resources:
          - "*"
```

Install the Helm release

```shell
helm upgrade --install -f values.overrides.yaml .
```
