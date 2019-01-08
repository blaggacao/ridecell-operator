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
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
)

var _ = Describe("notifications Component", func() {

	BeforeEach(func() {
		instance.Spec.SlackChannelName = "test-channel"
		instance.Spec.SlackAPIEndpoint = "foo"
	})

	Describe("isReconcilable", func() {
		It("Check if reconcilable without SlackChannel set", func() {
			instance.Spec.SlackChannelName = ""
			comp := summoncomponents.NewNotification()
			Expect(comp.IsReconcilable(ctx)).To(Equal(false))
		})

		It("Check if reconcilable without SlackAPIEndpoint set", func() {
			instance.Spec.SlackAPIEndpoint = ""
			comp := summoncomponents.NewNotification()
			Expect(comp.IsReconcilable(ctx)).To(Equal(false))
		})

		It("Set StatusReady, match versions", func() {
			instance.Status.Status = summonv1beta1.StatusReady
			comp := summoncomponents.NewNotification()
			Expect(comp.IsReconcilable(ctx)).To(Equal(false))
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

		It("Set StatusError, mismatch versions, match errors", func() {
			instance.Status.Status = summonv1beta1.StatusError
			instance.Status.Notification.NotifyVersion = "v9000.1"
			errorMessage := "testError"
			instance.Status.Message = errorMessage

			s := sha1.New()
			hash := s.Sum([]byte(errorMessage))
			encodedHash := hex.EncodeToString(hash)
			instance.Status.Notification.LastErrorHash = encodedHash
			comp := summoncomponents.NewNotification()
			Expect(comp.IsReconcilable(ctx)).To(Equal(false))
			Expect(instance.Status.Notification.NotifyVersion).To(Equal("v9000.1"))
		})
	})

	It("Set StatusReady, mismatch versions, reconcile", func() {
		instance.Status.Status = summonv1beta1.StatusReady
		instance.Status.Notification.NotifyVersion = "v9000.1"
		fakeAPIKey := "testAPIKey"
		apiKeySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: instance.Spec.NotificationSecretRef.Name, Namespace: "default"},
			Data: map[string][]byte{
				instance.Spec.NotificationSecretRef.Key: []byte(fakeAPIKey),
			},
		}
		ctx.Client = fake.NewFakeClient(apiKeySecret)

		mockServer := getMockHTTPServer(fakeAPIKey, "", "#36a64f", "Deployed")
		defer mockServer.Close()
		// Set SlackAPIEndpoint to the mock server we just created
		instance.Spec.SlackAPIEndpoint = mockServer.URL
		comp := summoncomponents.NewNotification()
		Expect(comp.IsReconcilable(ctx)).To(Equal(true))

		Expect(instance.Spec.Version).To(Equal(instance.Status.Notification.NotifyVersion))
	})

	It("Set StatusError, match versions, mistmatch errors, reconcile", func() {
		// Create an apikey
		fakeAPIKey := "testAPIKey"
		errorMessage := "testError"
		instance.Status.Status = summonv1beta1.StatusError
		instance.Status.Message = errorMessage
		apiKeySecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: instance.Spec.NotificationSecretRef.Name, Namespace: "default"},
			Data: map[string][]byte{
				instance.Spec.NotificationSecretRef.Key: []byte(fakeAPIKey),
			},
		}
		ctx.Client = fake.NewFakeClient(apiKeySecret)

		mockServer := getMockHTTPServer(fakeAPIKey, errorMessage, "#FF0000", "Error")
		defer mockServer.Close()
		// Set SlackAPIEndpoint to the mock server we just created
		instance.Spec.SlackAPIEndpoint = mockServer.URL

		comp := summoncomponents.NewNotification()

		Expect(comp.IsReconcilable(ctx)).To(Equal(true))
		Expect(comp).To(ReconcileContext(ctx))

		Expect(instance.Status.Notification.LastErrorHash).To(Equal("746573744572726f72da39a3ee5e6b4b0d3255bfef95601890afd80709"))
	})
})

func getMockHTTPServer(fakeAPIKey string, messageText string, messageColor string, messageTitle string) *httptest.Server {
	// Create HTTP test server
	testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This was added so that ginkgo can catch a panic if the Expect() in this block fails
		defer GinkgoRecover()
		requestBody, err := ioutil.ReadAll(r.Body)
		defer r.Body.Close()
		var badRequest bool
		if err != nil {
			badRequest = true
		}
		expectedPayload := &summoncomponents.Payload{
			Channel: instance.Spec.SlackChannelName,
			Token:   fakeAPIKey,
			Text:    messageText,
			Attachments: []summoncomponents.Attachments{
				{
					Color:      messageColor,
					AuthorName: "Kubernetes Alert",
					Title:      instance.Spec.Hostname,
					TitleLink:  instance.Spec.Hostname,
					Fields: []summoncomponents.Fields{
						{
							Title: messageTitle,
							Value: instance.Spec.Version,
						},
					},
				},
			},
		}

		var payload *summoncomponents.Payload
		err = json.Unmarshal(requestBody, &payload)
		if err != nil {
			badRequest = true
		}

		Expect(payload).To(Equal(expectedPayload))

		if badRequest {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	return testServer
}
