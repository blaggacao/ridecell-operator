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
	"context"
	"encoding/json"
	"fmt"
	"time"

	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	iamusercomponents "github.com/Ridecell/ridecell-operator/pkg/controller/iamuser/components"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type mockIAMClient struct {
	iamiface.IAMAPI
	mockUserExists      bool
	mockhasUserPolicies bool
	mockExtraUserPolicy bool
}

var _ = Describe("iam_user aws Component", func() {

	It("runs basic reconcile with no existing user", func() {
		comp := iamusercomponents.NewIAMUser()
		instance.Spec.UserName = fmt.Sprintf("%s-k8s-summon-platform", instance.Name)
		mockIAM := &mockIAMClient{}
		comp.InjectIAMAPI(mockIAM)

		Expect(comp).To(ReconcileContext(ctx))

		fetchAccessKey := &corev1.Secret{}
		err := ctx.Client.Get(ctx.Context, types.NamespacedName{Name: fmt.Sprintf("%s-access-key", instance.Name), Namespace: instance.Namespace}, fetchAccessKey)
		Expect(err).ToNot(HaveOccurred())
		Expect(string(fetchAccessKey.Data["access_key_id"])).To(Equal("1234567890123456"))
		Expect(string(fetchAccessKey.Data["secret_access_key"])).To(Equal("FakeSecretKey00123"))
	})

	It("reconciles with existing user and credentials", func() {
		comp := iamusercomponents.NewIAMUser()
		instance.Spec.UserName = fmt.Sprintf("%s-k8s-summon-platform", instance.Name)
		mockIAM := &mockIAMClient{
			mockUserExists:      true,
			mockhasUserPolicies: true,
		}
		comp.InjectIAMAPI(mockIAM)

		accessKey := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-access-key", instance.Name), Namespace: instance.Namespace},
			Data: map[string][]byte{
				"access_key_id":     []byte("test_access_key"),
				"secret_access_key": []byte("test_secret_key"),
			},
		}
		ctx.Client = fake.NewFakeClient(accessKey)
		Expect(comp).To(ReconcileContext(ctx))

		fetchAccessKey := &corev1.Secret{}
		err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s-access-key", instance.Name), Namespace: instance.Namespace}, fetchAccessKey)
		Expect(err).ToNot(HaveOccurred())
		Expect(fetchAccessKey.Data).To(Equal(accessKey.Data))
	})

	It("has extra items attached to user", func() {
		comp := iamusercomponents.NewIAMUser()
		instance.Spec.UserName = fmt.Sprintf("%s-k8s-summon-platform", instance.Name)
		mockIAM := &mockIAMClient{
			mockUserExists:      true,
			mockExtraUserPolicy: true,
		}
		comp.InjectIAMAPI(mockIAM)
		Expect(comp).To(ReconcileContext(ctx))

		fetchAccessKey := &corev1.Secret{}
		err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s-access-key", instance.Name), Namespace: instance.Namespace}, fetchAccessKey)
		Expect(err).ToNot(HaveOccurred())
	})
})

// Mock aws functions below

func (m *mockIAMClient) GetUser(input *iam.GetUserInput) (*iam.GetUserOutput, error) {
	if aws.StringValue(input.UserName) != instance.Spec.UserName {
		return &iam.GetUserOutput{}, errors.New("awsmock_getuser: given username does not match spec")
	}
	if m.mockUserExists {
		return &iam.GetUserOutput{User: &iam.User{UserName: input.UserName}}, nil
	}
	return &iam.GetUserOutput{}, awserr.New(iam.ErrCodeNoSuchEntityException, "awsmock_getuser: user does not exist", errors.New(""))
}

func (m *mockIAMClient) CreateUser(input *iam.CreateUserInput) (*iam.CreateUserOutput, error) {
	if aws.StringValue(input.UserName) != instance.Spec.UserName {
		return &iam.CreateUserOutput{}, errors.New("awsmock_createuser: given username does not match spec")
	}
	return &iam.CreateUserOutput{User: &iam.User{UserName: input.UserName}}, nil
}

func (m *mockIAMClient) ListUserPolicies(input *iam.ListUserPoliciesInput) (*iam.ListUserPoliciesOutput, error) {
	if aws.StringValue(input.UserName) != instance.Spec.UserName {
		return &iam.ListUserPoliciesOutput{}, errors.New("awsmock_listuserpolicies: given username does not match spec")
	}
	if m.mockhasUserPolicies {
		inlinePoliciesPointers := []*string{}
		for k := range instance.Spec.InlinePolicies {
			inlinePoliciesPointers = append(inlinePoliciesPointers, aws.String(k))
		}
		return &iam.ListUserPoliciesOutput{PolicyNames: inlinePoliciesPointers}, nil
	}
	if m.mockExtraUserPolicy {
		inlinePoliciesPointers := []*string{}
		for k := range instance.Spec.InlinePolicies {
			inlinePoliciesPointers = append(inlinePoliciesPointers, aws.String(k))
		}
		inlinePoliciesPointers = append(inlinePoliciesPointers, aws.String("mock1"))
		return &iam.ListUserPoliciesOutput{PolicyNames: inlinePoliciesPointers}, nil
	}
	return &iam.ListUserPoliciesOutput{}, nil
}

func (m *mockIAMClient) GetUserPolicy(input *iam.GetUserPolicyInput) (*iam.GetUserPolicyOutput, error) {
	if aws.StringValue(input.UserName) != instance.Spec.UserName {
		return &iam.GetUserPolicyOutput{}, errors.New("awsmock_getuserpolicy: given username does not match spec")
	}
	if m.mockhasUserPolicies {
		inputPolicy := instance.Spec.InlinePolicies[aws.StringValue(input.PolicyName)]
		inputPolicyBytes, err := json.Marshal(inputPolicy)
		if err != nil {
			return &iam.GetUserPolicyOutput{}, errors.New("awsmock_getuserpolicy: unable to unmarshal json")
		}
		return &iam.GetUserPolicyOutput{PolicyDocument: aws.String(string(inputPolicyBytes))}, nil
	}
	if m.mockExtraUserPolicy {
		inputPolicy, ok := instance.Spec.InlinePolicies[aws.StringValue(input.PolicyName)]
		if !ok {
			inputPolicy = `{"Version": "2012-10-17", "Statement": {"Effect": "Allow", "Action": ["s3:*"] "Resource": "*"}}`
		}
		inputPolicyBytes, err := json.Marshal(inputPolicy)
		if err != nil {
			return &iam.GetUserPolicyOutput{}, errors.New("awsmock_getuserpolicy: unable to unmarshal json")
		}
		return &iam.GetUserPolicyOutput{PolicyDocument: aws.String(string(inputPolicyBytes))}, nil
	}
	return &iam.GetUserPolicyOutput{}, nil
}

func (m *mockIAMClient) DeleteUserPolicy(input *iam.DeleteUserPolicyInput) (*iam.DeleteUserPolicyOutput, error) {
	if aws.StringValue(input.UserName) != instance.Spec.UserName {
		return &iam.DeleteUserPolicyOutput{}, errors.New("awsmock_deleteuserpolicy: username did not match spec")
	}
	_, ok := instance.Spec.InlinePolicies[aws.StringValue(input.PolicyName)]
	if !ok {
		return &iam.DeleteUserPolicyOutput{}, nil
	}
	return &iam.DeleteUserPolicyOutput{}, errors.New("awsmock_deleteuserpolicy: policy shouldn't be getting deleted")
}

func (m *mockIAMClient) CreateAccessKey(input *iam.CreateAccessKeyInput) (*iam.CreateAccessKeyOutput, error) {
	if aws.StringValue(input.UserName) != instance.Spec.UserName {
		return &iam.CreateAccessKeyOutput{}, awserr.New(iam.ErrCodeNoSuchEntityException, "awsmock_createaccesskey: username did not match spec", errors.New(""))
	}
	curTime := time.Now()
	return &iam.CreateAccessKeyOutput{
		AccessKey: &iam.AccessKey{
			AccessKeyId:     aws.String("1234567890123456"),
			CreateDate:      &curTime,
			SecretAccessKey: aws.String("FakeSecretKey00123"),
			Status:          aws.String("Active"),
			UserName:        input.UserName,
		},
	}, nil
}
