{{ define "componentName" }}static{{ end }}
{{ define "componentType" }}web{{ end }}
{{ define "command" }}[caddy, "-port", "8000", "-root", /var/www, "-log", stdout]{{ end }}
{{ define "replicas" }}{{ .Instance.Spec.StaticReplicas }}{{ end }}
{{ template "deployment" . }}
