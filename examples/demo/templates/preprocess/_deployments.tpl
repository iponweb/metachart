{{- define "metachart.preprocess.deployments" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $definition := $params.definition }}
{{- $name := $params.name }}
{{- $component := $params.component }}
{{- /* Execution */}}
{{- /* Set spec.selector */}}
{{- $_ := set $definition.spec "selector" (dict "matchLabels" (include "metachart.selectorLabels" (merge (dict "params"
  (dict
    "component" $component
  )) $context) | fromJson) ) }}
{{- /* Read template definition */}}
{{- $template := get $definition.spec "template" | deepCopy }}
{{- $_ = unset $definition.spec "template" }}
{{- /* Apply metadata */}}
{{- $_ = set $template "metadata" (include "metachart.resourceMeta" (merge (dict "params"
  (dict
    "definition" $template
    "name" $name
    "component" $component
    "withName" false
  )) $context) | fromJson) }}
{{- /* Set spec.template */}}
{{- $_ = set $definition.spec "template" $template }}
{{- /* Return */}}
{{- $definition | toJson }}
{{- end }}
