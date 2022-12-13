# Config

Each Chart configuration is split in 2 parts: schema, resources and custom
definition.

## Schema

Path: `config/schema.yaml`

JSON Schema definitions mapping to be used to generate the Chart's JSON schema.

Keys:

- `definitions` - List URI's to files containing JSON Schema definitions.

  Supported URL schemas:
  - `http` and `https`
  - `file` - Local file absolute path
  - `gitlab-api` - Access files in private Gitlab installations using API.
    Requires `METACHART_GITLAB_API_TOKEN` environment variable
    containing Gitlab API token to be defined.

    Example:

    ```
    gitlab-api://gitlab.example.net/path/to/the/repo/-/blob/main/path/to/the/file/json
    ```

- `rules` - List of definitions conversion rules
- `rules[].source` - Definition name existing in one of JSON Schemas provided
  above
- `rules[].target` - Definition name to be used in the chart
- `rules[].allowed` - Allowed properties list
- `rules[].required` - Additional required properties list
- `rules[].disallowed` - Disallowed properties list
- `rules[].properties` - Custom properties to be added or to override existing
  ones
- `rules[].related` - Map of Kind and JSON Schema reference of resources
  allowed to be related.

`metachart init` creates this file with:

- `defaultDisallowed`, `defaultProperties` and `defaultRootKey` anchors
  simplifying rules definition
- `metachart.api.meta.v1.ObjectMeta` - ObjectMeta definition recommended to
  be used because it includes everything required for the `checksums` feature
  work

See also: [schema.yaml](../pkg/chart/resources/init/config/schema.yaml) base
content.

## Resources

Path: `config/resources.yaml`

Values interface configuration.

Keys:

- `resources.KIND` - Settings of a resource kind `KIND`. As kind recommended
  to use plural lowercase version. Example: `configmaps`
- `resources.KIND.apiVerion` - `ApiVersion` to be used in the resource
  definition
- `resources.KIND.kind` - `Kind` to be used in the resource definition
- `resources.KIND.jsonSchemaRef` - JSON Schema definition reference defined
  in the [config/schema.yaml](#Schema).
- `resources.KIND.template` - Whether the kind must be rendered.
  Default - `true`
- `resources.KIND.defaults` - Whether the kind must have the `defaults`
  feature enabled. Default - `true`

Example:
```yaml
resources:

  #: rbac.authorization.k8s.io/v1
  roles:
    apiVersion: rbac.authorization.k8s.io/v1
    kind: Role
    jsonSchemaRef: metachart.api.io.k8s.api.rbac.v1.Role
```
