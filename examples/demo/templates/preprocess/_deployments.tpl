{{- define "metachart.preprocess.deployments" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $resource := $params.resource }}
{{- $name := $params.name }}
{{- $component := $params.component }}
{{- /* Execution */}}
{{- /* Set spec.selector */}}
{{- $_ := set $resource.spec "selector" (dict "matchLabels" (include "metachart.selectorLabels" (merge (dict "params"
  (dict
    "component" $component
  )) $context) | fromJson) ) }}
{{- /* Read template definition */}}
{{- $template := get $resource.spec "template" | deepCopy }}
{{- $_ = unset $resource.spec "template" }}
{{- /* Apply metadata */}}
{{- $_ = set $template "metadata" (include "metachart.resourceMeta" (merge (dict "params"
  (dict
    "resource" $template
    "resourceName" $name
    "component" $component
    "withName" false
  )) $context) | fromJson) }}
{{- /* Set spec.template */}}
{{- $_ = set $resource.spec "template" $template }}
{{- /* Return */}}
{{- $resource | toJson }}
{{- end }}
