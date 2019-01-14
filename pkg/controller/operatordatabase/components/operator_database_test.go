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

package components_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	//corev1 "k8s.io/api/core/v1"

	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	operatordatabasecomponents "github.com/Ridecell/ridecell-operator/pkg/controller/operatordatabase/components"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("operatordatabase Component", func() {

	BeforeEach(func() {
	})

	It("Adds new users", func() {
		comp := operatordatabasecomponents.NewOperatorDatabase()
		postgresObj := &postgresv1.Postgresql{
			ObjectMeta: metav1.ObjectMeta{Name: "fakeDB", Namespace: "fakeDBNamespace"},
			Spec: postgresv1.PostgresSpec{
				TeamID:            instance.Name,
				NumberOfInstances: int32(1),
				Users: map[string]postgresv1.UserFlags{
					"test-user": postgresv1.UserFlags{"superuser"},
				},
				Databases: map[string]string{
					"test": "test-user",
				},
			},
		}
		ctx.Client = fake.NewFakeClient(postgresObj)
		instance.Spec.DatabaseRef = dbv1beta1.PostgresDBRef{
			Name:      "fakeDB",
			Namespace: "fakeDBNamespace",
		}
		instance.Spec.Users = map[string][]string{
			"test-user": []string{"test"},
			"new-user":  []string{"test"},
			"new-user2": []string{"test", "test2"},
		}

		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())

		fetchPostgresObj := &postgresv1.Postgresql{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "fakeDB", Namespace: "fakeDBNamespace"}, fetchPostgresObj)
		Expect(err).ToNot(HaveOccurred())

		expectedUsers := map[string]postgresv1.UserFlags{
			"test-user": postgresv1.UserFlags{"superuser", "test"},
			"new-user":  postgresv1.UserFlags{"test"},
			"new-user2": postgresv1.UserFlags{"test", "test2"},
		}
		Expect(fetchPostgresObj.Spec.Users).To(Equal(expectedUsers))
		Expect(fetchPostgresObj.Spec.Databases).To(Equal(map[string]string{"test": "test-user"}))
	})

	It("Adds new databases", func() {
		comp := operatordatabasecomponents.NewOperatorDatabase()
		postgresObj := &postgresv1.Postgresql{
			ObjectMeta: metav1.ObjectMeta{Name: "fakeDB", Namespace: "fakeDBNamespace"},
			Spec: postgresv1.PostgresSpec{
				TeamID:            instance.Name,
				NumberOfInstances: int32(1),
				Users: map[string]postgresv1.UserFlags{
					"test-user": postgresv1.UserFlags{"superuser"},
				},
				Databases: map[string]string{
					"test": "test-user",
				},
			},
		}
		ctx.Client = fake.NewFakeClient(postgresObj)
		instance.Spec.DatabaseRef = dbv1beta1.PostgresDBRef{
			Name:      "fakeDB",
			Namespace: "fakeDBNamespace",
		}
		instance.Spec.Databases = map[string]string{
			"test-db":  "test",
			"test-db2": "test2",
		}

		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())

		fetchPostgresObj := &postgresv1.Postgresql{}
		err = ctx.Client.Get(context.TODO(), types.NamespacedName{Name: "fakeDB", Namespace: "fakeDBNamespace"}, fetchPostgresObj)
		Expect(err).ToNot(HaveOccurred())

		expectedDatabases := map[string]string{
			"test":     "test-user",
			"test-db":  "test",
			"test-db2": "test2",
		}

		expectedUsers := map[string]postgresv1.UserFlags{
			"test-user": postgresv1.UserFlags{"superuser"},
		}
		Expect(fetchPostgresObj.Spec.Databases).To(Equal(expectedDatabases))
		Expect(fetchPostgresObj.Spec.Users).To(Equal(expectedUsers))
	})
})
