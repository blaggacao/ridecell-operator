apiVersion: v1
kind: ConfigMap
metadata:
  name: {{ .Instance.Name }}-config
  labels:
    app.kubernetes.io/name: config
    app.kubernetes.io/instance: {{ .Instance.Name }}-config
    app.kubernetes.io/version: {{ .Instance.Spec.Version }}
    app.kubernetes.io/component: config
    app.kubernetes.io/part-of: {{ .Instance.Name }}
    app.kubernetes.io/managed-by: summon-operator
data:
  summon-platform.yml: |
    AMAZON_S3_USED: False
    ASGI_URL: 'redis://{{ .Instance.Name }}-redis/0'
    AWS_REGION: 'us-west-2'
    AWS_STORAGE_BUCKET_NAME: ''
    CACHE_URL: 'redis://{{ .Instance.Name }}-redis/1'
    CARSHARING_V1_API_DISABLED: False
    CLOUDFRONT_DISTRIBUTION: ''
    COMPRESS_ENABLED: False
    CSBE_CONNECTION_USED: False
    DATA_PIPELINE_SQS_QUEUE_NAME: 'master-data-pipeline'
    DEBUG: True
    ENABLE_NEW_RELIC: False
    ENABLE_SENTRY: False
    FACEBOOK_AUTHENTICATION_EMPLOYEE_PERMISSION_REQUIRED: False
    FIREBASE_APP: 'instant-stage'
    FIREBASE_ROOT_NODE: 'unknown-local'
    GDPR_ENABLED: True
    GOOGLE_ANALYTICS_ID: 'UA-37653074-1'
    INTERNATIONAL_OUTGOING_SMS_NUMBER: '14152345773'
    NEWRELIC_NAME: ''
    OAUTH_HOSTED_DOMAIN: ''
    OUTGOING_SMS_NUMBER: '41254'
    PLATFORM_ENV: 'DEV'
    SAML_EMAIL_ATTRIBUTE: 'eduPersonPrincipalName'
    SAML_FIRST_NAME_ATTRIBUTE: 'givenName'
    SAML_IDP_ENTITY_ID: 'https://idp.testshib.org/idp/shibboleth'
    SAML_IDP_METADATA_FILENAME: ''
    SAML_IDP_METADATA_URL: 'https://www.testshib.org/metadata/testshib-providers.xml'
    SAML_IDP_PUBLIC_KEY_FILENAME: 'testshib.crt'
    SAML_IDP_SSO_URL: 'https://idp.testshib.org/idp/profile/SAML2/Redirect/SSO'
    SAML_LAST_NAME_ATTRIBUTE: 'sn'
    SAML_NAME_ID_FORMAT: 'urn:oasis:names:tc:SAML:2.0:nameid-format:transient'
    SAML_PRIVATE_KEY_FILENAME: 'sp.key'
    SAML_PRIVATE_KEY_FILENAME: 'sp.key'
    SAML_PUBLIC_KEY_FILENAME: 'sp.crt'
    SAML_PUBLIC_KEY_FILENAME: 'sp.crt'
    SAML_SERVICE_NAME: 'RideCell SAML Test'
    SAML_USE_LOCAL_METADATA: ''
    SAML_VALID_FOR_HOURS: 24
    SESSION_COOKIE_AGE: 1209600
    TENANT_ID: 'unknown-local'
    TIME_ZONE: 'America/Los_Angeles'
    USE_FACEBOOK_AUTHENTICATION_FOR_RIDERS: False
    USE_GOOGLE_AUTHENTICATION_FOR_RIDERS: False
    USE_SAML_AUTHENTICATION_FOR_RIDERS: False
    WEB_URL: 'http://localhost:38080'
    XMLSEC_BINARY_LOCATION: '/usr/bin/xmlsec1'
