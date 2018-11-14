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

// import (
// 	"context"
// 	"log"
// 	"net/http"
// 	"reflect"

// 	"github.com/golang/glog"
// 	"k8s.io/apimachinery/pkg/api/errors"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/runtime"
// 	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
// 	"sigs.k8s.io/controller-runtime/pkg/manager"
// 	"sigs.k8s.io/controller-runtime/pkg/reconcile"

// 	"github.com/Ridecell/ridecell-operator/pkg/templates"
// )

// func NewController(mgr manager.Manager, top runtime.Object, templates http.FileSystem, components []Component) *ComponentController {
// 	return &ComponentController{
// 		Client:     mgr.GetClient(),
// 		Scheme:     mgr.GetScheme(),
// 		Top:        top,
// 		Templates:  templates,
// 		Components: components,
// 	}
// }

// func (controller *ComponentController) Reconcile(request reconcile.Request) (reconcile.Result, error) {
// 	top := controller.Top.DeepCopyObject()
// 	err := controller.Get(context.TODO(), request.NamespacedName, top)
// 	if err != nil {
// 		if errors.IsNotFound(err) {
// 			// Top object not found, likely already deleted.
// 			return reconcile.Result{}, nil
// 		}
// 		// Some other fetch error, try again on the next tick.
// 		return reconcile.Result{Requeue: true}, err
// 	}

// 	originalTop := top.DeepCopyObject()
// 	ctx := &ComponentContext{
// 		ComponentController: controller,
// 		Context:             context.TODO(),
// 		Top:                 top,
// 	}

// 	result, err := ReconcileComponents(ctx, controller.Components)
// 	if err != nil {
// 		log.Printf("ERROR! %s\n", err.Error())
// 		top.(Statuser).SetErrorStatus(err.Error())
// 	}
// 	if !reflect.DeepEqual(top.(Statuser).GetStatus(), originalTop.(Statuser).GetStatus()) {
// 		// Update the top object status.
// 		log.Printf("Updating status\n")
// 		err = controller.Status().Update(ctx.Context, top)
// 		if err != nil {
// 			// Something went wrong, we definitely want to rerun, unless ...
// 			oldRequeue := result.Requeue
// 			result.Requeue = true
// 			if errors.IsNotFound(err) {
// 				// Older Kubernetes which doesn't support status subobjects, so use a GET+UPDATE
// 				// because the controller-runtime client doesn't support PATCH calls.
// 				freshTop := controller.Top.DeepCopyObject()
// 				err = controller.Get(ctx.Context, request.NamespacedName, freshTop)
// 				if err != nil {
// 					// What?
// 					return result, err
// 				}
// 				freshTop.(Statuser).SetStatus(top.(Statuser).GetStatus())
// 				err = controller.Update(ctx.Context, freshTop)
// 				if err != nil {
// 					// Update failed, probably another update got there first.
// 					return result, err
// 				} else {
// 					// Update worked, so no error for the final return.
// 					result.Requeue = oldRequeue
// 					err = nil
// 				}
// 			}
// 		}
// 	}
// 	return result, err
// }

// func (controller *ComponentController) WatchTypes() []runtime.Object {
// 	types := []runtime.Object{}
// 	for _, component := range controller.Components {
// 		types = append(types, component.WatchTypes()...)
// 	}
// 	return types
// }

// func ReconcileComponents(ctx *ComponentContext, components []Component) (reconcile.Result, error) {
// 	instance := ctx.Top.(metav1.Object)
// 	ready := []Component{}
// 	for _, component := range components {
// 		glog.V(10).Infof("[%s/%s] ReconcileComponents: Checking if %#v is available to reconcile", instance.GetNamespace(), instance.GetName(), component)
// 		if component.IsReconcilable(ctx) {
// 			glog.V(9).Infof("[%s/%s] ReconcileComponents: %#v is available to reconcile", instance.GetNamespace(), instance.GetName(), component)
// 			ready = append(ready, component)
// 		}
// 	}
// 	res := reconcile.Result{}
// 	for _, component := range ready {
// 		innerRes, err := component.Reconcile(ctx)
// 		// Update result. This should be checked before the err!=nil because sometimes
// 		// we want to requeue immediately on error.
// 		if innerRes.Requeue {
// 			res.Requeue = true
// 		}
// 		if innerRes.RequeueAfter != 0 && (res.RequeueAfter == 0 || res.RequeueAfter > innerRes.RequeueAfter) {
// 			res.RequeueAfter = innerRes.RequeueAfter
// 		}
// 		if err != nil {
// 			return res, err
// 		}
// 	}
// 	return res, nil
// }
