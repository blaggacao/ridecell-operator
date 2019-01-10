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
	"fmt"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("app_secrets Component", func() {

	It("Runs a reconcile while secret_key exists ", func() {
		comp := summoncomponents.NewSecretKey()

		newKey := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s.secret-key", instance.Name), Namespace: instance.Namespace},
			Data:       map[string][]byte{"SECRET_KEY": []byte("testing")},
		}

		ctx.Client = fake.NewFakeClient(newKey)
		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())

		fetchKey := &corev1.Secret{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s.secret-key", instance.Name), Namespace: instance.Namespace}, fetchKey)
		Expect(err).ToNot(HaveOccurred())
		val, ok := fetchKey.Data["SECRET_KEY"]
		Expect(ok).To(Equal(true))
		Expect(val).To(Equal([]byte("testing")))
	})

	It("Runs reconcile with no secret", func() {
		comp := summoncomponents.NewSecretKey()

		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())

		fetchKey := &corev1.Secret{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s.secret-key", instance.Name), Namespace: instance.Namespace}, fetchKey)
		Expect(err).ToNot(HaveOccurred())
		val, ok := fetchKey.Data["SECRET_KEY"]
		Expect(ok).To(Equal(true))
		Expect(val).To(HaveLen(86))
	})

	It("Runs reconcile with invalid secret", func() {
		comp := summoncomponents.NewSecretKey()

		newKey := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s.secret-key", instance.Name), Namespace: instance.Namespace},
			Data:       map[string][]byte{"foo": []byte("bar")},
		}

		ctx.Client = fake.NewFakeClient(newKey)
		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
		fetchKey := &corev1.Secret{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: fmt.Sprintf("%s.secret-key", instance.Name), Namespace: instance.Namespace}, fetchKey)
		Expect(err).ToNot(HaveOccurred())

		_, ok := fetchKey.Data["foo"]
		Expect(ok).To(Equal(false))

		val, ok := fetchKey.Data["SECRET_KEY"]
		Expect(ok).To(Equal(true))
		Expect(val).To(HaveLen(86))
	})
})
