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

package s3bucket_test

import (
	"time"

	. "github.com/onsi/ginkgo"
	//. "github.com/onsi/gomega"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/types"

	//s3bucketv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/aws/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
)

const timeout = time.Second * 20

var _ = Describe("s3bucket controller", func() {
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
