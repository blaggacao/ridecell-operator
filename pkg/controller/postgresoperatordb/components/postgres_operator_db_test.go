/*
Copyright 2018-2019 Ridecell, Inc.

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

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//corev1 "k8s.io/api/core/v1"

	. "github.com/Ridecell/ridecell-operator/pkg/test_helpers/matchers"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	postgresoperatordbcomponents "github.com/Ridecell/ridecell-operator/pkg/controller/postgresoperatordb/components"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("operatordatabase Component", func() {

	BeforeEach(func() {
	})

	It("Adds new database", func() {
		comp := postgresoperatordbcomponents.NewPostgresOperatorDB()
		postgresObj := &postgresv1.Postgresql{
			ObjectMeta: metav1.ObjectMeta{Name: "fakedb", Namespace: instance.Namespace},
			Spec: postgresv1.PostgresSpec{
				TeamID:            instance.Name,
				NumberOfInstances: int32(1),
				Databases: map[string]string{
					"test-db": "test-db",
				},
				Users: map[string]postgresv1.UserFlags{
					"test-db": postgresv1.UserFlags{},
				},
			},
		}
		ctx.Client = fake.NewFakeClient(postgresObj)
		instance.Spec.DatabaseRef = dbv1beta1.PostgresDBRef{
			Name: "fakedb",
		}
		instance.Spec.Database = "test-db2"

		Expect(comp).To(ReconcileContext(ctx))

		fetchPostgresObj := &postgresv1.Postgresql{}
		err := ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "fakedb", Namespace: instance.Namespace}, fetchPostgresObj)
		Expect(err).ToNot(HaveOccurred())

		expectedDatabases := map[string]string{
			"test-db":  "test-db",
			"test-db2": "test-db2",
		}

		expectedUsers := map[string]postgresv1.UserFlags{
			"test-db":  postgresv1.UserFlags{},
			"test-db2": postgresv1.UserFlags{},
		}

		Expect(fetchPostgresObj.Spec.Databases).To(Equal(expectedDatabases))
		Expect(fetchPostgresObj.Spec.Users).To(Equal(expectedUsers))
	})

	It("does not find postgres object", func() {
		comp := postgresoperatordbcomponents.NewPostgresOperatorDB()
		ctx.Client = fake.NewFakeClient()

		instance.Spec.Database = "test-db"
		Expect(comp).ToNot(ReconcileContext(ctx))
	})
	It("is given a postgres object that is owned", func() {
		comp := postgresoperatordbcomponents.NewPostgresOperatorDB()
		postgresObj := &postgresv1.Postgresql{
			ObjectMeta: metav1.ObjectMeta{Name: "fakedb", Namespace: instance.Namespace},
			Spec: postgresv1.PostgresSpec{
				TeamID:            instance.Name,
				NumberOfInstances: int32(1),
				Databases: map[string]string{
					"test-db": "test-db",
				},
				Users: map[string]postgresv1.UserFlags{
					"test-db": postgresv1.UserFlags{},
				},
			},
		}
		// Sets ownership of postgres object
		postgresObj.SetOwnerReferences([]metav1.OwnerReference{metav1.OwnerReference{}})
		ctx.Client = fake.NewFakeClient(postgresObj)
		instance.Spec.DatabaseRef = dbv1beta1.PostgresDBRef{
			Name: "fakedb",
		}
		instance.Spec.Database = "test-db2"

		Expect(comp).ToNot(ReconcileContext(ctx))
	})

	It("makes no changes during reconcile", func() {
		comp := postgresoperatordbcomponents.NewPostgresOperatorDB()
		postgresObj := &postgresv1.Postgresql{
			ObjectMeta: metav1.ObjectMeta{Name: "fakedb", Namespace: instance.Namespace},
			Spec: postgresv1.PostgresSpec{
				TeamID:            instance.Name,
				NumberOfInstances: int32(1),
				Databases: map[string]string{
					"test-db": "test-db",
				},
				Users: map[string]postgresv1.UserFlags{
					"test-db": postgresv1.UserFlags{},
				},
			},
		}
		ctx.Client = fake.NewFakeClient(postgresObj)
		instance.Spec.DatabaseRef = dbv1beta1.PostgresDBRef{
			Name: "fakedb",
		}
		instance.Spec.Database = "test-db"
		Expect(comp).To(ReconcileContext(ctx))
	})
})
