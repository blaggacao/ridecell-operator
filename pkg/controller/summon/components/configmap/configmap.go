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

package configmap

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type configmapComponent struct {
	templatePath string
}

func New(templatePath string) *configmapComponent {
	return &configmapComponent{templatePath: templatePath}
}

func (comp *configmapComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&corev1.ConfigMap{},
	}
}

func (_ *configmapComponent) IsReconcilable(_ *components.ComponentContext) bool {
	// ConfigMaps have no dependencies, always reconcile.
	return true
}

func (comp *configmapComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	res, _, err := ctx.CreateOrUpdate(comp.templatePath, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*corev1.ConfigMap)
		existing := existingObj.(*corev1.ConfigMap)
		// Copy the data over.
		existing.Data = goal.Data
		return nil
	})
	return res, err
}
