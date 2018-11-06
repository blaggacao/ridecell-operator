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

package pull_secret

import (
	"fmt"
	"os"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type pullSecretComponent struct{}

func New() *pullSecretComponent {
	return &pullSecretComponent{}
}

func (comp *pullSecretComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&corev1.Secret{},
	}
}

func (_ *pullSecretComponent) IsReconcilable(_ *components.ComponentContext) bool {
	// Secrets have no dependencies, always reconcile.
	return true
}

func (comp *pullSecretComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	operatorNamespace := os.Getenv("NAMESPACE")
	if operatorNamespace == "" {
		instance.Status.PullSecretStatus = summonv1beta1.StatusError
		return reconcile.Result{}, fmt.Errorf("$NAMESPACE is not set")
	}

	target := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: *instance.Spec.PullSecret, Namespace: operatorNamespace}, target)
	if err != nil {
		if errors.IsNotFound(err) {
			instance.Status.PullSecretStatus = summonv1beta1.StatusErrorSecretNotFound
		} else {
			instance.Status.PullSecretStatus = summonv1beta1.StatusError
		}
		return reconcile.Result{Requeue: true}, err
	}

	_, err = controllerutil.CreateOrUpdate(ctx.Context, ctx, target.DeepCopyObject(), func(existingObj runtime.Object) error {
		existing := existingObj.(*corev1.Secret)
		// Set owner ref.
		err := controllerutil.SetControllerReference(instance, existing, ctx.Scheme)
		if err != nil {
			instance.Status.PullSecretStatus = summonv1beta1.StatusError
			return err
		}
		// Sync important fields.
		existing.ObjectMeta.Labels = target.ObjectMeta.Labels
		existing.ObjectMeta.Annotations = target.ObjectMeta.Annotations
		existing.Type = target.Type
		existing.Data = target.Data
		return nil
	})
	if err != nil {
		instance.Status.PullSecretStatus = summonv1beta1.StatusError
		return reconcile.Result{Requeue: true}, err
	}

	instance.Status.PullSecretStatus = summonv1beta1.StatusReady
	return reconcile.Result{}, nil
}
