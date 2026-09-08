{{- /* Resources definition */}}
{{- define "metachart.settings" }}
configmaps:
  apiVersion: v1
  kindCamelCase: ConfigMap
  preprocess: false
deployments:
  apiVersion: apps/v1
  kindCamelCase: Deployment
  preprocess: true
pods:
  apiVersion: v1
  kindCamelCase: Pod
  preprocess: false
services:
  apiVersion: v1
  kindCamelCase: Service
  preprocess: true
{{- end }}

{{- /* Root keys merged from global.<key> by metachart.applyGlobal; settings is always merged */}}
{{- define "metachart.globalKeys" }}
[]
{{- end }}