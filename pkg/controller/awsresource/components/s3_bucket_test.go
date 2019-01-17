/*
Copyright 2018 Ridecell, Inc.

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
	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3iface"
	"github.com/pkg/errors"
	//"sigs.k8s.io/controller-runtime/pkg/client/fake"
	//awsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/aws/v1beta1"
	awsresourcecomponents "github.com/Ridecell/ridecell-operator/pkg/controller/awsresource/components"
	//corev1 "k8s.io/api/core/v1"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockS3Client struct {
	s3iface.S3API
}

var _ = Describe("awsresource Database Component", func() {

	It("runs basic reconcile", func() {
		comp := awsresourcecomponents.NewS3Bucket()
		mockS3 := &mockS3Client{}
		comp.InjectS3API(mockS3)
		Expect(comp).To(ReconcileContext(ctx))

	})

	It("reconciles with existing bucket", func() {
		comp := awsresourcecomponents.NewS3Bucket()
		mockS3 := &mockS3Client{}
		comp.InjectS3API(mockS3)

		// Set BucketName to mock BucketAlreadyExists error
		instance.Spec.BucketName = "already_exists"
		Expect(comp).To(ReconcileContext(ctx))
	})
})

func (m *mockS3Client) CreateBucket(input *s3.CreateBucketInput) (*s3.CreateBucketOutput, error) {
	if aws.StringValue(input.Bucket) != instance.Spec.BucketName {
		return &s3.CreateBucketOutput{}, errors.New("mock_createbucket: bucket name was incorrect")
	} else if aws.StringValue(input.CreateBucketConfiguration.LocationConstraint) != instance.Spec.Region {
		return &s3.CreateBucketOutput{}, errors.New("mock_createbucket: region was incorrect")
		// Hacky way to simulate bucket already existing.
	} else if aws.StringValue(input.Bucket) == "already_exists" {
		return &s3.CreateBucketOutput{}, awserr.New(s3.ErrCodeBucketAlreadyOwnedByYou, "mocks3", nil)
	}
	return &s3.CreateBucketOutput{}, nil
}
