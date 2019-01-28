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

package postgresextension_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	apihelpers "github.com/Ridecell/ridecell-operator/pkg/apis/helpers"
	"github.com/Ridecell/ridecell-operator/pkg/dbpool"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers/fake_sql"
)

const timeout = time.Second * 5

var _ = Describe("PostgresExtension controller", func() {
	var helpers *test_helpers.PerTestHelpers

	BeforeEach(func() {
		helpers = testHelpers.SetupTest()
		dbpool.Dbs.Store("postgres host=test-database port=5432 dbname=summon user=root password='secretdbpass' sslmode=require", fake_sql.Open())
	})

	AfterEach(func() {
		helpers.TeardownTest()
		dbpool.Dbs.Delete("postgres host=test-database port=5432 dbname=summon user=root password='secretdbpass' sslmode=require")
	})

	It("runs a basic reconcile", func() {
		c := helpers.Client

		secret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "creds", Namespace: helpers.Namespace},
			StringData: map[string]string{
				"passwd": "secretdbpass",
			},
		}
		err := c.Create(context.TODO(), secret)
		Expect(err).ToNot(HaveOccurred())

		ext := &dbv1beta1.PostgresExtension{
			ObjectMeta: metav1.ObjectMeta{Name: "postgis", Namespace: helpers.Namespace},
			Spec: dbv1beta1.PostgresExtensionSpec{
				Database: dbv1beta1.PostgresConnection{
					Host:     "test-database",
					Username: "root",
					Database: "summon",
					PasswordSecretRef: apihelpers.SecretRef{
						Name: "creds",
						Key:  "passwd",
					},
				},
			},
		}
		err = c.Create(context.TODO(), ext)
		Expect(err).ToNot(HaveOccurred())

		Eventually(func() (string, error) {
			err := c.Get(context.TODO(), types.NamespacedName{Name: "postgis", Namespace: helpers.Namespace}, ext)
			if err != nil {
				return "", err
			}
			return ext.Status.Status, nil
		}, timeout).Should(Equal(dbv1beta1.StatusReady))
	})
})
