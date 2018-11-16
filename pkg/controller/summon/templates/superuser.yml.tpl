apiVersion: summon.ridecell.io/v1beta1
kind: DjangoUser
metadata:
  name: dispatcher.ridecell.com
  namespace: {{ .Instance.Namespace }}
spec:
  active: true
  staff: true
  superuser: true
  database:
    host: {{ .Instance.Name }}-database.{{ .Instance.Namespace }}
    username: summon
    database: summon
    passwordSecretRef:
      name: summon.{{ .Instance.Name }}-database.credentials
