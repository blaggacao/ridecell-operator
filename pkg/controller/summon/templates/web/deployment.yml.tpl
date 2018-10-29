{{ define "componentName" }}web{{ end }}
{{ define "componentType" }}web{{ end }}
{{ define "command" }}[python, "-m", gunicorn.app.wsgiapp, "-b", "0.0.0.0:8000", summon_platform.wsgi]{{ end }}
{{ define "replicas" }}{{ .Instance.Spec.WebReplicas }}{{ end }}
{{ template "deployment" . }}
