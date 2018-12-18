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

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

    secretscomponents "github.com/Ridecell/ridecell-operator/pkg/controller/secrets/components"
    secretsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/secrets/v1beta1"

)

var _ = Describe("pull_secret Component", func() {

	It("Runs reconcile with no value set", func() {
	    comp := secretscomponents.NewSecret()
	    _, err := comp.Reconcile(ctx)
	    Expect(err).To(HaveOccurred())
        Expect(instance.Status.Status).To(Equal(secretsv1beta1.StatusErrorSecretNotFound))
	})

	It("Sets valid secret, runs reconcile", func() {
	    comp := secretscomponents.NewSecret()
	    newPullSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "secrets.ridecell.us", Namespace: "default"},
			Data: map[string][]byte{
				".dockerconfigjson": []byte("dslakfjlskdj3"),
			},
		}
		ctx.Client = fake.NewFakeClient(newPullSecret)
	    _, err := comp.Reconcile(ctx)
	    Expect(err).ToNot(HaveOccurred())
        Expect(instance.Status.Status).To(Equal(secretsv1beta1.StatusReady))
	})
})

