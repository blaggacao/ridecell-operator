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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

    secretscomponents "github.com/Ridecell/ridecell-operator/pkg/controller/secrets/components"
    secretsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/secrets/v1beta1"
)

var _ = Describe("pull_secret Component", func() {

	It("Runs reconcile with no value set", func() {
	    instance := ctx.Top.(*secretsv1beta1.PullSecret)
	    comp := secretscomponents.NewSecret()
	    _, err := comp.Reconcile(ctx)
	    Expect(err).To(HaveOccurred())
        Expect(instance.Status.Status).To(Equal(secretsv1beta1.StatusErrorSecretNotFound))
	})

})
