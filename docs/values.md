# Values file

Complete expected values file layout example with only one kind named `kind`
supported and one resource with name `name` defined.

```yaml
# context is a recommended way to define user-provided free form data to be
# used in templates
context: {}

# fullnameOverride overrides value of $.Release.Name used for release resources
# name prefixes
fullnameOverride:

# global holds values shared with sub-charts: settings and every root key
# marked `global: true` in the chart config are merged into the top-level keys
# of the same name (local keys win, lists are concatenated)
global:
  settings: {}

# settings configures chart behaviour
settings:
  global:
    # labels to be applied to all release resources
    labels: {}

    # annotations to be applied to all release resources
    annotations: {}

  kind:
    #: Whether all resources of the kind must be disabled
    disabled: false

    #: Defaults to be applied to all resources of the kind
    defaults: {}

#: kind resources definition
kind:
  name:
    # enabled indicates whether the resource must be rendered: a boolean or a
    # template string rendering to true/false
    enabled: true

    metadata:
      # checksums feature settings
      checksums: {}

    # related feature settings
    related: {}
```
