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

	// "github.com/golang/glog"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"github.com/Ridecell/ridecell-operator/pkg/templates"
)

func (ctx *ComponentContext) GetTemplate(path string) (runtime.Object, error) {
	if ctx.reconciler.templates == nil {
		return nil, fmt.Errorf("no templates loaded for this reconciler")
	}
	return templates.Get(ctx.reconciler.templates, path, struct{ Instance runtime.Object }{Instance: ctx.Top})
}

func (ctx *ComponentContext) CreateOrUpdate(path string, mutateFn func(runtime.Object, runtime.Object) error) (reconcile.Result, controllerutil.OperationResult, error) {
	target, err := ctx.GetTemplate(path)
	if err != nil {
		return reconcile.Result{}, controllerutil.OperationResultNone, err
	}

	op, err := controllerutil.CreateOrUpdate(ctx.Context, ctx, target.DeepCopyObject(), func(existing runtime.Object) error {
		// Set owner ref.
		err := controllerutil.SetControllerReference(ctx.Top.(metav1.Object), existing.(metav1.Object), ctx.Scheme)
		if err != nil {
			return err
		}
		// Run the component-level mutator.
		err = mutateFn(target, existing)
		if err != nil {
			return err
		}
		// Sync the metadata fields.
		targetMeta := target.(metav1.ObjectMetaAccessor).GetObjectMeta().(*metav1.ObjectMeta)
		existingMeta := existing.(metav1.ObjectMetaAccessor).GetObjectMeta().(*metav1.ObjectMeta)
		return ReconcileMeta(targetMeta, existingMeta)
	})
	if err != nil {
		return reconcile.Result{Requeue: true}, op, err
	}

	return reconcile.Result{}, op, nil
}

// ComponentContext implements inject.Client.
// A client will be automatically injected.
var _ inject.Client = &ComponentContext{}

// InjectClient injects the client.
func (ctx *ComponentContext) InjectClient(c client.Client) error {
	ctx.Client = c
	return nil
}

// ComponentContext implements inject.Scheme.
// A scheme will be automatically injected.
var _ inject.Scheme = &ComponentContext{}

// InjectScheme injects the client.
func (ctx *ComponentContext) InjectScheme(s *runtime.Scheme) error {
	ctx.Scheme = s
	return nil
}