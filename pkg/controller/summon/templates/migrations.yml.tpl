apiVersion: batch/v1
kind: Job
metadata:
  name: {{ .Instance.Name }}-migrations
  namespace: {{ .Instance.Namespace }}
  labels:
    app.kubernetes.io/name: migrations
    app.kubernetes.io/instance: {{ .Instance.Name }}-migrations
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: migration
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: summon-operator
spec:
  template:
    metadata:
      labels:
        app.kubernetes.io/name: migrations
        app.kubernetes.io/instance: {{ .Instance.Name }}-migrations
        app.kubernetes.io/version: {{ .Instance.Spec.Version }}
        app.kubernetes.io/component: migration
        app.kubernetes.io/part-of: {{ .Instance.Name }}
        app.kubernetes.io/managed-by: summon-operator
    spec:
      restartPolicy: Never
      imagePullSecrets:
      - name: pull-secret
      containers:
      - name: default
        image: us.gcr.io/ridecell-1/summon:{{ .Instance.Spec.Version }}
        imagePullPolicy: Always
        command:
        - sh
        - "-c"
        - python manage.py migrate
        resources:
          requests:
            memory: 512M
            cpu: 200m
          limits:
            memory: 1G
            cpu: 500m
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
        - name: secrets-orig
          mountPath: /etc/secrets-orig
        - name: app-secrets
          mountPath: /etc/secrets
        - name: postgres-credentials
          mountPath: /postgres-credentials

      volumes:
        - name: config-volume
          configMap:
            name: {{ .Instance.Name }}-config
        - name: secrets-orig
          secret:
            secretName: {{ .Instance.Spec.Secret }}
        - name: app-secrets
          secret:
            secretName: summon.{{ .Instance.Name }}.app-secrets
        - name: postgres-credentials
          secret:
            secretName: summon.{{ .Instance.Name }}-database.credentials
