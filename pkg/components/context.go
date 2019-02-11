/*
Copyright 2018-2019 Ridecell, Inc.

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
	"net/http"

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/runtime/inject"

	"github.com/blaggacao/ridecell-operator/pkg/templates"
)

func (ctx *ComponentContext) GetTemplate(path string, extraData map[string]interface{}) (runtime.Object, error) {
	if ctx.templates == nil {
		err := fmt.Errorf("no templates loaded for this reconciler")
		ctx.Logger.Error(err, "no templates loaded for this reconciler")
		return nil, err
	}
	return templates.Get(ctx.templates, path, struct {
		Instance runtime.Object
		Extra    map[string]interface{}
	}{Instance: ctx.Top, Extra: extraData})
}

func (ctx *ComponentContext) CreateOrUpdate(path string, extraData map[string]interface{}, mutateFn func(runtime.Object, runtime.Object) error) (Result, controllerutil.OperationResult, error) {
	target, err := ctx.GetTemplate(path, extraData)
	if err != nil {
		return Result{}, controllerutil.OperationResultNone, err
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
		return Result{Requeue: true}, op, err
	}

	return Result{}, op, nil
}

func (ctx *ComponentContext) UpdateTopMeta(mutateFn func(*metav1.ObjectMeta) error) (Result, controllerutil.OperationResult, error) {
	target := ctx.Top
	op, err := controllerutil.CreateOrUpdate(ctx.Context, ctx, target.DeepCopyObject(), func(existing runtime.Object) error {
		// Sync the metadata fields.
		targetMeta := target.(metav1.ObjectMetaAccessor).GetObjectMeta().(*metav1.ObjectMeta)
		existingMeta := existing.(metav1.ObjectMetaAccessor).GetObjectMeta().(*metav1.ObjectMeta)
		// Run the matadata mutator.
		err := mutateFn(targetMeta)
		if err != nil {
			return err
		}
		return ReconcileMeta(targetMeta, existingMeta)
	})
	if err != nil {
		return Result{Requeue: true}, op, err
	}

	ctx.Logger.V(1).Info("updated top object metadata", "object", target, "operation", op)
	return Result{}, op, nil
}

func (ctx *ComponentContext) GetOne(obj runtime.Object, labels map[string]string) (Result, runtime.Object, error) {
	accessor := meta.NewAccessor()
	namespace, err := accessor.Namespace(ctx.Top)
	if err != nil {
		return Result{}, nil, err
	}

	listoptions := client.InNamespace(namespace)
	listoptions.MatchingLabels(labels)
	err = ctx.List(ctx.Context, listoptions, obj)
	if err != nil {
		return Result{}, nil, err
	}
	gvk, err := apiutil.GVKForObject(obj, ctx.Scheme)
	if err != nil {
		// What?
		return Result{}, nil, err
	}
	items, err := meta.ExtractList(obj)
	if len(items) > 1 {
		err := fmt.Errorf("more than one match found")
		ctx.Logger.Error(err, "failed", "labels", labels, "group", gvk.Group, "kind", gvk.Kind)
		return Result{}, nil, err
	} else if len(items) < 1 {
		err := fmt.Errorf("no match found")
		ctx.Logger.Error(err, "labels", labels, "group", gvk.Group, "kind", gvk.Kind)
		return Result{Requeue: true}, nil, err
	}
	return Result{}, items[0], nil

}

// Method for creating a test context, for use in component unit tests.
func NewTestContext(top runtime.Object, templates http.FileSystem) *ComponentContext {
	// This method is ugly and I don't like it. I should rebuild this whole subsytem around interfaces and have an explicit fake for it.
	return &ComponentContext{
		Top:       top,
		Client:    fake.NewFakeClient(top),
		Scheme:    scheme.Scheme,
		templates: templates,
	}
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
