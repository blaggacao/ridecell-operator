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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"

	awsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/aws/v1beta1"
	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
)

var _ = Describe("SummonPlatform s3bucket Component", func() {

	It("creates an S3Bucket object", func() {
		comp := summoncomponents.NewS3Bucket("aws/s3bucket.yml.tpl")
		Expect(comp).To(ReconcileContext(ctx))
		target := &awsv1beta1.S3Bucket{}
		err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, target)
		Expect(err).ToNot(HaveOccurred())
	})
})
