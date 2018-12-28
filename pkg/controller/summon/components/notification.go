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

	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	corev1 "k8s.io/api/core/v1"
)

const defaultSlackEndpoint = "https://slack.com/api/chat.postMessage"

type notificationComponent struct{}

// Fields is nested inside of of Attachments for building Json payload
type Fields struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

// Attachments is nested inside of Payload for building Json payload
type Attachments struct {
	Color      string   `json:"color"`
	AuthorName string   `json:"author_name"`
	Title      string   `json:"title"`
	TitleLink  string   `json:"title_link"`
	Fields     []Fields `json:"fields"`
}

// Payload is the base structure for building Json payload
type Payload struct {
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
	if instance.Spec.SlackChannelName == "" {
		return false
	}
	if instance.Status.Status == summonv1beta1.StatusReady {
		return comp.isMismatchedVersion(ctx)
	} else if instance.Status.Status == summonv1beta1.StatusError {
		hashedError := comp.hashStatus(instance.Status.Message)
		return comp.isMismatchedError(ctx, hashedError)
	}
	return false
}

func (comp *notificationComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	slackURL := instance.Spec.SlackAPIEndpoint
	if slackURL == "" {
		slackURL = defaultSlackEndpoint
	}
	// Try to find the Slack API Key
	secret := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: instance.Spec.NotificationSecretRef.Name, Namespace: instance.Namespace}, secret)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "notifications: Unable to load slackAPIKey secret %s/%s", instance.Namespace, instance.Spec.NotificationSecretRef.Name)
	}
	apiKeyByte, ok := secret.Data[instance.Spec.NotificationSecretRef.Key]
	if !ok {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: apiKey secret %s/%s has no key \"%s\"", instance.Namespace, instance.Spec.NotificationSecretRef.Name, instance.Spec.NotificationSecretRef.Key)
	}
	apiKey := string(apiKeyByte)

	rawPayload := comp.formatPayload(ctx, apiKey)

	payload, err := json.Marshal(rawPayload)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: Unable to json.Marshal(rawPayload)")
	}

	resp, err := http.Post(slackURL, "application/json", bytes.NewBuffer(payload))
	// Test if the request was actually sent, and make sure we got a 200
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: Unable to send POST request.")
	}
	// Set body to close after function call to avoid errors
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return reconcile.Result{}, errors.Errorf("notifications: Failed to read body from non 200 HTTP StatusCode")
		}
		return reconcile.Result{}, errors.Errorf("notifications: HTTP StatusCode = %v, body of response = %#v", resp.StatusCode, body)
	}

	// Update NotifyVersion if it needs to be changed.
	if instance.Status.Status == summonv1beta1.StatusReady && comp.isMismatchedVersion(ctx) {
		instance.Status.Notification.NotifyVersion = instance.Spec.Version
	}

	// Update LastErrorHash if it needs to be updated.
	encodedHash := comp.hashStatus(instance.Status.Message)
	if instance.Status.Status == summonv1beta1.StatusError && comp.isMismatchedError(ctx, encodedHash) {
		instance.Status.Notification.LastErrorHash = encodedHash
	}

	return reconcile.Result{}, nil
}

func (comp *notificationComponent) hashStatus(status string) string {
	// Turns instance.Status.Message into sha1 -> hex -> string
	hash := sha1.New().Sum([]byte(status))
	encodedHash := hex.EncodeToString(hash)
	return encodedHash
}

func (comp *notificationComponent) isMismatchedVersion(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	return instance.Spec.Version != instance.Status.Notification.NotifyVersion
}

func (comp *notificationComponent) isMismatchedError(ctx *components.ComponentContext, errorHash string) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	return instance.Status.Status == summonv1beta1.StatusError && errorHash != instance.Status.Notification.LastErrorHash
}

func (comp *notificationComponent) formatPayload(ctx *components.ComponentContext, apiKey string) Payload {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
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

	rawPayload := Payload{
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

	return rawPayload
}
