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

package components_test

import (
	"database/sql"
	"fmt"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	djangousercomponents "github.com/Ridecell/ridecell-operator/pkg/controller/djangouser/components"
	"github.com/Ridecell/ridecell-operator/pkg/dbpool"
)

var _ = Describe("DjangoUser Database Component", func() {
	var dbMock sqlmock.Sqlmock
	var db *sql.DB

	BeforeEach(func() {
		var err error
		db, dbMock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
		dbpool.Dbs.Store("postgres host=foo-database port=5432 dbname=summon user=summon password='secretdbpass' sslmode=verify-full", db)
	})

	AfterEach(func() {
		db.Close()
		dbpool.Dbs.Delete("postgres host=foo-database port=5432 dbname=summon user=summon password='secretdbpass' sslmode=verify-full")

		// Check for any unmet expectations.
		err := dbMock.ExpectationsWereMet()
		if err != nil {
			Fail(fmt.Sprintf("there were unfulfilled database expectations: %s", err))
		}
	})

	It("creates a user", func() {
		instance.Spec = summonv1beta1.DjangoUserSpec{
			Email:          "foo@example.com",
			PasswordSecret: "foo-credentials",
			Database: summonv1beta1.DatabaseConnection{
				Host:     "foo-database",
				Port:     5432,
				Username: "summon",
				Database: "summon",
				PasswordSecretRef: summonv1beta1.SecretRef{
					Name: "summon.foo-database.credentials",
					Key:  "password",
				},
			},
		}
		dbSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "summon.foo-database.credentials", Namespace: "default"},
			Data: map[string][]byte{
				"password": []byte("secretdbpass"),
			},
		}
		userSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "foo-credentials", Namespace: "default"},
			Data: map[string][]byte{
				"password": []byte("djangopass"),
			},
		}
		ctx.Client = fake.NewFakeClient(dbSecret, userSecret)

		rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
		dbMock.ExpectQuery("INSERT INTO auth_user").WithArgs("foo@example.com", "djangopasshash", "", "", false, false, false).WillReturnRows(rows)

		comp := djangousercomponents.NewDatabase()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Status.Status).To(Equal(summonv1beta1.StatusReady))
		Expect(instance.Status.Message).To(Equal("User 1 created"))
	})
})
