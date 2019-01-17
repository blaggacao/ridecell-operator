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
	"fmt"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/aws/aws-sdk-go/service/iam/iamiface"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"

	awsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/aws/v1beta1"
)

type policyDocument struct {
	Version   string
	Statement []statementEntry
}

type statementEntry struct {
	Effect   string
	Action   []string
	Resource string
}

type iamUserComponent struct {
	iamAPI iamiface.IAMAPI
}

func NewS3Bucket() *iamUserComponent {
	sess := session.Must(session.NewSession())
	s3Service := iam.New(sess)
	return &iamUserComponent{iamAPI: s3Service}
}

func (comp *iamUserComponent) InjectIAMAPI(iamapi iamiface.IAMAPI) {
	comp.iamAPI = iamapi
}

func (_ *iamUserComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *iamUserComponent) IsReconcilable(_ *components.ComponentContext) bool {
	return true
}

func (comp *iamUserComponent) Reconcile(ctx *components.ComponentContext) (components.Result, error) {
	instance := ctx.Top.(*awsv1beta1.AwsResource)

	// Just about to add the policies bit here.

	iamUserInput := &iam.CreateUserInput{}
	user, err := comp.iamAPI.CreateUser(iamUserInput)
	if err != nil {
		if aerr, ok := err.(awserr.Error); ok {
			// If the error was not about the user already existing return the error
			if aerr.Code() != iam.ErrCodeEntityAlreadyExistsException {
				return components.Result{}, errors.Wrapf(aerr, "iam_user: failed to create iam user")
			}
		}
	}

	return components.Result{}, nil
}
