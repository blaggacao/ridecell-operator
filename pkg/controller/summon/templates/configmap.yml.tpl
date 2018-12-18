apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Instance.Name }}-config
  namespace: {{ .Instance.Namespace }}
  labels:
    app.kubernetes.io/name: config
    app.kubernetes.io/instance: {{ .Instance.Name }}-config
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: config
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: summon-operator
data:
  summon-platform.yml: |
    {{ .Extra.SummonYaml }}
