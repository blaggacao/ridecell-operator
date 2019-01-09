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
	"fmt"
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

const defaultSlackEndpoint = "https://slack.com/api/"

type notificationComponent struct{}

// Fields is nested inside of of Attachments for building Json payload
type Fields struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
}

// Attachments is nested inside of Payload for building Json payload
type Attachments struct {
	Color      string   `json:"color,omitempty"`
	AuthorName string   `json:"author_name,omitempty"`
	Title      string   `json:"title,omitempty"`
	TitleLink  string   `json:"title_link,omitempty"`
	Fields     []Fields `json:"fields,omitempty"`
}

// Payload is the base structure for building Json payload
type Payload struct {
	Channel     string        `json:"channel,omitempty"`
	Text        string        `json:"text,omitempty"`
	Name        string        `json:"name,omitempty"`
	AsUser      bool          `json:"AsUser,omitempty"`
	Attachments []Attachments `json:"attachments,omitempty"`
	Validate    bool          `json:"valdiate,omitempty"`
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
	var slackMessageURL string
	var slackJoinChannelURL string
	// If this is not set, use the actual endpoint + api path
	// Else use the mock http server
	if slackURL == "" {
		slackMessageURL = fmt.Sprintf("%schat.postMessage", defaultSlackEndpoint)
		slackJoinChannelURL = fmt.Sprintf("%schannels.join", defaultSlackEndpoint)
	} else {
		slackMessageURL = slackURL
		slackJoinChannelURL = slackURL
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

	// Create our POST payload
	rawPayload := comp.formatPayload(ctx)
	payload, err := json.Marshal(rawPayload)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: Unable to json.Marshal(rawPayload)")
	}

	// create POST request with payload, add headers, execute
	client := &http.Client{}
	req, err := http.NewRequest("POST", slackMessageURL, bytes.NewBuffer(payload))
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: failed to create post request")
	}
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", apiKey))
	req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	// Test if the request was actually sent, and make sure we got a 200
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "notifications: Unable to send POST request.")
	}
	// Set body to close after function call to avoid errors
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)

	// Slack almost always returns 200s, if this occurs it's likely not a code issue
	if resp.StatusCode != http.StatusOK {
		if err != nil {
			return reconcile.Result{}, errors.New("notifications: Failed to read body from non 200 HTTP StatusCode")
		}
		return reconcile.Result{}, errors.Errorf("notifications: HTTP StatusCode = %v, body of response = %#v", resp.StatusCode, body)
	}

	// Unpackage response from our request
	var jsonResponse map[string]interface{}
	json.Unmarshal(body, &jsonResponse)

	// If respStatus is not true slack returned us an error
	respStatus, ok := jsonResponse["ok"]
	if !ok || err != nil {
		return reconcile.Result{}, errors.New(`notifications: could not find "ok" in slack response`)
	}

	if respStatus == false {
		// This value is only returned when an error occurs
		respError := jsonResponse["error"]
		// If our slack api key is wrong exit reconcile
		if respStatus == false && respError == "invalid_auth" {
			return reconcile.Result{}, errors.New("notifications: invalid auth token for slack request")
			// if our bot is not in the slack channel join it and requeue the reconcile
			// Message should send on next attempt
		} else if respStatus == false && respError == "not_in_channel" {
			err = comp.joinSlackChannel(ctx, apiKey, slackJoinChannelURL)
			if err != nil {
				return reconcile.Result{}, errors.Wrapf(err, "notifications: error in joinSlackChannel")
			}
			return reconcile.Result{Requeue: true}, errors.New("notifications: bot is not in slack channel, sent join request and requeued")
		}
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

func (comp *notificationComponent) joinSlackChannel(ctx *components.ComponentContext, slackAPIKey string, slackJoinChannelURL string) error {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	rawPayload := Payload{
		Name:     instance.Spec.SlackChannelName,
		Validate: true,
	}
	payload, err := json.Marshal(rawPayload)
	if err != nil {
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", slackJoinChannelURL, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}

	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", slackAPIKey))
	req.Header.Add("Content-type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	responseBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	respBody := make(map[string]interface{})
	json.Unmarshal(responseBody, &respBody)
	respStatus, ok := respBody["ok"]
	if !ok {
		return errors.New(`unable to find "ok" in channel join request`)
	}
	respError := respBody["error"]
	if respStatus == false {
		return errors.Errorf("%s", respError)
	}
	return nil
}

func (comp *notificationComponent) formatPayload(ctx *components.ComponentContext) Payload {
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
