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

package components

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

type notificationComponent struct{}

func NewNotification() *notificationComponent {
	return &notificationComponent{}
}

func (comp *notificationComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (comp *notificationComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	if instance.Status.Status == summonv1beta1.StatusReady || instance.Status.Status == summonv1beta1.StatusError {
		hashedError := comp.HashStatus(instance.Status.Message)
		if comp.isMismatchedVersion(ctx) {
			return true

		} else if comp.isMismatchedError(ctx, hashedError) {
			return true
		}
	}

	return false
}

func (comp *notificationComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	// Try to find the Slack API Key
	secret := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: instance.Spec.SlackAPISecret, Namespace: instance.Namespace}, secret)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "notifications: Unable to load slackAPIKey secret %s/%s", instance.Namespace, instance.Spec.SlackAPISecret)
	}
	apiKeyByte, ok := secret.Data["slackAPIKey"]
	if !ok {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: apiKey secret %s/%s has no key \"slackAPIKey\"", instance.Namespace, instance.Spec.SlackAPISecret)
	}
	apiKey := string(apiKeyByte)

	// Send Slack Notification if ChannelName is set
	if instance.Spec.SlackChannelName != "" {

		alertText := "Test"

		// Check if we are running tests
		dryRunEnv := os.Getenv("DRY_RUN")
		var dryRun bool
		if dryRunEnv != "" {
			dryRun = true
		}

		// If running tests create mock HTTP server and direct requests to it.
		var postURL string
		if dryRun {
			testServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer r.Body.Close()
				requestBody, err := ioutil.ReadAll(r.Body)
				var badRequest bool
				if err != nil {
					badRequest = true
				}
				type jsonStruct struct {
					Token   string `json:"token"`
					Channel string `json:"channel"`
					Text    string `json:"text"`
				}
				var jsonPayload *jsonStruct
				err = json.Unmarshal(requestBody, &jsonPayload)
				if err != nil {
					badRequest = true
				}
				if jsonPayload.Token != apiKey || jsonPayload.Channel != instance.Spec.SlackChannelName || jsonPayload.Text != alertText {
					badRequest = true
				}
				if badRequest {
					w.WriteHeader(http.StatusBadRequest)
				} else {
					w.WriteHeader(http.StatusOK)
				}

			}))
			defer testServer.Close()
			postURL = testServer.URL
		} else {
			// Hardcoding this for now.
			postURL = ""
		}
		// Send POST request
		rawPayload := map[string]string{"token": apiKey, "channel": instance.Spec.SlackChannelName, "text": alertText}
		payload, err := json.Marshal(rawPayload)
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "notifications: Unable to json.Marshal(rawPayload)")
		}

		resp, err := http.Post(postURL, "application/json", bytes.NewBuffer(payload))
		// Test if the request was actually sent, and make sure we got a 200
		if err != nil {
			return reconcile.Result{}, errors.Wrapf(err, "notifications: Unable to send POST request.")
		}
		if resp.StatusCode != http.StatusOK {
			return reconcile.Result{}, errors.Errorf("notifications: HTTP StatusCode = %v", resp.StatusCode)
		}
	}

	// Update NotifyVersion if it needs to be changed.
	if comp.isMismatchedVersion(ctx) {
		instance.Status.Notification.NotifyVersion = instance.Spec.Version
	}

	// Update LastErrorHash if it needs to be updated.
	encodedHash := comp.HashStatus(instance.Status.Message)
	if comp.isMismatchedError(ctx, encodedHash) {
		instance.Status.Notification.LastErrorHash = encodedHash
	}

	return reconcile.Result{}, nil
}

func (comp *notificationComponent) HashStatus(status string) string {
	// Turns instance.Status.Message into sha1 -> hex -> string
	hash := sha1.New().Sum([]byte(status))
	encodedHash := hex.EncodeToString(hash)
	return encodedHash
}

func (comp *notificationComponent) isMismatchedVersion(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	if instance.Spec.Version != instance.Status.Notification.NotifyVersion {
		return true
	}
	return false
}

func (comp *notificationComponent) isMismatchedError(ctx *components.ComponentContext, errorHash string) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	if instance.Status.Status == summonv1beta1.StatusError {
		if errorHash != instance.Status.Notification.LastErrorHash {
			return true
		}
	}
	return false
}
