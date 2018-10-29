{{ define "componentName" }}static{{ end }}
{{ define "componentType" }}web{{ end }}
{{ define "ingressPath" }}/static{{ end }}
{{ template "ingress" . }}
