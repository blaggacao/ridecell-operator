kind: PostgresOperatorDatabase
apiVersion: db.ridecell.io/v1beta1
metadata:
 name: {{ .Instance.Name }}
 namespace: {{ .Instance.Namespace }}
spec:
  databaseRef:
    name: {{ .Instance.Spec.DatabaseSpec.SharedDatabaseName }}
  database: {{ .Instance.Name }}
