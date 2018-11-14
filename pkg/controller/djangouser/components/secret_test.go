/*
Copyright 2018 Ridecell, Inc..

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	djangousercomponents "github.com/Ridecell/ridecell-operator/pkg/controller/djangouser/components"
)

var _ = Describe("DjangoUser Secret Component", func() {
	It("creates a password if no secret exists", func() {
		instance := &summonv1beta1.DjangoUser{
			ObjectMeta: metav1.ObjectMeta{Name: "foo.example.com", Namespace: "default"},
			Spec: summonv1beta1.DjangoUserSpec{
				PasswordSecret: "foo-credentials",
				Database: &summonv1beta1.DatabaseConnection{
					PasswordSecretRef: &summonv1beta1.SecretRef{},
				},
			},
		}
		ctx := &components.ComponentContext{Top: instance, Client: fake.NewFakeClient(), Scheme: scheme.Scheme}

		comp := djangousercomponents.NewSecret()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		secret := &corev1.Secret{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-credentials", Namespace: "default"}, secret)
		Expect(err).NotTo(HaveOccurred())
		password, ok := secret.Data["password"]
		Expect(ok).To(BeTrue())
		Expect(len(password)).To(Equal(22))
	})

	It("creates a password if the secret exists but has no password key", func() {
		instance := &summonv1beta1.DjangoUser{
			ObjectMeta: metav1.ObjectMeta{Name: "foo.example.com", Namespace: "default"},
			Spec: summonv1beta1.DjangoUserSpec{
				PasswordSecret: "foo-credentials",
				Database: &summonv1beta1.DatabaseConnection{
					PasswordSecretRef: &summonv1beta1.SecretRef{},
				},
			},
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "foo-credentials", Namespace: "default"},
		}
		ctx := &components.ComponentContext{Top: instance, Client: fake.NewFakeClient(secret), Scheme: scheme.Scheme}

		comp := djangousercomponents.NewSecret()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		newSecret := &corev1.Secret{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-credentials", Namespace: "default"}, newSecret)
		Expect(err).NotTo(HaveOccurred())
		password, ok := newSecret.Data["password"]
		Expect(ok).To(BeTrue())
		Expect(len(password)).To(Equal(22))
	})

	It("does not change an existing password", func() {
		instance := &summonv1beta1.DjangoUser{
			ObjectMeta: metav1.ObjectMeta{Name: "foo.example.com", Namespace: "default"},
			Spec: summonv1beta1.DjangoUserSpec{
				PasswordSecret: "foo-credentials",
				Database: &summonv1beta1.DatabaseConnection{
					PasswordSecretRef: &summonv1beta1.SecretRef{},
				},
			},
		}
		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "foo-credentials", Namespace: "default"},
			Data: map[string][]byte{
				"password": []byte("foo"),
			},
		}
		ctx := &components.ComponentContext{Top: instance, Client: fake.NewFakeClient(secret), Scheme: scheme.Scheme}

		comp := djangousercomponents.NewSecret()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		newSecret := &corev1.Secret{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-credentials", Namespace: "default"}, newSecret)
		Expect(err).NotTo(HaveOccurred())
		password, ok := newSecret.Data["password"]
		Expect(ok).To(BeTrue())
		Expect(password).To(Equal([]byte("foo")))
	})
})
