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
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	secretsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/secrets/v1beta1"
	secretscomponents "github.com/Ridecell/ridecell-operator/pkg/controller/secrets/components"
)

var _ = Describe("pull_secret Component", func() {

	BeforeEach(func() {
		os.Setenv("NAMESPACE", instance.ObjectMeta.Namespace)
	})

	It("Runs reconcile with no value set", func() {
		comp := secretscomponents.NewSecret()
		_, err := comp.Reconcile(ctx)
		Expect(err).To(HaveOccurred())
		Expect(instance.Status.Status).To(Equal(secretsv1beta1.StatusErrorSecretNotFound))
	})

	It("Sets valid secret, runs reconcile", func() {
		comp := secretscomponents.NewSecret()
		newPullSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: instance.Spec.PullSecret, Namespace: "default"},
			Data: map[string][]byte{
				".dockerconfigjson": []byte("dslakfjlskdj3"),
			},
		}
		ctx.Client = fake.NewFakeClient(newPullSecret)
		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(instance.Status.Status).To(Equal(secretsv1beta1.StatusReady))

	})

	It("Ensures details remain the same after reconcile", func() {
		newPullSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: instance.Spec.PullSecret, Namespace: "default", Labels: map[string]string{"Name": "stuff"}},
			Data: map[string][]byte{
				".dockerconfigjson": []byte("dslakfjlskdj3"),
			},
		}
		ctx.Client = fake.NewFakeClient(newPullSecret)
		comp := secretscomponents.NewSecret()
		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
		target := &corev1.Secret{}
		err = ctx.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.PullSecret, Namespace: "default"}, target)
		Expect(err).ToNot(HaveOccurred())
		Expect(target.ObjectMeta.Labels).To(Equal(map[string]string{"Name": "stuff"}))
		Expect(target.Data).To(Equal(map[string][]byte{".dockerconfigjson": []byte("dslakfjlskdj3")}))
	})
})
