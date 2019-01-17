/*
Copyright 2019 Ridecell, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package components

import (
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"

	awsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/aws/v1beta1"
)

type s3BucketComponent struct {
	s3API s3iface.S3API
}

func NewS3Bucket() *s3BucketComponent {
	sess := session.Must(session.NewSession())
	s3Service := s3.New(sess)
	return &s3BucketComponent{s3API: s3Service}
}

func (comp *s3BucketComponent) InjectS3API(s3api s3iface.S3API) {
	comp.s3API = s3api
}

func (_ *s3BucketComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *s3BucketComponent) IsReconcilable(_ *components.ComponentContext) bool {
	return true
}

func (comp *s3BucketComponent) Reconcile(ctx *components.ComponentContext) (components.Result, error) {
	instance := ctx.Top.(*awsv1beta1.AwsResource)

	bucketInput := &s3.CreateBucketInput{
		Bucket: aws.String(instance.Spec.BucketName),
		CreateBucketConfiguration: &s3.CreateBucketConfiguration{
			LocationConstraint: aws.String(instance.Spec.Region),
		},
	}
	_, err := comp.s3API.CreateBucket(bucketInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			// If the error was not about the bucket already existing return the error
			if aerr.Code() != s3.ErrCodeBucketAlreadyExists && aerr.Code() != s3.ErrCodeBucketAlreadyOwnedByYou {
				return components.Result{}, errors.Wrapf(aerr, "s3_bucket: failed to create s3 bucket")
			}
		}
	}

	return components.Result{}, nil
}
