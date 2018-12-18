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
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/pkg/errors"
)

type configmapComponent struct {
	templatePath string
}

func NewConfigMap(templatePath string) *configmapComponent {
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
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	// Create the map that will be the summon-platform.yml
	config := map[string]interface{}{}
	for key, value := range instance.Spec.Config {
		config[key] = value.ToNilInterface()
	}

	// Render to JSON (which is a subset of YAML).
	b, err := json.Marshal(config)
	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "configmap: unable to serialize config JSON for %s/%s", instance.Namespace, instance.Name)
	}

	// Set up the extra data map for the template.
	extra := map[string]interface{}{}
	extra["SummonYaml"] = string(b)

	res, _, err := ctx.CreateOrUpdate(comp.templatePath, extra, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*corev1.ConfigMap)
		existing := existingObj.(*corev1.ConfigMap)
		// Copy the data over.
		existing.Data = goal.Data
		return nil
	})
	return res, err
}
