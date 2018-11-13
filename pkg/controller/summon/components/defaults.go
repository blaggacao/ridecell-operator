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

package components

import (
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type defaultsComponent struct {
}

func NewDefaults() *defaultsComponent {
	return &defaultsComponent{}
}

func (_ *defaultsComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *defaultsComponent) IsReconcilable(_ *components.ComponentContext) bool {
	return true
}

func (comp *defaultsComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	// Fill in defaults.
	if instance.Spec.Hostname == "" {
		instance.Spec.Hostname = instance.Name + ".ridecell.us"
	}
	if instance.Spec.PullSecret == "" {
		instance.Spec.PullSecret = "pull-secret"
	}
	defaultReplicas := int32(1)
	if instance.Spec.WebReplicas == nil {
		instance.Spec.WebReplicas = &defaultReplicas
	}
	if instance.Spec.DaphneReplicas == nil {
		instance.Spec.DaphneReplicas = &defaultReplicas
	}
	if instance.Spec.WorkerReplicas == nil {
		instance.Spec.WorkerReplicas = &defaultReplicas
	}
	if instance.Spec.ChannelWorkerReplicas == nil {
		instance.Spec.ChannelWorkerReplicas = &defaultReplicas
	}
	if instance.Spec.StaticReplicas == nil {
		instance.Spec.StaticReplicas = &defaultReplicas
	}

	return reconcile.Result{}, nil
}
