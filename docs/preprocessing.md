# Preprocessing

List of recommended template function to be used in preprocessors:

- `metachart.fullname`
- `metachart.selectorLabels`
- `metachart.resourceMeta`
- `metachart.setDefaults`

Preprocessors receive the chart context plus `$.Metachart.Context` (the
effective context, see [Resource context](rendering.md#resource-context)) and
`$.Metachart.Resource` (kind, name, component, context), so nested `enabled`
templates evaluated by `metachart.enabled` / `metachart.filterEnabled` can
depend on the resource being built. A preprocessor that descends into a nested
definition with its own `context` (for example a container) merges that
context over `$.Metachart.Context`, replaces `Metachart.Context` in the
context it passes down and strips the key from the output, so that the same
`$.Metachart.Context.<key>` reads the right layer everywhere.

See the [_metachart.tpl](/pkg/chart/resources/init/templates/_metachart.tpl)
file for complete available functions list.

Gotcha: build the `Metachart` dict with `set`, not with sprig `merge`. `merge`
treats `false` as an empty value and fills it from the other dict, so a
`false` in a narrowed context would silently turn back into the `true` of the
outer layer. `metachart.mergeConcatLists` does not have this problem.
