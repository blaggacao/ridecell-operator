{{ define "componentName" }}channelworker{{ end }}
{{ define "componentType" }}worker{{ end }}
{{ define "command" }}[python, manage.py, runworker, "-v2", "--threads", "2"]{{ end }}
{{ define "replicas" }}{{ .Instance.Spec.ChannelWorkerReplicas }}{{ end }}
{{ template "deployment" . }}
