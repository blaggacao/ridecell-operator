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
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"os"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/nlopes/slack"
	"k8s.io/apimachinery/pkg/runtime"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
)

// Interface for the Slack client to allow for a mock implementation.
type SlackClient interface {
	PostMessage(string, ...self.MsgOption) (string, string, error)
}

type notificationComponent struct {
	slackClient SlackSender
}

func NewNotification() *notificationComponent {
	var slackClient *slack.Client
	slackApiKey = os.Getenv("SLACK_API_KEY")
	if slackApiKey != "" {
		slackClient = slack.New(slackApiKey)
	}
	return &notificationComponent{slackClient: slackClient}
}

func (c *notificationComponent) InjectSlackClient(client SlackClient) {
	c.slackClient = client
}

func (_ *notificationComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *notificationComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	return instance.Spec.Notifications.SlackChannel != ""
}

func (c *notificationComponent) Reconcile(ctx *components.ComponentContext) (components.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	if instance.Status.Status == summonv1beta1.StatusReady {
		return c.handleSuccess(instance)
	} else if instance.Status.Status == summonv1beta1.StatusError {
		return c.handleError(instance, instance.Status.Message)
	}

	// No notifications needed.
	return components.Result{}, nil
}

// ReconcileError implements components.ErrorHandler.
func (c *notificationComponent) ReconcileError(ctx *components.ComponentContext, err error) (components.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	return c.handleError(instance, err.Error())
}

// Send a deploy notification if needed.
func (c *notificationComponent) handleSuccess(instance *summonv1beta1.SummonPlatform) (components.Result, error) {
	if instance.Spec.Version == instance.Status.Notification.NotifyVersion {
		// Already notified about this version, we're good.
		return components.Result{}, nil
	}

	// Send to Slack.
	attachment := c.formatSuccessNotification(instance)
	_, _, err := c.slackClient.PostMessage(instance.Spec.Notifications.SlackChannel, attachment)
	if err != nil {
		return components.Result{}, err
	}

	// Update status. Close over `version` in case it changes during a collision.
	version := instance.Spec.Version
	return components.Result{StatusModifier: func(obj runtime.Object) error {
		instance := obj.(*summonv1beta1.SummonPlatform)
		instance.Status.Notification.NotifyVersion = version
		return nil
	}}, nil
}

// Send an error notification if needed.
func (c *notificationComponent) handleError(instance *summonv1beta1.SummonPlatform, errorMessage string) (components.Result, error) {
	// Compute the hash of the current error, so we don't actually store the whole error message twice.
	rawErrorHash := sha1.New().Sum([]byte(errorMessage))
	errorHash := hex.EncodeToString(rawErrorHash)

	if errorHash == instance.Status.Notification.LastErrorHash {
		// Already notified about this error, we're good.
		return components.Result{}, nil
	}

	// Send to Slack.
	attachment := c.formatErrorNotification(instance, errorMessage)
	_, _, err := c.slackClient.PostMessage(instance.Spec.Notifications.SlackChannel, attachment)
	if err != nil {
		return components.Result{}, err
	}

	// Update status.
	return components.Result{StatusModifier: func(obj runtime.Object) error {
		instance := obj.(*summonv1beta1.SummonPlatform)
		instance.Status.Notification.LastErrorHash = errorHash
		return nil
	}}, nil
}

// Render the nofiication attachement for a deploy notification.
func (comp *notificationComponent) formatSuccessNotification(instance *summonv1beta1.SummonPlatform) slack.Attachment {
	return slack.Attachment{
		Title:     fmt.Sprintf("%s Deployment", instance.Spec.Hostname),
		TitleLink: fmt.Sprintf("https://%s/", instance.Spec.Hostname),
		Color:     "good",
		Text:      fmt.Sprintf("<https://%s/|%s> deployed version %s successfully", instance.Spec.Hostname, instance.Spec.Hostname, instance.Spec.Version),
		Fallback:  fmt.Sprintf("%s deployed version %s successfully", instance.Spec.Hostname, instance.Spec.Version),
	}
}

// Render the nofiication attachement for an error notification.
func (comp *notificationComponent) formatErrorNotification(instance *summonv1beta1.SummonPlatform, errorMessage string) slack.Attachment {
	return slack.Attachment{
		Title:     fmt.Sprintf("%s Deployment", instance.Spec.Hostname),
		TitleLink: fmt.Sprintf("https://%s/", instance.Spec.Hostname),
		Color:     "danger",
		Text:      fmt.Sprintf("<https://%s/|%s> has error: %s", instance.Spec.Hostname, instance.Spec.Hostname, errorMessage),
		Fallback:  fmt.Sprintf("%s has error: %s", instance.Spec.Hostname, errorMessage),
	}
}
