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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	. "github.com/onsi/gomega/gstruct"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
)

var _ = Describe("SummonPlatform Defaults Component", func() {
	It("does nothing on a filled out object", func() {
		replicas := int32(2)
		instance := &summonv1beta1.SummonPlatform{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Spec: summonv1beta1.SummonPlatformSpec{
				Hostname:              "foo.example.com",
				PullSecret:            "foo-secret",
				WebReplicas:           &replicas,
				DaphneReplicas:        &replicas,
				ChannelWorkerReplicas: &replicas,
				StaticReplicas:        &replicas,
			},
		}
		ctx := &components.ComponentContext{Top: instance}

		comp := summoncomponents.NewDefaults()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Spec.Hostname).To(Equal("foo.example.com"))
		Expect(instance.Spec.PullSecret).To(Equal("foo-secret"))
		Expect(instance.Spec.WebReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.DaphneReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.ChannelWorkerReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.StaticReplicas).To(PointTo(BeEquivalentTo(2)))
	})

	It("sets a default hostname", func() {
		replicas := int32(2)
		instance := &summonv1beta1.SummonPlatform{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Spec: summonv1beta1.SummonPlatformSpec{
				PullSecret:            "foo-secret",
				WebReplicas:           &replicas,
				DaphneReplicas:        &replicas,
				ChannelWorkerReplicas: &replicas,
				StaticReplicas:        &replicas,
			},
		}
		ctx := &components.ComponentContext{Top: instance}

		comp := summoncomponents.NewDefaults()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Spec.Hostname).To(Equal("foo.ridecell.us"))
		Expect(instance.Spec.PullSecret).To(Equal("foo-secret"))
		Expect(instance.Spec.WebReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.DaphneReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.ChannelWorkerReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.StaticReplicas).To(PointTo(BeEquivalentTo(2)))
	})

	It("sets a default pull secret", func() {
		replicas := int32(2)
		instance := &summonv1beta1.SummonPlatform{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Spec: summonv1beta1.SummonPlatformSpec{
				Hostname:              "foo.example.com",
				WebReplicas:           &replicas,
				DaphneReplicas:        &replicas,
				ChannelWorkerReplicas: &replicas,
				StaticReplicas:        &replicas,
			},
		}
		ctx := &components.ComponentContext{Top: instance}

		comp := summoncomponents.NewDefaults()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Spec.Hostname).To(Equal("foo.example.com"))
		Expect(instance.Spec.PullSecret).To(Equal("pull-secret"))
		Expect(instance.Spec.WebReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.DaphneReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.ChannelWorkerReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.StaticReplicas).To(PointTo(BeEquivalentTo(2)))
	})

	It("sets a default web replicas", func() {
		replicas := int32(2)
		instance := &summonv1beta1.SummonPlatform{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Spec: summonv1beta1.SummonPlatformSpec{
				Hostname:              "foo.example.com",
				PullSecret:            "foo-secret",
				DaphneReplicas:        &replicas,
				ChannelWorkerReplicas: &replicas,
				StaticReplicas:        &replicas,
			},
		}
		ctx := &components.ComponentContext{Top: instance}

		comp := summoncomponents.NewDefaults()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Spec.Hostname).To(Equal("foo.example.com"))
		Expect(instance.Spec.PullSecret).To(Equal("foo-secret"))
		Expect(instance.Spec.WebReplicas).To(PointTo(BeEquivalentTo(1)))
		Expect(instance.Spec.DaphneReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.ChannelWorkerReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.StaticReplicas).To(PointTo(BeEquivalentTo(2)))
	})

	It("allows 0 web replicas", func() {
		replicas := int32(2)
		zeroReplicas := int32(0)
		instance := &summonv1beta1.SummonPlatform{
			ObjectMeta: metav1.ObjectMeta{Name: "foo"},
			Spec: summonv1beta1.SummonPlatformSpec{
				Hostname:              "foo.example.com",
				PullSecret:            "foo-secret",
				WebReplicas:           &zeroReplicas,
				DaphneReplicas:        &replicas,
				ChannelWorkerReplicas: &replicas,
				StaticReplicas:        &replicas,
			},
		}
		ctx := &components.ComponentContext{Top: instance}

		comp := summoncomponents.NewDefaults()
		_, err := comp.Reconcile(ctx)
		Expect(err).NotTo(HaveOccurred())
		Expect(instance.Spec.Hostname).To(Equal("foo.example.com"))
		Expect(instance.Spec.PullSecret).To(Equal("foo-secret"))
		Expect(instance.Spec.WebReplicas).To(PointTo(BeEquivalentTo(0)))
		Expect(instance.Spec.DaphneReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.ChannelWorkerReplicas).To(PointTo(BeEquivalentTo(2)))
		Expect(instance.Spec.StaticReplicas).To(PointTo(BeEquivalentTo(2)))
	})
})
