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

package postgresextension_test

import (
	"database/sql"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	"github.com/Ridecell/ridecell-operator/pkg/dbpool"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
)

const timeout = time.Second * 5

var _ = Describe("PostgresExtension controller", func() {
	var helpers *test_helpers.PerTestHelpers
	var dbMock sqlmock.Sqlmock
	var db *sql.DB

	BeforeEach(func() {
		helpers = testHelpers.SetupTest()

		var err error
		db, dbMock, err = sqlmock.New()
		Expect(err).NotTo(HaveOccurred())
		dbpool.Dbs.Store("postgres host=foo-database port=5432 dbname=summon user=summon password='secretdbpass' sslmode=require", db)
	})

	AfterEach(func() {
		helpers.TeardownTest()
		db.Close()
		dbpool.Dbs.Delete("postgres host=foo-database port=5432 dbname=summon user=summon password='secretdbpass' sslmode=require")

		// Check for any unmet expectations.
		err := dbMock.ExpectationsWereMet()
		if err != nil {
			Fail(fmt.Sprintf("there were unfulfilled database expectations: %s", err))
		}
	})

	It("runs a basic reconcile", func() {
	})
})
