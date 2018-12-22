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
	"crypto/sha1"
	"encoding/hex"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("notifications Component", func() {

	BeforeEach(func() {
		os.Setenv("DRY_RUN", "true")
	})

	It("Set StatusReady, match versions", func() {
		instance.Status.Status = summonv1beta1.StatusReady
		comp := summoncomponents.NewNotification()
		Expect(comp.IsReconcilable(ctx)).To(Equal(false))
	})

	It("Set StatusReady, mismatch versions", func() {
		instance.Status.Status = summonv1beta1.StatusReady
		instance.Status.Notification.NotifyVersion = "v9000.1"
		comp := summoncomponents.NewNotification()
		Expect(comp.IsReconcilable(ctx)).To(Equal(true))
	})

	It("Set StatusError, match versions, match errors", func() {
		instance.Status.Status = summonv1beta1.StatusError
		errorMessage := "testError"
		instance.Status.Message = errorMessage

		s := sha1.New()
		hash := s.Sum([]byte(errorMessage))
		encodedHash := hex.EncodeToString(hash)
		instance.Status.Notification.LastErrorHash = encodedHash

		comp := summoncomponents.NewNotification()
		Expect(comp.IsReconcilable(ctx)).To(Equal(false))
	})

	It("Set StatusError, mismatch versions", func() {
		instance.Status.Status = summonv1beta1.StatusError
		instance.Status.Message = "testError"
		instance.Status.Notification.NotifyVersion = "v9000.1"
		comp := summoncomponents.NewNotification()
		Expect(comp.IsReconcilable(ctx)).To(Equal(true))
	})

	It("Set StatusError, match versions, mismatch errors", func() {
		instance.Status.Status = summonv1beta1.StatusError
		instance.Status.Message = "testError"
		comp := summoncomponents.NewNotification()
		Expect(comp.IsReconcilable(ctx)).To(Equal(true))
	})

	It("Set StatusError, match versions, mistmatch errors, reconcile", func() {
		comp := summoncomponents.NewNotification()
		errorMessage := "testError"
		instance.Status.Status = summonv1beta1.StatusError
		instance.Status.Message = errorMessage
		instance.Spec.SlackChannelName = "TestChannel"
		apiKeySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: instance.Namespace, Namespace: "default"},
			Data: map[string][]byte{
				"slackAPIKey": []byte("testapikey"),
			},
		}

		ctx.Client = fake.NewFakeClient(apiKeySecret)
		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())
		Expect(instance.Status.Notification.LastErrorHash).To(Equal(comp.HashStatus(errorMessage)))
	})
})
