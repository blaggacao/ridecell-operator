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
	"testing"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
)

var testHelpers *test_helpers.TestHelpers

func TestTemplates(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "Summon controller Suite")
}

var _ = ginkgo.BeforeSuite(func() {
	testHelpers = test_helpers.Start()
})

var _ = ginkgo.AfterSuite(func() {
	testHelpers.Stop()
})

// SetupTestReconcile returns a reconcile.Reconcile implementation that delegates to inner and
// writes the request to requests after Reconcile is finished.
func SetupTestReconcile(inner reconcile.Reconciler) (reconcile.Reconciler, chan reconcile.Request) {
	requests := make(chan reconcile.Request)
	fn := reconcile.Func(func(req reconcile.Request) (reconcile.Result, error) {
		result, err := inner.Reconcile(req)
		requests <- req
		return result, err
	})
	return fn, requests
}

// StartTestManager adds recFn
func StartTestManager(mgr manager.Manager) chan struct{} {
	stop := make(chan struct{})
	go func() {
		err := mgr.Start(stop)
		gomega.Expect(err).NotTo(gomega.HaveOccurred())
	}()
	return stop
}
