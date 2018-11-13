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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	ducomponents "github.com/Ridecell/ridecell-operator/pkg/controller/djangouser/components"
)

var _ = Describe("DjangoUser Defaults Component", func() {
	It("does nothing on a filled out object", func() {
		instance := &summonv1beta1.DjangoUser{
			ObjectMeta: metav1.ObjectMeta{Name: "foo.example.com"},
			Spec: summonv1beta1.DjangoUserSpec{
				Username:       "foo@bar.com",
				PasswordSecret: "foo-credentials",
			},
		}
		ctx := &components.ComponentContext{Top: instance}

		comp := ducomponents.NewDefaults()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Spec.Username).To(Equal("foo@bar.com"))
		Expect(instance.Spec.PasswordSecret).To(Equal("foo-credentials"))
	})

	It("sets a default password secret", func() {
		instance := &summonv1beta1.DjangoUser{
			ObjectMeta: metav1.ObjectMeta{Name: "foo.example.com"},
			Spec: summonv1beta1.DjangoUserSpec{
				Username: "foo@bar.com",
			},
		}
		ctx := &components.ComponentContext{Top: instance}

		comp := ducomponents.NewDefaults()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Spec.Username).To(Equal("foo@bar.com"))
		Expect(instance.Spec.PasswordSecret).To(Equal("foo.bar.com-credentials"))
	})

	It("sets a default username", func() {
		instance := &summonv1beta1.DjangoUser{
			ObjectMeta: metav1.ObjectMeta{Name: "foo.example.com"},
			Spec: summonv1beta1.DjangoUserSpec{
				PasswordSecret: "foo-credentials",
			},
		}
		ctx := &components.ComponentContext{Top: instance}

		comp := ducomponents.NewDefaults()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Spec.Username).To(Equal("foo@example.com"))
		Expect(instance.Spec.PasswordSecret).To(Equal("foo-credentials"))
	})

	It("sets a default username and password secret", func() {
		instance := &summonv1beta1.DjangoUser{
			ObjectMeta: metav1.ObjectMeta{Name: "foo.example.com"},
		}
		ctx := &components.ComponentContext{Top: instance}

		comp := ducomponents.NewDefaults()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Spec.Username).To(Equal("foo@example.com"))
		Expect(instance.Spec.PasswordSecret).To(Equal("foo.example.com-credentials"))
	})
})
