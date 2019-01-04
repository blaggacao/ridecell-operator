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
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"

	secretsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/secrets/v1beta1"
	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
)

var _ = Describe("SummonPlatform pull_secret Component", func() {

	It("creates a Pullsecret object", func() {
		comp := summoncomponents.NewPullSecret("pullsecret/pullsecret.yml.tpl")
		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
		target := &secretsv1beta1.PullSecret{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: instance.Name + "-pullsecret", Namespace: instance.Namespace}, target)
		Expect(err).ToNot(HaveOccurred())
	})
})
