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

package components_test

import (
	"encoding/json"

	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/pkg/errors"

	s3bucketcomponents "github.com/Ridecell/ridecell-operator/pkg/controller/s3bucket/components"
)

type mockS3Client struct {
	s3iface.S3API
	mockBucketExists     bool
	mockBucketHasPolicy  bool
	mockEditBucketPolicy bool
	mockBucketNameTaken  bool
}

var _ = Describe("s3bucket aws Component", func() {

	It("runs basic reconcile with no existing bucket", func() {
		comp := s3bucketcomponents.NewS3Bucket()
		mockS3 := &mockS3Client{}
		comp.InjectS3API(mockS3)

		instance.Spec.BucketName = "foo-default-static"
		Expect(comp).To(ReconcileContext(ctx))
	})

	It("reconciles with existing bucket policy", func() {
		comp := s3bucketcomponents.NewS3Bucket()
		mockS3 := &mockS3Client{
			mockBucketExists:    true,
			mockBucketHasPolicy: true,
		}
		comp.InjectS3API(mockS3)

		instance.Spec.BucketName = "foo-default-static"
		instance.Spec.BucketPolicy = `{
			"Version": "2008-10-17",
			"Statement": [{
				 "Sid": "PublicReadForGetBucketObjects",
				 "Effect": "Allow",
				 "Principal": {
					 "AWS": "*"
				 },
				 "Action": "s3:GetObject",
				 "Resource": "arn:aws:s3:::foo-default-static/*"
			 }]
		}`
		Expect(comp).To(ReconcileContext(ctx))
	})

	It("edit existing bucket policy", func() {
		comp := s3bucketcomponents.NewS3Bucket()
		mockS3 := &mockS3Client{
			mockBucketExists:     true,
			mockEditBucketPolicy: true,
		}
		comp.InjectS3API(mockS3)

		instance.Spec.BucketName = "foo-default-static"
		instance.Spec.BucketPolicy = `{
			"Version": "2008-10-17",
			"Statement": [{
				 "Sid": "PublicReadForGetBucketObjects",
				 "Effect": "Allow",
				 "Principal": {
					 "AWS": "*"
				 },
				 "Action": "s3:GetObject",
				 "Resource": "arn:aws:s3:::foo-default-static/*"
			 }]
		}`
		Expect(comp).To(ReconcileContext(ctx))
	})

	It("fails because bucket name is taken", func() {
		comp := s3bucketcomponents.NewS3Bucket()
		mockS3 := &mockS3Client{
			mockBucketNameTaken: true,
		}
		comp.InjectS3API(mockS3)

		instance.Spec.BucketName = "foo-default-static"

		Expect(comp).ToNot(ReconcileContext(ctx))
	})
})

// Mock aws functions below

func (m *mockS3Client) ListBuckets(input *s3.ListBucketsInput) (*s3.ListBucketsOutput, error) {
	if m.mockBucketExists {
		return &s3.ListBucketsOutput{
			Buckets: []*s3.Bucket{
				&s3.Bucket{Name: aws.String(instance.Spec.BucketName)},
			}}, nil
	} else {
		return &s3.ListBucketsOutput{}, nil
	}
}

func (m *mockS3Client) CreateBucket(input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	if aws.StringValue(input.Bucket) != instance.Spec.BucketName {
		return &s3.CreateBucketOutput{}, errors.New("awsmock_createbucket: bucket name was incorrect")
	}
	if aws.StringValue(input.CreateBucketConfiguration.LocationConstraint) != instance.Spec.Region {
		return &s3.CreateBucketOutput{}, errors.New("awsmock_createbucket: region was incorrect")
	}
	if m.mockBucketNameTaken {
		return &s3.CreateBucketOutput{}, errors.New("awsmock_createbucket: bucket name taken")
	}
	return &s3.CreateBucketOutput{}, nil
}

func (m *mockS3Client) GetBucketPolicy(input *s3.GetBucketPolicyInput) (*s3.GetBucketPolicyOutput, error) {
	if aws.StringValue(input.Bucket) != instance.Spec.BucketName {
		return &s3.GetBucketPolicyOutput{}, errors.New("awsmock_getbucketpolicy: bucketname was incorrect")
	}
	if m.mockBucketHasPolicy {
		inputBucketPolicyBytes, err := json.Marshal(instance.Spec.BucketPolicy)
		if err != nil {
			return &s3.GetBucketPolicyOutput{}, errors.New("awsmock_getbucketpolicy: unable to marshal input policy")
		}
		return &s3.GetBucketPolicyOutput{Policy: aws.String(string(inputBucketPolicyBytes))}, nil
	}
	return &s3.GetBucketPolicyOutput{}, nil
}

func (m *mockS3Client) PutBucketPolicy(input *s3.PutBucketPolicyInput) (*s3.PutBucketPolicyOutput, error) {
	// Get our policies into a comparable data type
	givenBucketPolicyBytes, err := json.Marshal(aws.StringValue(input.Policy))
	if err != nil {
		return &s3.PutBucketPolicyOutput{}, errors.New("awsmock_putbucketpolicy: unable to marshal given policy")
	}
	inputBucketPolicyBytes, err := json.Marshal(instance.Spec.BucketPolicy)
	if err != nil {
		return &s3.PutBucketPolicyOutput{}, errors.New("awsmock_putbucketpolicy: unable to marshal input policy")
	}
	// start mock bits
	if aws.StringValue(input.Bucket) != instance.Spec.BucketName {
		return &s3.PutBucketPolicyOutput{}, errors.New("awsmock_putbucketpolicy: bucket name was incorrect")
	}
	if string(givenBucketPolicyBytes) != string(inputBucketPolicyBytes) && !m.mockEditBucketPolicy {
		return &s3.PutBucketPolicyOutput{}, errors.New("awsmock_putbucketpolicy: given bucket policy did not match input")
	}
	return &s3.PutBucketPolicyOutput{}, nil
}
