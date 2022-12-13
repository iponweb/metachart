{{/*
Create a default fully qualified app name.
Truncate at 63 chars because some Kubernetes name fields are limited to this
(by the DNS naming spec).
Use only helm release name because helm chart is made to be used by different
kinds of applications.

Return: string
*/}}
{{- define "metachart.fullname" -}}
{{- if $.Values.fullnameOverride }}
{{- $.Values.fullnameOverride | trunc 63 | trimSuffix "-" }}
{{- else }}
{{- $.Release.Name | trunc 63 | trimSuffix "-" }}
{{- end }}
{{- end }}

{{/*
Compute chart name-version to be used by the chart label

Return: string
*/}}
{{- define "metachart.chart" -}}
{{- printf "%s-%s" $.Chart.Name $.Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Compute Chart labels

Return: dict in json format
*/}}
{{- define "metachart.chartLabels" -}}
{{- $result := dict
  "helm.sh/chart" (include "metachart.chart" $)
  "app.kubernetes.io/instance" $.Release.Name
  "app.kubernetes.io/managed-by" $.Release.Service
}}
{{- if $.Chart.AppVersion }}
  {{- $_ := set $result "app.kubernetes.io/version" $.Chart.AppVersion }}
{{- end }}
{{- /* Return */}}
{{- $result | toJson }}
{{- end }}

{{/*
Label selector to determine if a resource belongs to a component. Only
necessary and sufficient set of labels is used.

Params:

  component : str - Component value which the resource belongs to

Return: dict in json format
*/}}
{{- define "metachart.selectorLabels" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $component := default nil $params.component }}
{{- /* Execution */}}
{{- $result := dict "app.kubernetes.io/instance" $.Release.Name }}
{{- with $component }}
  {{- $_ := set $result "app.kubernetes.io/component" $component}}
{{- end }}
{{- /* Return */}}
{{- $result | toJson }}
{{- end }}

{{/*
Compute resource labels

Params:

  definition : dict - Resource definition
  component : str - Resource component value
  relatedComponent : str - Related resource component value

Return: dict in json format
*/}}
{{- define "metachart.resourceLabels" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $definition := $params.definition }}
{{- $component := default nil $params.component }}
{{- $relatedComponent := default nil $params.relatedComponent }}
{{- /* Execution */}}
{{- $resourceMeta := default dict $definition.metadata }}
{{- $chartLabels := include "metachart.chartLabels" $context | fromJson }}
{{- $globalLabels := default dict (default dict (default dict $.Values.settings).global).labels | deepCopy }}
{{- $resourceLabels := default dict $resourceMeta.labels | deepCopy }}
{{- if $relatedComponent }}
  {{- $_ := set $chartLabels "app.kubernetes.io/component" $relatedComponent}}
{{- else if $component }}
  {{- $_ := set $chartLabels "app.kubernetes.io/component" $component}}
{{- end }}
{{- /* Validate if forbidden labels are used */}}
{{- range $reservedLabel := keys $chartLabels }}
  {{- if hasKey $globalLabels $reservedLabel }}
    {{- required (printf "Label %s is reserved for internal usage and can not be overrided" $reservedLabel) "" }}
  {{- end }}
  {{- if hasKey $resourceLabels $reservedLabel }}
    {{- required (printf "Label %s is reserved for internal usage and can not be overrided" $reservedLabel) "" }}
  {{- end }}
{{- end }}
{{- $result := merge $resourceLabels $globalLabels $chartLabels }}
{{- /* Return */}}
{{- $result | toJson }}
{{- end }}

{{/*
Compute checksums annotations for a resource

Params:

  definition : dict - Resource definition

Return: dict in json format
*/}}
{{- define "metachart.resourceChecksumAnnotations" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $definition := $params.definition }}
{{- /* Execution */}}
{{- $result := dict }}
{{- $checksums := default dict (default dict $definition.metadata).checksums }}
{{- range $kind, $names := $checksums }}
  {{- if $names }}
    {{- if (kindIs "slice" $names) }}
      {{- range $name := $names }}
        {{- $_ := set $result (printf "checksum-%s-%s" $kind $name) (include "metachart.checksumSingle" (merge (dict "params"
          (dict
            "kind" $kind
            "name" $name
          )) $)) }}
      {{- end }}
    {{- else if eq $names "*" }}
      {{- $_ := set $result (printf "checksum-%s" $kind) (include "metachart.checksumKinds" (merge (dict "params"
        (dict
          "kinds" (list $kind)
        )) $)) }}
    {{- end }}
  {{- end }}
{{- end }}
{{- /* Return */}}
{{- $result | toJson}}
{{- end }}

{{/*
Compute resource annotations

Params:

  definition : dict - Resource definition

Return: dict in json format
*/}}
{{- define "metachart.resourceAnnotations" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $definition := $params.definition }}
{{- /* Execution */}}
{{- $resourceMeta := default dict $definition.metadata }}
{{- $checksums := (include "metachart.resourceChecksumAnnotations" (merge (dict "params" $params) $context) | fromJson) }}
{{- $result := merge (default dict ($resourceMeta.annotations) | deepCopy) (default dict $.Values.annotations) $checksums }}
{{- /* Return */}}
{{- $result | toJson}}
{{- end }}

{{/*
Compute resource ObjectMeta

Params:

  definition : dict - Resource definition
  name : string - Resource name as defined in the values file
  nameSuffix : string - Suffix to be added to the resource name
  withName : bool - Whether name key must be added (default: true)
  withNameFullnamePrefix : bool - Whether fullname prefix must be added to the name (default: true)

Return: dict in json format
*/}}
{{- define "metachart.resourceMeta" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $definition := $params.definition }}
{{- $name := $params.name }}
{{- $nameSuffix := $params.nameSuffix }}
{{- $withName := (hasKey $params "withName" | ternary $params.withName true) }}
{{- $withNameFullnamePrefix := (hasKey $params "withNameFullnamePrefix" | ternary $params.withNameFullnamePrefix true) }}
{{- /* Execution */}}
{{- $resourceMeta := default dict $definition.metadata }}
{{- /* Execution */}}
{{- $result := omit $resourceMeta "labels" "annotations" "name" "checksums" }}
{{- $fullnamePrefix := "" }}
{{- if $withNameFullnamePrefix }}
{{- $fullnamePrefix = printf "%s-" (include "metachart.fullname" $context) }}
{{- end }}
{{- if $withName }}
  {{- if $resourceMeta.name }}
    {{- $_ := set $result "name" $resourceMeta.name }}
  {{- else if $name }}
    {{- if $nameSuffix }}
      {{- $_ := set $result "name" (printf "%s%s-%s" $fullnamePrefix $name $nameSuffix) }}
    {{- else }}
      {{- $_ := set $result "name" (printf "%s%s" $fullnamePrefix $name) }}
    {{- end }}
  {{- else }}
    {{- if $nameSuffix }}
      {{- $_ := set $result "name" (printf "%s%s" $fullnamePrefix $nameSuffix) }}
    {{- else }}
      {{- $_ := set $result "name" (printf "%s" (include "metachart.fullname" $context)) }}
    {{- end }}
  {{- end }}
{{- end }}
{{- $_ := set $result "labels" (include "metachart.resourceLabels" (merge (dict "params" $params) $context) | fromJson) }}
{{- $_ = set $result "annotations" (include "metachart.resourceAnnotations" (merge (dict "params" $params) $context) | fromJson) }}
{{- /* Return */}}
{{- $result | toJson }}
{{- end }}

{{/*
Discover all string values and render them as templates

Params:

  data : any - data to be processed

Return: dict in json format
*/}}
{{- define "metachart.deepRender" -}}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $data := ($params.data | deepCopy) }}
{{- /* Execution */}}
{{- if kindIs "string" $data }}
  {{- $result := tpl $data $context }}
  {{- /* Return */}}
  {{- $result | toJson }}
{{- else if kindIs "map" $data }}
  {{- $result := dict }}
  {{- range $key, $value := $data }}
    {{- if kindIs "string" $value }}
      {{- $newValue := tpl $value $context }}
      {{- $_ := set $result $key $newValue }}
    {{- else if kindIs "map" $value }}
      {{- $newValue := include "metachart.deepRender" (merge (dict "params" (dict "data" $value)) $context) | fromJson }}
      {{- $_ := set $result $key $newValue }}
    {{- else if kindIs "slice" $value }}
      {{- $newValue := include "metachart.deepRender" (merge (dict "params" (dict "data" $value)) $context) | fromJsonArray }}
      {{- $_ := set $result $key $newValue }}
    {{- else }}
      {{- $_ := set $result $key $value }}
    {{- end }}
  {{- end }}
  {{- $result | toJson }}
{{- else if kindIs "slice" $data }}
  {{- $result := list }}
  {{- range $value := $data }}
    {{- if kindIs "string" $value }}
      {{- $newValue := tpl $value $context }}
      {{- $result = append $result $newValue }}
    {{- else if kindIs "map" $value }}
      {{- $newValue := include "metachart.deepRender" (merge (dict "params" (dict "data" $value)) $context) | fromJson }}
      {{- $result = append $result $newValue }}
    {{- else if kindIs "slice" $value }}
      {{- $newValue := include "metachart.deepRender" (merge (dict "params" (dict "data" $value)) $context) | fromJsonArray }}
      {{- $result = append $result $newValue }}
    {{- else }}
      {{- $result = append $result $value }}
    {{- end }}
  {{- end }}
  {{- /* Return */}}
  {{- $result | toJson }}
{{- else }}
  {{- $result := $data }}
  {{- /* Return */}}
  {{- $result | toJson }}
{{- end }}
{{- end }}

{{/*
Deep merge like function which concats 2 slices if meets in instead of
override. This implementation takes into account that each of source and target
values can be none one of specific type or nil.

Params:

  source : dict | slice - Merge from
  target : dict | slice - Merge to

Return: dict | slice in json format
*/}}
{{- define "metachart.mergeConcatLists" }}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $source := ($params.source | deepCopy) }}
{{- $target := ($params.target | deepCopy) }}
{{- /* Execution */}}
{{- $result := dict }}
{{- $keys := keys (default dict $source) (default dict $target) | uniq }}
{{- range $key := $keys }}
  {{- $sourceValue := get $source $key }}
  {{- $targetValue := get $target $key }}
  {{- if or (kindIs "map" $sourceValue) (kindIs "map" $targetValue) }}
    {{- /* Normalize types */}}
    {{- if kindIs "map" $sourceValue }}{{- else }}{{- $sourceValue = dict }}{{- end }}
    {{- if kindIs "map" $targetValue }}{{- else }}{{- $targetValue = dict }}{{- end }}
    {{- $newValue := include "metachart.mergeConcatLists" (dict "params" (dict
      "source" $sourceValue
      "target" $targetValue
    )) | fromJson }}
    {{- $_ := set $result $key $newValue }}
  {{- else if or (kindIs "slice" $sourceValue) (kindIs "slice" $targetValue) }}
    {{- /* Normalize types */}}
    {{- if kindIs "slice" $sourceValue }}{{- else }}{{- $sourceValue = list }}{{- end }}
    {{- if kindIs "slice" $targetValue }}{{- else }}{{- $targetValue = list }}{{- end }}
    {{- $newValue := concat $sourceValue $targetValue }}
    {{- $_ := set $result $key $newValue }}
  {{- else if hasKey $target $key }}
    {{- $_ := set $result $key $targetValue }}
  {{- else }}
    {{- $_ := set $result $key $sourceValue }}
  {{- end }}
{{- end }}
{{- /* Return */}}
{{- $result | toJson }}
{{- end }}

{{/*
Discover defaults for the specific kind and apply them to the resource

Params:

  definition : dict - Resource definition
  kind : string - Resource kind

Return: dict in json format
*/}}
{{- define "metachart.setDefaults" }}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $definition := ($params.definition | deepCopy) }}
{{- $kind := $params.kind }}
{{- $kindSettings := ternary (get $.Values.settings $kind) (dict) (hasKey (default dict $.Values.settings) $kind) }}
{{- /* Execution */}}
{{- $defaults := default dict $kindSettings.defaults }}
{{- include "metachart.mergeConcatLists" (dict "params" (dict
  "source" $defaults
  "target" $definition
)) }}
{{- end }}

{{/*
Discover all available resources (standalone and related) of specific kind.
It takes into account:

- If resource kind is not disable in settings
- If resource definition.enabled is not false

Params:

  kind : string - Resource kind

Return: dict in json format
*/}}
{{- define "metachart.discover" }}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $kind := $params.kind }}
{{- /* Execution */}}
{{- $result := dict }}
{{- $kindSettings := dict }}
{{- if hasKey (default dict $.Values.settings) $kind }}
  {{- $kindSettings = get (default dict $.Values.settings) $kind }}
{{- end }}
{{- if not $kindSettings.disabled }}
  {{- if hasKey $.Values $kind }}
    {{- range $resourceName, $resourceDefinition := get $.Values $kind }}
      {{- if (hasKey $resourceDefinition "enabled" | ternary $resourceDefinition.enabled true) }}
        {{- $_ := set $result $resourceName $resourceDefinition}}
      {{- end }}
    {{- end }}
  {{- end }}
  {{- /* Discover related resources */}}
  {{- $settingsKindAll := include "metachart.settings" $context | fromYaml | keys }}
  {{- range $settingsKind := $settingsKindAll }}
    {{- $resourceSettings := dict }}
    {{- if hasKey (default dict $.Values.settings) $settingsKind }}
      {{- $resourceSettings = get (default dict $.Values.settings) $settingsKind }}
    {{- end }}
    {{- if not $resourceSettings.disabled }}
      {{- if hasKey $.Values $settingsKind }}
        {{- range $resourceName, $resourceDefinition := get $.Values $settingsKind }}
          {{- if (hasKey $resourceDefinition "enabled" | ternary $resourceDefinition.enabled true) }}
            {{- $resourceRelated := default dict $resourceDefinition.related }}
             {{- if hasKey $resourceRelated $kind }}
              {{- range $name, $definition := get $resourceRelated $kind }}
                {{- if hasKey $result $name }}
                  {{- fail (printf "Resource %s/%s defined in global and related scopes" $kind $name) }}
                {{- else }}
                  {{- $patchedDefinition := $definition | deepCopy }}
                  {{- $_ := set $patchedDefinition "relatedComponent" (printf "%s-%s" $settingsKind $resourceName)}}
                  {{- $_ = set $result $name $patchedDefinition}}
                {{- end }}
              {{- end }}
            {{- end }}
          {{- end }}
        {{- end }}
      {{- end }}
    {{- end }}
  {{- end }}
{{- end }}
{{- /* Return */}}
{{- $result | toJson }}
{{- end }}

{{/*
Build a complete resource ready for rendering

Params:

  name : string - Resource name as defined in the values file
  kind : string - Resource kind
  definition : dict - Resource definition
  apiVersion : strint - Resource ApiVersion
  kindCamelCase : string - Resource kind in CamelCase format
  preprocess : bool - Whether the resource kind has a preprocessor

Return: dict in json format
*/}}
{{- define "metachart.buildResource" }}
{{- /* Cleanup context from the function params */}}
{{- $params := $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $name := $params.name }}
{{- $definition := $params.definition }}
{{- $apiVersion := $params.apiVersion }}
{{- $kindCamelCase := $params.kindCamelCase }}
{{- $preprocess := $params.preprocess }}
{{- $kind := $params.kind }}
{{- /* Execution */}}
{{- $component := (printf "%s-%s" $kind $name) }}
{{- $relatedComponent := $definition.relatedComponent }}
{{- $resource := dict
  "apiVersion" $apiVersion
  "kind" $kindCamelCase
}}
{{- $resource = merge $resource (omit ($definition | deepCopy) "enabled" "metadata" "related" "relatedComponent") }}
{{- $_ := set $resource "metadata" (include "metachart.resourceMeta" (merge (dict "params"
  (dict
    "definition" $definition
    "name" $name
    "component" $component
    "relatedComponent" $relatedComponent
  )) $context) | fromJson) }}
{{- /* Apply defaults */}}
{{- $resource = include "metachart.setDefaults" (merge (dict "params"
  (dict
    "definition" $resource
    "kind" $kind
  )) $context) | fromJson }}
{{- /* Preprocessing */}}
{{- $preprocessed := $resource }}
{{- if $preprocess }}
  {{- $preprocessor := printf "metachart.preprocess.%s" $kind }}
  {{- $preprocessed = include $preprocessor (merge (dict "params"
    (dict
      "definition" $resource
      "name" $name
      "component" $component
      "relatedComponent" $relatedComponent
    )) $context) | fromJson }}
{{- end }}
{{- /* Render */}}
{{- $result := include "metachart.deepRender" (merge (dict "params" (dict "data" $preprocessed)) $) | fromJson }}
{{- /* Return */}}
{{- $result | toJson }}
{{- end }}

{{/*
Render all release resources

Return: yaml
*/}}
{{- define "metachart.renderAll" }}
{{- /* Cleanup context from the function params */}}
{{- $params := default dict $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- /* Execution */}}
{{- $kinds := (include "metachart.settings" $context) | fromYaml | keys }}
{{- $result := include "metachart.renderKinds" (merge (dict "params"
    (dict
      "kinds" $kinds
    )) $) }}
{{- /* Return */}}
{{- $result }}
{{- end }}

{{/*
Render specific resource kinds

Params:

  kinds : slice - List of kinds

Return: yaml
*/}}
{{- define "metachart.renderKinds" }}
{{- /* Cleanup context from the function params */}}
{{- $params := default dict $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $kinds := $params.kinds }}
{{- /* Execution */}}
{{- $result := "" }}
{{- range $kind := $kinds }}
  {{- $settings := get (include "metachart.settings" $context | fromYaml) $kind }}
  {{- range $name, $definition := (include "metachart.discover" (merge (dict "params"
    (dict
      "kind" $kind
    )) $context) | fromJson) }}
  {{- $rendered := include "metachart.buildResource" (merge (dict "params"
    (dict
      "kind"          $kind
      "apiVersion"    $settings.apiVersion
      "kindCamelCase" $settings.kindCamelCase
      "name"          $name
      "definition"    $definition
      "preprocess"    $settings.preprocess
    )) $context) | fromJson | toYaml | nindent 0 }}
  {{- $result = printf "%s---%s\n...\n" $result $rendered }}
  {{- end }}
{{- end }}
{{- /* Return */}}
{{- $result }}
{{- end }}

{{/*
Calculate checksum of resources of specific kind

Params:

  kinds : slice - List of kinds

Return: string
*/}}
{{- define "metachart.checksumKinds" }}
{{- /* Cleanup context from the function params */}}
{{- $params := default dict $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- /* Execution */}}
{{- $result := include "metachart.renderKinds" (merge (dict "params"
    (dict
      "params" $params
    )) $) }}
{{- /* Return */}}
{{- $result | sha256sum }}
{{- end }}

{{/*
Render single resource

Params:

  name : string - Resource name as defined in the values file
  kind : string - Resource kind

Return: yaml
*/}}
{{- define "metachart.renderSingle" }}
{{- /* Cleanup context from the function params */}}
{{- $params := default dict $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- $kind := $params.kind }}
{{- $name := $params.name }}
{{- /* Execution */}}
{{- $settings := get (include "metachart.settings" $context | fromYaml) $kind }}
{{- $definitionsAll := (include "metachart.discover" (merge (dict "params"
  (dict
    "kind" $kind
  )) $context) | fromJson) }}
{{- if hasKey $definitionsAll $name }}{{- else }}{{- fail (printf "can not find resource %s/%s" $kind $name)}}{{- end }}
{{- $definition := get $definitionsAll $name }}
{{- $result := include "metachart.buildResource" (merge (dict "params"
  (dict
    "kind"          $kind
    "apiVersion"    $settings.apiVersion
    "kindCamelCase" $settings.kindCamelCase
    "name"          $name
    "definition"    $definition
    "preprocess"    $settings.preprocess
  )) $context) | fromJson | toYaml | nindent 0 }}
{{- $result }}
{{- end }}

{{/*
Calculate checksum of a specific resource

Params:

  name : string - Resource name as defined in the values file
  kind : string - Resource kind

Return: string
*/}}
{{- define "metachart.checksumSingle" }}
{{- /* Cleanup context from the function params */}}
{{- $params := default dict $.params | deepCopy }}
{{- $context := omit $ "params" }}
{{- /* Get params */}}
{{- /* Execution */}}
{{- $result := include "metachart.renderSingle" (merge (dict "params"
    (dict
      "params" $params
    )) $) }}
{{- /* Return */}}
{{- $result | sha256sum }}
{{- end }}
