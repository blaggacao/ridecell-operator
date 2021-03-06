/*
Copyright 2019 Ridecell, Inc.

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
	"context"
	"fmt"

	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("deployment Component", func() {

	It("runs a basic reconcile", func() {
		comp := summoncomponents.NewDeployment("static/deployment.yml.tpl")

		// Set this value so created template does not contain a nil value
		numReplicas := int32(1)
		instance.Spec.StaticReplicas = &numReplicas

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-config", instance.Name), Namespace: instance.Namespace},
			Data:       map[string]string{"summon-platform.yml": "{}\n"},
		}

		appSecrets := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("summon.%s.app-secrets", instance.Name), Namespace: instance.Namespace},
			Data: map[string][]byte{
				"filler": []byte("test"),
				"test":   []byte("another_test"),
			},
		}

		ctx.Client = fake.NewFakeClient(appSecrets, configMap)
		Expect(comp).To(ReconcileContext(ctx))

		expectedAppSecrets := "7b2266696c6c6572223a226447567a64413d3d222c2274657374223a22595735766447686c636c39305a584e30227dda39a3ee5e6b4b0d3255bfef95601890afd80709"
		expectedConfigHash := "7b2273756d6d6f6e2d706c6174666f726d2e796d6c223a227b7d5c6e227dda39a3ee5e6b4b0d3255bfef95601890afd80709"
		deployment := &appsv1.Deployment{}
		err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-static", Namespace: instance.Namespace}, deployment)
		Expect(err).ToNot(HaveOccurred())
		deploymentPodAnnotations := deployment.Spec.Template.Annotations
		Expect(deploymentPodAnnotations["summon.ridecell.io/appSecretsHash"]).To(Equal(expectedAppSecrets))
		Expect(deploymentPodAnnotations["summon.ridecell.io/configHash"]).To(Equal(expectedConfigHash))
	})

	It("makes sure keys are sorted before hash", func() {
		comp := summoncomponents.NewDeployment("static/deployment.yml.tpl")

		// Set this value so created template does not contain a nil value
		numReplicas := int32(1)
		instance.Spec.StaticReplicas = &numReplicas

		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-config", instance.Name), Namespace: instance.Namespace},
			Data:       map[string]string{"summon-platform.yml": "{}\n"},
		}

		appSecrets := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("summon.%s.app-secrets", instance.Name), Namespace: instance.Namespace},
			Data: map[string][]byte{
				"test":   []byte("another_test"),
				"filler": []byte("test"),
			},
		}

		ctx.Client = fake.NewFakeClient(appSecrets, configMap)
		Expect(comp).To(ReconcileContext(ctx))

		expectedAppSecrets := "7b2266696c6c6572223a226447567a64413d3d222c2274657374223a22595735766447686c636c39305a584e30227dda39a3ee5e6b4b0d3255bfef95601890afd80709"
		deployment := &appsv1.Deployment{}
		err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-static", Namespace: instance.Namespace}, deployment)
		Expect(err).ToNot(HaveOccurred())
		deploymentPodAnnotations := deployment.Spec.Template.Annotations
		Expect(deploymentPodAnnotations["summon.ridecell.io/appSecretsHash"]).To(Equal(expectedAppSecrets))
	})

	It("updates existing hashes", func() {
		comp := summoncomponents.NewDeployment("static/deployment.yml.tpl")

		// Set this value so created template does not contain a nil value
		numReplicas := int32(1)
		instance.Spec.StaticReplicas = &numReplicas

		// Create our first hashes
		configMap := &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-config", instance.Name), Namespace: instance.Namespace},
			Data:       map[string]string{"summon-platform.yml": "{}\n"},
		}

		appSecrets := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("summon.%s.app-secrets", instance.Name), Namespace: instance.Namespace},
			Data:       map[string][]byte{"filler": []byte("test")},
		}

		ctx.Client = fake.NewFakeClient(appSecrets, configMap)
		Expect(comp).To(ReconcileContext(ctx))

		// Create our second hashes
		configMap = &corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s-config", instance.Name), Namespace: instance.Namespace},
			Data:       map[string]string{"summon-platform.yml": "{test}\n"},
		}

		appSecrets = &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("summon.%s.app-secrets", instance.Name), Namespace: instance.Namespace},
			Data:       map[string][]byte{"filler": []byte("test2")},
		}

		ctx.Client = fake.NewFakeClient(appSecrets, configMap)
		Expect(comp).To(ReconcileContext(ctx))

		expectedAppSecrets := "7b2266696c6c6572223a226447567a6444493d227dda39a3ee5e6b4b0d3255bfef95601890afd80709"
		expectedConfigHash := "7b2273756d6d6f6e2d706c6174666f726d2e796d6c223a227b746573747d5c6e227dda39a3ee5e6b4b0d3255bfef95601890afd80709"
		deployment := &appsv1.Deployment{}
		err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "foo-static", Namespace: instance.Namespace}, deployment)
		Expect(err).ToNot(HaveOccurred())
		deploymentPodAnnotations := deployment.Spec.Template.Annotations
		Expect(deploymentPodAnnotations["summon.ridecell.io/appSecretsHash"]).To(Equal(expectedAppSecrets))
		Expect(deploymentPodAnnotations["summon.ridecell.io/configHash"]).To(Equal(expectedConfigHash))

	})
})
