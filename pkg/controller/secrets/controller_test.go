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

package secrets_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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

	It("Gets pull-secret", func() {
		// Create Pull Secret
		pullSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "pull-secret", Namespace: helpers.Namespace}, Type: "kubernetes.io/dockerconfigjson", StringData: map[string]string{".dockerconfigjson": "{\"auths\": {}}"}}
		err := helpers.Client.Create(context.TODO(), pullSecret)
		Expect(err).NotTo(HaveOccurred())
		Eventually(func() error {
			return helpers.Client.Get(context.TODO(), types.NamespacedName{Name: "pull-secret", Namespace: helpers.Namespace}, &corev1.Secret{})
		}, timeout).Should(Succeed())
	})
})
