apiVersion: secrets.ridecell.io/v1beta1
kind: PullSecret
metadata:
  name: {{ .Instance.Name }}-pullsecret
  namespace: {{ .Instance.Namespace }}
spec:
  pullSecretName: {{ .Instance.Spec.PullSecret }}
