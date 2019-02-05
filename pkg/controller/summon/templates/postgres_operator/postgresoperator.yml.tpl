apiVersion: db.ridecell.io/v1beta1
kind: PostgresOperatorDatabase
metadata:
 name: {{ .Instance.Name }}
 namespace: {{ .Instance.Namespace }}
spec:
  databaseRef:
    name: {{ .Instance.Spec.DatabaseSpec.SharedDatabaseName }}
  database: {{ .Instance.Name }}
