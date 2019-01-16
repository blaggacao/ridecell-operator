{{ define "componentName" }}web{{ end }}
{{ define "componentType" }}web{{ end }}
{{ define "command" }}[python, -m, twisted, web, --listen, tcp:8000, --wsgi, summon_platform.wsgi]{{ end }}
{{ define "replicas" }}{{ .Instance.Spec.WebReplicas }}{{ end }}
{{ template "deployment" . }}
