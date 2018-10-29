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

package summon

import (
	"fmt"
	"time"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
)

const timeout = time.Second * 5

var _ = Describe("Summon controller", func() {
	var helpers *test_helpers.PerTestHelpers
	var c client.Client
	var stopChannel chan struct{}
	var requests chan reconcile.Request

	BeforeEach(func() {
		// Setup the Manager and Controller.  Wrap the Controller Reconcile function so it writes each request to a
		// channel when it is finished.
		mgr, err := manager.New(testHelpers.Cfg, manager.Options{})
		Expect(err).NotTo(HaveOccurred())
		c = mgr.GetClient()
		helpers = testHelpers.SetupTest(c)

		var recFn reconcile.Reconciler
		recFn, requests = SetupTestReconcile(newReconciler(mgr))
		err = add(mgr, recFn)
		Expect(err).NotTo(HaveOccurred())

		stopChannel = StartTestManager(mgr)
	})

	AfterEach(func() {
		close(stopChannel)
		helpers.TeardownTest()
	})

	It("works", func() {
		instance := &summonv1beta1.SummonPlatform{ObjectMeta: metav1.ObjectMeta{Name: "foo", Namespace: helpers.Namespace}}
		expectedRequest := reconcile.Request{NamespacedName: types.NamespacedName{Name: "foo", Namespace: helpers.Namespace}}
		depKey := types.NamespacedName{Name: "foo-deployment", Namespace: helpers.Namespace}

		// Create the Summon object and expect the Reconcile and Deployment to be created
		err := c.Create(context.TODO(), instance)
		// The instance object may not be a valid object because it might be missing some required fields.
		// Please modify the instance object by adding required fields and then remove the following if statement.
		if apierrors.IsInvalid(err) {
			Fail(fmt.Sprintf("failed to create object, got an invalid object error: %v", err))
		}
		Expect(err).NotTo(HaveOccurred())
		defer c.Delete(context.TODO(), instance)
		Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))

		deploy := &appsv1.Deployment{}
		Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).Should(Succeed())

		// Delete the Deployment and expect Reconcile to be called for Deployment deletion
		Expect(c.Delete(context.TODO(), deploy)).NotTo(HaveOccurred())
		Eventually(requests, timeout).Should(Receive(Equal(expectedRequest)))
		Eventually(func() error { return c.Get(context.TODO(), depKey, deploy) }, timeout).Should(Succeed())
	})
})
