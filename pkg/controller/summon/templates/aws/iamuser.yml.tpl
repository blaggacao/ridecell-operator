kind: IAMUser
apiVersion: aws.ridecell.io/v1beta1
metadata:
 name: {{ .Instance.Name }}-iamuser
 namespace: {{ .Instance.Namespace }}
spec:
 username: {{ .Instance.Name }}-summon-platform
 inlinePolicies:
   allow_s3: |
            {
               "Version": "2012-10-17",
               "Statement": {
                 "Effect": "Allow",
                 "Action": [
                    "s3:ListBucket",
                    "s3:GetObject",
                    "s3:DeleteObject",
                    "s3:PutObject"
                  ],
                 "Resource": "arn:aws:s3:::ridecell-{{ .Instance.Name }}-static*"
               }
            }
   allow_sqs: |
            {
              "Version": "2012-10-17",
              "Statement": {
                "Sid": "",
                "Effect": "Allow",
                "Action": [
                  "sqs:SendMessageBatch",
                  "sqs:SendMessage",
                  "sqs:CreateQueue"
                ],
                "Resource": "arn:aws:sqs:{{ .Instance.Spec.AwsRegion }}::{{ .Instance.Spec.SQSQueue }}"
              }
            }
