apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: {{ .Instance.Name }}-celerybeat
  labels:
    app.kubernetes.io/name: celerybeat
    app.kubernetes.io/instance: {{ .Instance.Name }}-celerybeat
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: worker
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: summon-operator
spec:
  selector:
    matchLabels:
      app.kubernetes.io/instance: {{ .Instance.Name }}-celerybeat
  serviceName: {{ .Instance.Name }}-celerybeat
  template:
    metadata:
      labels:
        app.kubernetes.io/name: celerybeat
        app.kubernetes.io/instance: {{ .Instance.Name }}-celerybeat
        app.kubernetes.io/version: {{ .Instance.Spec.Version }}
        app.kubernetes.io/component: worker
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
            memory: 4M
            cpu: 10m
          limits:
            memory: 8M
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
      - name: volumeperms
        image: alpine:latest
        command: [chown, "1000:1000", /schedule]
        resources:
          requests:
            memory: 4M
            cpu: 10m
          limits:
            memory: 8M
            cpu: 10m
        volumeMounts:
        - name: beat-state
          mountPath: /schedule

      containers:
      - name: default
        image: us.gcr.io/ridecell-1/summon:{{ .Instance.Spec.Version }}
        imagePullPolicy: Always
        command: [python, "-m", celery, "-A", summon_platform, beat, "-l", info, "--schedule", /schedule/beat, --pidfile=]
        resources:
          requests:
            memory: 512M
            cpu: 100m
          limits:
            memory: 1G
            cpu: 200m
        volumeMounts:
        - name: config-volume
          mountPath: /etc/config
        - name: secrets-shared
          mountPath: /etc/secrets
        - name: beat-state
          mountPath: /schedule

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
  volumeClaimTemplates:
  - metadata:
      name: beat-state
    spec:
      accessModes: [ReadWriteOnce]
      resources:
        requests:
          storage: 1Gi # This only actually needs about 1Mb
