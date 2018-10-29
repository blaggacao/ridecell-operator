{{ define "componentName" }}daphne{{ end }}
{{ define "componentType" }}web{{ end }}
{{ define "ingressPath" }}/websockets{{ end }}
{{ template "ingress" . }}
