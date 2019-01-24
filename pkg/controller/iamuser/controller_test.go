/*
Copyright 2019 Ridecell, Inc.

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

package iamuser_test

import (
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
	. "github.com/onsi/ginkgo"
)

var _ = Describe("iamuser controller", func() {
	var helpers *test_helpers.PerTestHelpers

	BeforeEach(func() {
		helpers = testHelpers.SetupTest()
	})

	AfterEach(func() {
		helpers.TeardownTest()
	})

	It("runs a basic reconcile", func() {
		// We do not currently have a good way to test this controller
		// This is being left blank until we do
	})
})
