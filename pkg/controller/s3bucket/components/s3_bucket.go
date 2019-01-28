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
	"encoding/json"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/aws/aws-sdk-go/aws"
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
	instance := ctx.Top.(*awsv1beta1.S3Bucket)

	// Try to find our specified buckeet
	listBucketsOutput, err := comp.s3API.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return components.Result{}, errors.Wrapf(err, "s3_bucket: failed to list s3 buckets")
	}
	var foundBucket bool
	for _, bucket := range listBucketsOutput.Buckets {
		if aws.StringValue(bucket.Name) == instance.Spec.BucketName {
			foundBucket = true
		}
	}

	// If the bucket does not exist create it
	if !foundBucket {
		_, err = comp.s3API.CreateBucket(&s3.CreateBucketInput{
			Bucket: aws.String(instance.Spec.BucketName),
			CreateBucketConfiguration: &s3.CreateBucketConfiguration{
				LocationConstraint: aws.String(instance.Spec.Region),
			},
		})
		if err != nil {
			return components.Result{}, errors.Wrapf(err, "s3_bucket: failed to create s3 bucket")
		}
	}

	getBucketPolicyObj, err := comp.s3API.GetBucketPolicy(&s3.GetBucketPolicyInput{Bucket: aws.String(instance.Spec.BucketName)})
	if err != nil {
		return components.Result{}, errors.Wrapf(err, "s3_bucket: failed to get bucket policy")
	}

	// Make sure our bucket policy is sorted by marhsaling it
	getBucketPolicyBytes, err := json.Marshal(aws.StringValue(getBucketPolicyObj.Policy))
	if err != nil {
		return components.Result{}, errors.Wrapf(err, "s3_bucket: failed to marshal bucket policy")
	}

	// Make sure our input bucket policy is sorted by marshaling it
	inputBucketPolicyBytes, err := json.Marshal(instance.Spec.BucketPolicy)
	if err != nil {
		return components.Result{}, errors.Wrapf(err, "s3_bucket: failed to marshal input bucket policy")
	}

	// if the bucket policies do not match create or update it
	if string(getBucketPolicyBytes) != string(inputBucketPolicyBytes) {
		_, err := comp.s3API.PutBucketPolicy(&s3.PutBucketPolicyInput{
			Bucket: aws.String(instance.Spec.BucketName),
			Policy: aws.String(instance.Spec.BucketPolicy),
		})
		if err != nil {
			return components.Result{}, errors.Wrapf(err, "s3_bucket: failed to create or update bucket policy")
		}
	}

	return components.Result{StatusModifier: func(obj runtime.Object) error {
		instance := obj.(*awsv1beta1.S3Bucket)
		instance.Status.Status = awsv1beta1.StatusReady
		instance.Status.Message = "Bucket exists and has correct policy"
		return nil
	}}, nil
}
