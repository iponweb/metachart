# Preprocessing

List of recommended template function to be used in preprocessors:

- `metachart.fullname`
- `metachart.selectorLabels`
- `metachart.resourceMeta`
- `metachart.setDefaults`

Preprocessors receive the chart context plus `$.Metachart.Resource` (kind,
name, component and the resource `context`, see
[Resource context](rendering.md#resource-context)), so nested `enabled`
templates evaluated by `metachart.enabled` / `metachart.filterEnabled` can
depend on the resource being built.

See the [_metachart.tpl](/pkg/chart/resources/init/templates/_metachart.tpl)
file for complete available functions list.
