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
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	secretsv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/secrets/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type pullsecretcomponent struct{}

func NewSecret() *pullsecretcomponent {
	return &pullsecretcomponent{}
}

func (_ *pullsecretcomponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&corev1.Secret{},
	}
}

func (_ *pullsecretcomponent) IsReconcilable(_ *components.ComponentContext) bool {
	return true
}

func (comp *pullsecretcomponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*secretsv1beta1.PullSecret)

	existing := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: instance.Spec.PullSecret, Namespace: instance.Namespace}, existing)
	if err != nil {
	    instance.Status.Status = secretsv1beta1.StatusErrorSecretNotFound
		return reconcile.Result{Requeue: true}, fmt.Errorf("secret: unable to load secret: %v", err)
	} else if err == nil {
		// Loaded correctly, if the password exists then we're done.
		val, ok := existing.Data[".dockerconfigjson"]
		if !ok || !(len(val) > 0) {
		    instance.Status.Status = secretsv1beta1.StatusErrorSecretNotFound
		    return reconcile.Result{}, nil
		}
	}

	instance.Status.Status = secretsv1beta1.StatusReady
	return reconcile.Result{}, nil
}
