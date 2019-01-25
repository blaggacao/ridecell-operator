kind: S3Bucket
apiVersion: aws.ridecell.io/v1beta1
metadata:
 name: {{ .Instance.Name }}-s3bucket
 namespace: {{ .Instance.Namespace }}
spec:
 bucketName: {{ .Extra.bucketName }}
 region: {{ .Instance.AwsRegion }}
 bucketPolicy: |
               {
                 "Version": "2008-10-17",
                 "Statement": [{
                    "Sid": "PublicReadForGetBucketObjects",
                    "Effect": "Allow",
                    "Principal": {
                      "AWS": "*"
                    },
                    "Action": "s3:GetObject",
                    "Resource": "arn:aws:s3:::{{ .Extra.bucketName }}/*"
                  },],
               }
