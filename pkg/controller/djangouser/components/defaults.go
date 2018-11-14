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
	"strings"

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
	instance := ctx.Top.(*summonv1beta1.DjangoUser)

	// Fill in defaults.
	if instance.Spec.Username == "" {
		// Use the object name by default, replacing the first `.` with and `@` because `@` isn't allowed in object names.
		instance.Spec.Username = strings.Replace(instance.Name, ".", "@", 1)
	}
	if instance.Spec.PasswordSecret == "" {
		// Reverse the transform used for the default Username and add `-credentials`.
		instance.Spec.PasswordSecret = strings.Replace(instance.Spec.Username, "@", ".", 1) + "-credentials"
	}
	if instance.Spec.Database.PasswordSecretRef.Key == "" {
		// Use "password" as the default key.
		instance.Spec.Database.PasswordSecretRef.Key = "password"
	}

	return reconcile.Result{}, nil
}
