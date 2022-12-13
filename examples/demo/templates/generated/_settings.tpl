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
services:
  apiVersion: v1
  kindCamelCase: Service
  preprocess: true
{{- end }}