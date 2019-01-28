/*
Copyright 2018-2019 Ridecell, Inc.
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

package secrets_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	secretsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/secrets/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
	"k8s.io/apimachinery/pkg/types"
)

const timeout = time.Second * 10

var _ = Describe("Secrets controller", func() {
	var helpers *test_helpers.PerTestHelpers

	BeforeEach(func() {
		helpers = testHelpers.SetupTest()
	})

	AfterEach(func() {
		helpers.TeardownTest()
	})

	It("Gets pull-secret when it does not exist", func() {
		Eventually(func() error {
			return helpers.Client.Get(context.TODO(), types.NamespacedName{Name: "pull-secret", Namespace: helpers.Namespace}, &corev1.Secret{})
		}, timeout).ShouldNot(Succeed())
	})

	It("Sets secret, Gets secret", func() {
		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "pull-secret", Namespace: helpers.OperatorNamespace}}
		err := helpers.Client.Create(context.TODO(), secret)
		Expect(err).NotTo(HaveOccurred())
		pullSecret := &secretsv1beta1.PullSecret{ObjectMeta: metav1.ObjectMeta{Name: "secrets.ridecell.us", Namespace: helpers.Namespace}}
		err = helpers.Client.Create(context.TODO(), pullSecret)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return helpers.Client.Get(context.TODO(), types.NamespacedName{Name: "pull-secret", Namespace: helpers.Namespace}, &corev1.Secret{})
		}, timeout).Should(Succeed())
	})

	It("Sets secret in Operator namespace only", func() {
		pullSecret := &secretsv1beta1.PullSecret{ObjectMeta: metav1.ObjectMeta{Name: "pull-secret", Namespace: helpers.Namespace}}
		err := helpers.Client.Create(context.TODO(), pullSecret)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return helpers.Client.Get(context.TODO(), types.NamespacedName{Name: "pull-secret", Namespace: helpers.Namespace}, &corev1.Secret{})
		}, timeout).ShouldNot(Succeed())
	})

	It("Sets secret with non-default name, gets secret", func() {
		secret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: helpers.OperatorNamespace}}
		err := helpers.Client.Create(context.TODO(), secret)
		Expect(err).NotTo(HaveOccurred())
		pullSecret := &secretsv1beta1.PullSecret{ObjectMeta: metav1.ObjectMeta{Name: "secrets.ridecell.us", Namespace: helpers.Namespace}, Spec: secretsv1beta1.PullSecretSpec{PullSecretName: "foo"}}
		err = helpers.Client.Create(context.TODO(), pullSecret)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return helpers.Client.Get(context.TODO(), types.NamespacedName{Name: "foo", Namespace: helpers.Namespace}, &corev1.Secret{})
		}, timeout).Should(Succeed())
	})
})
