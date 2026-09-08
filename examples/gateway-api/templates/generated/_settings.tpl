{{- /* Resources definition */}}
{{- define "metachart.settings" }}
backendtlspolicies:
  apiVersion: gateway.networking.k8s.io/v1
  kindCamelCase: BackendTLSPolicy
  preprocess: false
gatewayclasses:
  apiVersion: gateway.networking.k8s.io/v1
  kindCamelCase: GatewayClass
  preprocess: false
gateways:
  apiVersion: gateway.networking.k8s.io/v1
  kindCamelCase: Gateway
  preprocess: false
grpcroutes:
  apiVersion: gateway.networking.k8s.io/v1
  kindCamelCase: GRPCRoute
  preprocess: false
httproutes:
  apiVersion: gateway.networking.k8s.io/v1
  kindCamelCase: HTTPRoute
  preprocess: false
listenersets:
  apiVersion: gateway.networking.k8s.io/v1
  kindCamelCase: ListenerSet
  preprocess: false
referencegrants:
  apiVersion: gateway.networking.k8s.io/v1
  kindCamelCase: ReferenceGrant
  preprocess: false
tlsroutes:
  apiVersion: gateway.networking.k8s.io/v1
  kindCamelCase: TLSRoute
  preprocess: false
{{- end }}

{{- /* Root keys merged from global.<key> by metachart.applyGlobal; settings is always merged */}}
{{- define "metachart.globalKeys" }}
[]
{{- end }}