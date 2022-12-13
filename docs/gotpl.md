# gotpl Guideline

`gotpl` is a templating language with some specifics. To use it as a form of
programming language, some rules have been developed and applied in
`metachart`.

There is a demonstration of a function consuming parameters and returning
complex value:

```gotemplate
{{- define "metachart.function" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $param1 := default nil $params.param1 }}
{{- $param2 := default nil $params.param2 }}
{{- /* Execution */}}
{{- $result := dict "param1" $param1 "param2" $param2 }}
{{- /* Return */}}
{{- $result | toJson }}
{{- end }}
```

Each function definition explicitly divided in separate sections:
- Cleanup context from the function params - prepares the `$context` variable
  with `$params` key excluded. Required to keep the root context untouched
  between function calls
- Get params - parse `$params` key
- Execution - prepares the body for the function output. Recommended to use the
  `$result` variable
- Return - return `$result` variable with explicit format definition

In the example above, `$result` returns in `json` format. It's highly
recommended to use `json` for communications between functions because it's
indentation agnostic.

Example: Call one function from another

```gotemplate
{{- define "metachart.function" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $param1 := default nil $params.param1 }}
{{- $param2 := default nil $params.param2 }}
{{- /* Execution */}}
{{- $result := (include "metachart.otherFunction" (merge (dict "params"
  (dict
    "$param1" $param1
    "$param2" $param2
  )) $context) | fromJson) }}
{{- /* Return */}}
{{- $result | toJson }}
{{- end }}
```

Example: Call function from the template

```gotemplate
{{- $param1 := "value1" }}
{{- $param2 := "value2" }}
{{- $result := (include "metachart.function" (merge (dict "params"
  (dict
    "$param1" $param1
    "$param2" $param2
  )) $) | fromJson) }}
```
