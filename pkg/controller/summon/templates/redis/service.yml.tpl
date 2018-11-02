kind: Service
apiVersion: v1
metadata:
  name: {{ .Instance.Name }}-redis
  namespace: {{ .Instance.Namespace }}
  labels:
    app.kubernetes.io/name: redis
    app.kubernetes.io/instance: {{ .Instance.Name }}-redis
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: database
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: summon-operator
spec:
  selector:
    app.kubernetes.io/instance: {{ .Instance.Name }}-redis
  ports:
  - protocol: TCP
    port: 6379
