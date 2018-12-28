apiVersion: db.ridecell.io/v1beta1
kind: PostgresExtension
metadata:
  name: {{ .Instance.Name }}-postgis
  namespace: {{ .Instance.Namespace }}
spec:
  extensionName: postgis
  database:
    host: {{ .Instance.Name }}-database.{{ .Instance.Namespace }}
    username: ridecell-admin
    database: summon
    passwordSecretRef:
      name: ridecell-admin.{{ .Instance.Name }}-database.credentials
