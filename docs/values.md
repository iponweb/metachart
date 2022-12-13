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
    # enabled indicates whether the resource must be rendered
    enabled: true

    metadata:
      # checksums feature settings
      checksums: {}

    # related feature settings
    related: {}
```
