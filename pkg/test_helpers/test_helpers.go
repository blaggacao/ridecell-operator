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

package test_helpers

import (
	"encoding/hex"
	"fmt"
	"math/rand"
	"path/filepath"
	"runtime"

	"github.com/onsi/gomega"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"

	"github.com/Ridecell/ridecell-operator/pkg/apis"
)

type TestHelpers struct {
	Environment *envtest.Environment
	Cfg         *rest.Config
}

type PerTestHelpers struct {
	*TestHelpers
	Namespace string
	Client    client.Client
}

func New() (*TestHelpers, error) {
	helpers := &TestHelpers{}
	_, callerLine, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("Unable to find current filename")
	}
	crdPath := filepath.Join(callerLine, "..", "..", "..", "config", "crds")
	helpers.Environment = &envtest.Environment{
		CRDDirectoryPaths: []string{crdPath},
	}
	apis.AddToScheme(scheme.Scheme)

	cfg, err := helpers.Environment.Start()
	if err != nil {
		return nil, err
	}
	helpers.Cfg = cfg

	return helpers, nil
}

// Start up the test environment. Call from BeforeSuite().
func Start() *TestHelpers {
	helpers, err := New()
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return helpers
}

// Shut down the test environment. Call from AfterSuite().
func (helpers *TestHelpers) Stop() {
	helpers.Environment.Stop()
}

// Client returns a controller-runtime REST client. Only use this for non-controller
// tests. In controller tests, use manager.GetClient() instead.
func (helpers *TestHelpers) Client() client.Client {
	client, err := client.New(helpers.Cfg, client.Options{Scheme: scheme.Scheme})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	return client
}

// Set up any needed per test values. Call from BeforeEach().
func (helpers *TestHelpers) SetupTest(client client.Client) *PerTestHelpers {
	newHelpers := &PerTestHelpers{TestHelpers: helpers, Client: client}

	namespaceNameBytes := make([]byte, 10)
	rand.Read(namespaceNameBytes)
	namespaceName := "test-" + hex.EncodeToString(namespaceNameBytes)
	namespace := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: namespaceName}}
	err := client.Create(context.TODO(), namespace)
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
	newHelpers.Namespace = namespaceName

	return newHelpers
}

// Clean up any per test state. Call from AfterEach().
func (helpers *PerTestHelpers) TeardownTest() {
	err := helpers.Client.Delete(context.TODO(), &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: helpers.Namespace}})
	gomega.Expect(err).NotTo(gomega.HaveOccurred())
}
