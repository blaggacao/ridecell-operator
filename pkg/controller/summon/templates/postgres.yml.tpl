apiVersion: acid.zalan.do/v1
kind: postgresql
metadata:
  name: {{ .Instance.Name }}-database
  namespace: {{ .Instance.Namespace }}
  labels:
    app.kubernetes.io/name: database
    app.kubernetes.io/instance: {{ .Instance.Name }}-database
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: database
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: summon-operator
spec:
  teamId: {{ .Instance.Name }}
  volume:
    size: 10Gi
  numberOfInstances: 1
  users:
    ridecell-admin: [superuser]
    summon: [superuser]
    reporting: []
    periscope: []
  databases:
    summon: ridecell-admin
  postgresql:
    version: "10"
