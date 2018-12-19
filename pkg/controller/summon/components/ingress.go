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
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type ingressComponent struct {
	templatePath string
}

func NewIngress(templatePath string) *ingressComponent {
	return &ingressComponent{templatePath: templatePath}
}

func (comp *ingressComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&extv1beta1.Ingress{},
	}
}

func (_ *ingressComponent) IsReconcilable(_ *components.ComponentContext) bool {
	return true
}

func (comp *ingressComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	res, _, err := ctx.CreateOrUpdate(comp.templatePath, nil, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*extv1beta1.Ingress)
		existing := existingObj.(*extv1beta1.Ingress)
		// Copy the Spec over.
		existing.Spec = goal.Spec
		return nil
	})
	return res, err
}
