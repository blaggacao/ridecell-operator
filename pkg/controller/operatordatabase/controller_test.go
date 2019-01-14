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

package operatordatabase_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//corev1 "k8s.io/api/core/v1"

	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
	"k8s.io/apimachinery/pkg/types"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const timeout = time.Second * 60

var _ = Describe("operatordatabase controller", func() {
	var helpers *test_helpers.PerTestHelpers

	BeforeEach(func() {
		helpers = testHelpers.SetupTest()
	})

	AfterEach(func() {
		helpers.TeardownTest()
	})

	It("Runs a basic reconcile", func() {

		postgresObj := &postgresv1.Postgresql{
			ObjectMeta: metav1.ObjectMeta{Name: "fakedb", Namespace: "fakedbnamespace"},
			Spec: postgresv1.PostgresSpec{
				TeamID:            "test",
				NumberOfInstances: int32(1),
				Users: map[string]postgresv1.UserFlags{
					"test-user": postgresv1.UserFlags{"superuser"},
				},
				Databases: map[string]string{
					"test": "test-user",
				},
			},
		}

		err := helpers.Client.Create(context.TODO(), postgresObj)
		Expect(err).ToNot(HaveOccurred())

		instance := &dbv1beta1.PostgresOperatorDatabase{
			ObjectMeta: metav1.ObjectMeta{Name: "test.example.com", Namespace: helpers.Namespace},
			Spec: dbv1beta1.PostgresOperatorDatabaseSpec{
				Databases: map[string]string{
					"test-db": "test-user",
				},
				Users: map[string][]string{
					"test-user": []string{"flag1", "flag2"},
				},
				DatabaseRef: dbv1beta1.PostgresDBRef{
					Name:      "fakedb",
					Namespace: "fakedbnamespace",
				},
			},
		}

		err = helpers.Client.Create(context.TODO(), instance)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() map[string]postgresv1.UserFlags {
			fetchedPostgresObj := &postgresv1.Postgresql{}
			err := helpers.Client.Get(context.TODO(), types.NamespacedName{Name: "fakedb", Namespace: "fakedbnamespace"}, fetchedPostgresObj)
			Expect(err).ToNot(HaveOccurred())
			return fetchedPostgresObj.Spec.Users
		}, timeout).Should(Equal(map[string]postgresv1.UserFlags{"test-user": postgresv1.UserFlags{"superuser", "flag1", "flag2"}}))

		Eventually(func() map[string]string {
			fetchedPostgresObj := &postgresv1.Postgresql{}
			err := helpers.Client.Get(context.TODO(), types.NamespacedName{Name: "fakedb", Namespace: "fakedbnamespace"}, fetchedPostgresObj)
			Expect(err).ToNot(HaveOccurred())
			return fetchedPostgresObj.Spec.Databases
		}, timeout).Should(Equal(map[string]string{"test": "test-user", "test-db": "test-user"}))

	})

})
