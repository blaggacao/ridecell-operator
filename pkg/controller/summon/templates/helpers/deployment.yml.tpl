{{ define "deployment" }}
apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
  namespace: {{ .Instance.Namespace }}
  labels:
    app.kubernetes.io/name: {{ block "componentName" . }}{{ end }}
    app.kubernetes.io/instance: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: {{ block "componentType" . }}{{ end }}
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: summon-operator
spec:
  replicas: {{ block "replicas" . }}1{{ end }}
  selector:
    matchLabels:
      app.kubernetes.io/instance: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
  template:
    metadata:
      labels:
        app.kubernetes.io/name: {{ block "componentName" . }}{{ end }}
        app.kubernetes.io/instance: {{ .Instance.Name }}-{{ block "componentName" . }}{{ end }}
        app.kubernetes.io/version: {{ .Instance.Spec.Version }}
        app.kubernetes.io/component: {{ block "componentType" . }}{{ end }}
        app.kubernetes.io/part-of: {{ .Instance.Name }}
        app.kubernetes.io/managed-by: summon-operator
    spec:
      imagePullSecrets:
      - name: pull-secret
      initContainers:
      - name: init
        image: us.gcr.io/ridecell-1/summon:{{ .Instance.Spec.Version }}
        imagePullPolicy: Always
        command:
        - sh
        - "-c"
        - |
          sed "s/xxPGPASSWORDxx/$(cat /postgres-credentials/password)/" </etc/secrets-orig/summon-platform.yml >/etc/secrets/summon-platform.yml
        resources:
          requests:
            memory: 8M
            cpu: 10m
          limits:
            memory: 16M
            cpu: 10m
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
        - name: secrets-orig
          mountPath: /etc/secrets-orig
        - name: secrets-shared
          mountPath: /etc/secrets
        - name: postgres-credentials
          mountPath: /postgres-credentials

      containers:
      - name: default
        image: us.gcr.io/ridecell-1/summon:{{ .Instance.Spec.Version }}
        imagePullPolicy: Always
        command: {{ block "command" . }}[]{{ end }}
        ports: {{ block "deploymentPorts" . }}[{containerPort: 8000}]{{ end }}
        resources:
          requests:
            memory: 512M
            cpu: 500m
          limits:
            memory: 1G
            cpu: 1000m
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
        - name: secrets-shared
          mountPath: /etc/secrets

      volumes:
        - name: config-volume
          configMap:
            name: {{ .Instance.Name }}-config
        - name: secrets-orig
          secret:
            secretName: {{ .Instance.Spec.Secret }}
        - name: secrets-shared
          emptyDir: {}
        - name: postgres-credentials
          secret:
            secretName: summon.{{ .Instance.Name }}-database.credentials
{{ end }}
