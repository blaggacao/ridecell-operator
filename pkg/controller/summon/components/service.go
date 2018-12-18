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

package components

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type serviceComponent struct {
	templatePath string
}

func NewService(templatePath string) *serviceComponent {
	return &serviceComponent{templatePath: templatePath}
}

func (comp *serviceComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&corev1.Service{},
	}
}

func (_ *serviceComponent) IsReconcilable(_ *components.ComponentContext) bool {
	// Services have no dependencies, always reconcile.
	return true
}

func (comp *serviceComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	res, _, err := ctx.CreateOrUpdate(comp.templatePath, nil, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*corev1.Service)
		existing := existingObj.(*corev1.Service)
		// Special case: Services mutate the ClusterIP value in the Spec and it should be preserved.
		goal.Spec.ClusterIP = existing.Spec.ClusterIP
		// Copy the Spec over.
		existing.Spec = goal.Spec
		return nil
	})
	return res, err
}
