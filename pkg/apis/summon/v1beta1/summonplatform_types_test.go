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

package v1beta1_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"golang.org/x/net/context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
)

var _ = Describe("SummonPlatform types", func() {
	var helpers *test_helpers.PerTestHelpers

	BeforeEach(func() {
		helpers = testHelpers.SetupTest()
	})

	AfterEach(func() {
		helpers.TeardownTest()
	})

	It("can create a SummonPlatform object", func() {
		c := helpers.Client
		key := types.NamespacedName{
			Name:      "foo",
			Namespace: helpers.Namespace,
		}
		created := &summonv1beta1.SummonPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: helpers.Namespace,
			}}
		fetched := &summonv1beta1.SummonPlatform{}
		err := c.Create(context.TODO(), created)
		Expect(err).NotTo(HaveOccurred())

		err = c.Get(context.TODO(), key, fetched)
		Expect(err).NotTo(HaveOccurred())
		Expect(fetched.Spec).To(Equal(created.Spec))
	})

	It("can update a SummonPlatform object", func() {
		c := helpers.Client

		key := types.NamespacedName{
			Name:      "foo",
			Namespace: helpers.Namespace,
		}
		created := &summonv1beta1.SummonPlatform{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: helpers.Namespace,
			}}
		fetched := &summonv1beta1.SummonPlatform{}
		err := c.Create(context.TODO(), created)
		Expect(err).NotTo(HaveOccurred())

		created.Labels = map[string]string{"hello": "world"}
		err = c.Update(context.TODO(), created)
		Expect(err).NotTo(HaveOccurred())

		err = c.Get(context.TODO(), key, fetched)
		Expect(err).NotTo(HaveOccurred())
		Expect(fetched.Labels).To(Equal(created.Labels))
	})
})

// func TestStorageSummon(t *testing.T) {
// 	key := types.NamespacedName{
// 		Name:      "foo",
// 		Namespace: "default",
// 	}
// 	created := &Summon{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      "foo",
// 			Namespace: "default",
// 		}}
// 	g := gomega.NewGomegaWithT(t)

// 	// Test Create
// 	fetched := &Summon{}
// 	g.Expect(c.Create(context.TODO(), created)).NotTo(gomega.HaveOccurred())

// 	g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
// 	g.Expect(fetched).To(gomega.Equal(created))

// 	// Test Updating the Labels
// updated := fetched.DeepCopy()
// updated.Labels = map[string]string{"hello": "world"}
// g.Expect(c.Update(context.TODO(), updated)).NotTo(gomega.HaveOccurred())

// g.Expect(c.Get(context.TODO(), key, fetched)).NotTo(gomega.HaveOccurred())
// g.Expect(fetched).To(gomega.Equal(updated))

// 	// Test Delete
// 	g.Expect(c.Delete(context.TODO(), fetched)).NotTo(gomega.HaveOccurred())
// 	g.Expect(c.Get(context.TODO(), key, fetched)).To(gomega.HaveOccurred())
// }
