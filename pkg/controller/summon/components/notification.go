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
	"net/http"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

type notificationComponent struct{}

// Fields is nested inside of of Attachments for building Json payload
type Fields struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// Attachments is nested inside of PayloadMessage for building Json payload
type Attachments struct {
	Color      string   `json:"color"`
	AuthorName string   `json:"author_name"`
	Title      string   `json:"title"`
	TitleLink  string   `json:"title_link"`
	Fields     []Fields `json:"fields"`
}

// PayloadMessage is the base structure for building Json payload
type PayloadMessage struct {
	Channel     string        `json:"channel"`
	Token       string        `json:"token"`
	Text        string        `json:"text"`
	Attachments []Attachments `json:"attachments"`
}

func NewNotification() *notificationComponent {
	return &notificationComponent{}
}

func (comp *notificationComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (comp *notificationComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	// Don't send notification is slackChannel or slackApiEndpoint are not defined.
	if instance.Spec.SlackChannelName == "" || instance.Spec.SlackAPIEndpoint == "" {
		return false
	}
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
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: instance.Spec.NotificationSecretRef.Name, Namespace: instance.Namespace}, secret)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "notifications: Unable to load slackAPIKey secret %s/%s", instance.Namespace, instance.Spec.NotificationSecretRef.Name)
	}
	apiKeyByte, ok := secret.Data[instance.Spec.NotificationSecretRef.Key]
	if !ok {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: apiKey secret %s/%s has no key \"slackAPIKey\"", instance.Namespace, instance.Spec.NotificationSecretRef.Name)
	}
	apiKey := string(apiKeyByte)

	var messageColor, messageText, messageTitle string
	if instance.Status.Status == summonv1beta1.StatusError {
		messageColor = "#FF0000"
		messageText = instance.Status.Message
		messageTitle = "Error"
	} else {
		messageColor = "#36a64f"
		messageText = ""
		messageTitle = "Deployed"
	}

	rawPayloadMessage := PayloadMessage{
		Channel: instance.Spec.SlackChannelName,
		Token:   apiKey,
		Text:    messageText,
		Attachments: []Attachments{
			{
				Color:      messageColor,
				AuthorName: "Kubernetes Alert",
				Title:      instance.Spec.Hostname,
				TitleLink:  instance.Spec.Hostname,
				Fields: []Fields{
					{
						Title: messageTitle,
						Value: instance.Spec.Version,
					},
				},
			},
		},
	}
	payload, err := json.Marshal(rawPayloadMessage)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: Unable to json.Marshal(rawPayload)")
	}

	resp, err := http.Post(instance.Spec.SlackAPIEndpoint, "application/json", bytes.NewBuffer(payload))
	// Test if the request was actually sent, and make sure we got a 200
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: Unable to send POST request.")
	}
	// Set body to close after function call to avoid errors
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return reconcile.Result{}, errors.Errorf("notifications: HTTP StatusCode = %v", resp.StatusCode)
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
