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
	"context"
	"log"
	"net/http"

	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	"github.com/Ridecell/ridecell-operator/pkg/templates"
)

func NewController(mgr manager.Manager, top runtime.Object, templates http.FileSystem, components []Component) *ComponentController {
	return &ComponentController{
		Client:     mgr.GetClient(),
		Scheme:     mgr.GetScheme(),
		Top:        top,
		Templates:  templates,
		Components: components,
	}
}

func (controller *ComponentController) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	top := controller.Top.DeepCopyObject()
	err := controller.Get(context.TODO(), request.NamespacedName, top)
	if err != nil {
		if errors.IsNotFound(err) {
			// Top object not found, likely already deleted.
			return reconcile.Result{}, nil
		}
		// Some other fetch error, try again on the next tick.
		return reconcile.Result{Requeue: true}, err
	}

	ctx := &ComponentContext{
		ComponentController: controller,
		Context:             context.TODO(),
		Top:                 top,
	}

	result, err := ReconcileComponents(ctx, controller.Components)
	if err != nil {
		log.Printf("ERROR! %s\n", err.Error())
	}
	return result, err
}

func (controller *ComponentController) WatchTypes() []runtime.Object {
	types := []runtime.Object{}
	for _, component := range controller.Components {
		types = append(types, component.WatchTypes()...)
	}
	return types
}

func ReconcileComponents(ctx *ComponentContext, components []Component) (reconcile.Result, error) {
	ready := []Component{}
	for _, component := range components {
		if component.IsReconcilable(ctx) {
			ready = append(ready, component)
		}
	}
	res := reconcile.Result{}
	for _, component := range ready {
		innerRes, err := component.Reconcile(ctx)
		// Update result. This should be checked before the err!=nil because sometimes
		// we want to requeue immediately on error.
		if innerRes.Requeue {
			res.Requeue = true
		}
		if innerRes.RequeueAfter != 0 && (res.RequeueAfter == 0 || res.RequeueAfter > innerRes.RequeueAfter) {
			res.RequeueAfter = innerRes.RequeueAfter
		}
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func ReconcileMeta(target, existing *metav1.ObjectMeta) error {
	if target.Labels != nil {
		if existing.Labels == nil {
			existing.Labels = map[string]string{}
		}
		for k, v := range target.Labels {
			existing.Labels[k] = v
		}
	}
	if target.Annotations != nil {
		if existing.Annotations == nil {
			existing.Annotations = map[string]string{}
		}
		for k, v := range target.Annotations {
			existing.Annotations[k] = v
		}
	}
	return nil
}

func (ctx *ComponentContext) GetTemplate(path string) (runtime.Object, error) {
	return templates.Get(ctx.Templates, path, struct{ Instance runtime.Object }{Instance: ctx.Top})
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
