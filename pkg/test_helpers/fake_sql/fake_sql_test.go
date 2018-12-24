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

package fake_sql_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/Ridecell/ridecell-operator/pkg/test_helpers/fake_sql"
)

var _ = Describe("Fake SQL", func() {
	It("Can open a connection", func() {
		db := fake_sql.Open()
		Expect(db).NotTo(BeNil())
	})

	It("Can run some Execs", func() {
		db := fake_sql.Open()
		result, err := db.Exec("DROP TABLE students")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
		_, err = db.Exec("INSERT INTO bobby_tables $1", "foo")
		Expect(err).NotTo(HaveOccurred())
		Expect(result).NotTo(BeNil())
	})

	It("Can run a QueryRow", func() {
		db := fake_sql.Open()
		row := db.QueryRow("SELECT foo, bar as whatever, another FROM")
		var x string
		var y int
		var z bool
		err := row.Scan(&x, &y, &z)
		Expect(err).NotTo(HaveOccurred())
		Expect(x).To(Equal("0"))
		Expect(y).To(Equal(0))
		Expect(z).To(Equal(false))
	})
})
