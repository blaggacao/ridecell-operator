/*
Copyright 2018 Ridecell, Inc..

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

package summon_test

import (
	"fmt"
	"time"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/manager"

	"github.com/Ridecell/ridecell-operator/pkg/controller/summon"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
)

const timeout = time.Second * 5

var _ = Describe("Summon controller", func() {
	var helpers *test_helpers.PerTestHelpers
	var stopChannel chan struct{}

	BeforeEach(func() {
		// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
		// channel when it is finished.
		mgr, err := manager.New(testHelpers.Cfg, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		helpers = testHelpers.SetupTest(mgr.GetClient())

		err = summon.Add(mgr)
		Expect(err).NotTo(HaveOccurred())

		stopChannel = StartTestManager(mgr)
	})

	AfterEach(func() {
		close(stopChannel)
		helpers.TeardownTest()
	})

	It("works", func() {
		c := helpers.Client
		instance := &summonv1beta1.SummonPlatform{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: helpers.Namespace}}
		depKey := types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}

		// Create the Summon object and expect the Reconcile and Deployment to be created
		err := c.Create(context.TODO(), instance)
		// The instance object may not be a valid object because it might be missing some required fields.
		// Please modify the instance object by adding required fields and then remove the following if statement.
		if apierrors.IsInvalid(err) {
			Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
		}
		Expect(err).NotTo(HaveOccurred())

		service := &corev1.Service{}
		Eventually(func() error { return c.Get(context.TODO(), depKey, service) }, timeout).Should(Succeed())

		// Delete the Service and expect Reconcile to be called for Deployment deletion
		Expect(c.Delete(context.TODO(), service)).NotTo(HaveOccurred())
		Eventually(func() error { return c.Get(context.TODO(), depKey, service) }, timeout).Should(Succeed())
	})

	It("runs a basic reconcile", func() {
		c := helpers.Client
		instance := &summonv1beta1.SummonPlatform{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: helpers.Namespace}, Spec: summonv1beta1.SummonPlatformSpec{
			Version: "1.2.3",
			Secret:  "testsecret",
		}}

		// Create the SummonPlatform object and expect the Reconcile to be created.
		err := c.Create(context.TODO(), instance)
		Expect(err).NotTo(HaveOccurred())
		err = c.Status().Update(context.TODO(), instance)
		Expect(err).NotTo(HaveOccurred())

		// Check the Postgresql object.
		postgres := &postgresv1.Postgresql{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-database", Namespace: helpers.Namespace}, postgres)
		}, timeout).
			Should(Succeed())
		Expect(postgres.Spec.Databases["summon"]).To(Equal("ridecell-admin"))

		// Set the status of the DB to ready.
		postgres.Status = postgresv1.ClusterStatusRunning
		err = c.Status().Update(context.TODO(), postgres)
		Expect(err).NotTo(HaveOccurred())

		// Check that a migration Job was created.
		job := &batchv1.Job{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-migrations", Namespace: helpers.Namespace}, job)
		}, timeout).
			Should(Succeed())

		// Mark the migrations as successful.
		job.Status.Succeeded = 1
		err = c.Status().Update(context.TODO(), job)
		Expect(err).NotTo(HaveOccurred())

		// Check the web Deployment object.
		deploy := &appsv1.Deployment{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
		}, timeout).
			Should(Succeed())
		Expect(deploy.Spec.Replicas).To(PointTo(BeEquivalentTo(1)))
		Expect(deploy.Spec.Template.Spec.Containers[0].Command).To(Equal([]string{"python", "-m", "gunicorn.app.wsgiapp", "-b", "0.0.0.0:8000", "summon_platform.wsgi"}))
		Expect(deploy.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort).To(BeEquivalentTo(8000))

		// Check the web Service object.
		service := &corev1.Service{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, service)
		}, timeout).
			Should(Succeed())
		Expect(service.Spec.Ports[0].Port).To(BeEquivalentTo(8000))

		// Check the web Ingress object.
		ingress := &extv1beta1.Ingress{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, ingress)
		}, timeout).
			Should(Succeed())
		Expect(ingress.Spec.TLS[0].SecretName).To(Equal("testsecret-tls"))

		// Delete the Deployment and expect Reconcile to be called for Deployment deletion
		Expect(c.Delete(context.TODO(), deploy)).NotTo(HaveOccurred())
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
		}, timeout).
			Should(Succeed())
	})

	It("reconciles labels", func() {
		c := helpers.Client
		instance := &summonv1beta1.SummonPlatform{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: helpers.Namespace}, Spec: summonv1beta1.SummonPlatformSpec{
			Version: "1.2.3",
			Secret:  "testsecret",
		}, Status: summonv1beta1.SummonPlatformStatus{
			MigrateVersion: "1.2.3",
		}}

		// Create the SummonPlatform object.
		err := c.Create(context.TODO(), instance)
		Expect(err).NotTo(HaveOccurred())
		err = c.Status().Update(context.TODO(), instance)
		Expect(err).NotTo(HaveOccurred())

		// Set the status of the DB to ready.
		postgres := &postgresv1.Postgresql{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-database", Namespace: helpers.Namespace}, postgres)
		}, timeout).
			Should(Succeed())
		postgres.Status = postgresv1.ClusterStatusRunning
		err = c.Status().Update(context.TODO(), postgres)
		Expect(err).NotTo(HaveOccurred())

		// Fetch the Deployment and check the initial label.
		deploy := &appsv1.Deployment{}
		Eventually(func() error {
			return c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
		}, timeout).
			Should(Succeed())
		Expect(deploy.Labels["app.kubernetes.io/instance"]).To(Equal("foo-web"))
		Expect(deploy.Labels).ToNot(HaveKey("other"))

		// Modify the labels.
		deploy.Labels["app.kubernetes.io/instance"] = "boom"
		deploy.Labels["other"] = "foo"
		err = c.Update(context.TODO(), deploy)
		Expect(err).NotTo(HaveOccurred())

		// Check that the labels end up correct.
		Eventually(func() error {
			err := c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
			if err != nil {
				return err
			}
			value, ok := deploy.Labels["other"]
			if !ok || value != "foo" {
				return fmt.Errorf("No update yet (other)")
			}
			return nil
		}, timeout).Should(Succeed())
		Eventually(func() error {
			err := c.Get(context.TODO(), types.NamespacedName{Name: "foo-web", Namespace: helpers.Namespace}, deploy)
			if err != nil {
				return err
			}
			value, ok := deploy.Labels["app.kubernetes.io/instance"]
			if !ok || value != "foo-web" {
				return fmt.Errorf("No update yet (app.kubernetes.io/instance)")
			}
			return nil
		}, timeout).Should(Succeed())
	})
})
