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

package service

import (
	// "context"
	"fmt"
	// "net/http"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	// "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/Ridecell/ridecell-operator/pkg/templates"
)

type serviceComponent struct {
	templatePath string
}

func New(templatePath string) *serviceComponent {
	return &serviceComponent{templatePath: templatePath}
}

func (_ *serviceComponent) IsReconcilable(_ *components.ComponentContext) bool {
	// Services have no dependencies, always reconcile.
	return true
}

func (comp *serviceComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	summon := ctx.Top.(*summonv1beta1.SummonPlatform)

	serviceObj, err := templates.Get(ctx.Templates, comp.templatePath, struct{ Instance *summonv1beta1.SummonPlatform }{Instance: summon})
	if err != nil {
		return reconcile.Result{}, err
	}

	service := serviceObj.(*corev1.Service)

	_, err = controllerutil.CreateOrUpdate(ctx.Context, ctx, service.DeepCopyObject(), func(existingObj runtime.Object) error {
		existing := existingObj.(*corev1.Service)
		// Set owner ref.
		err := controllerutil.SetControllerReference(summon, existing, ctx.Scheme)
		if err != nil {
			return err
		}
		// Special case: Services mutate the ClusterIP value in the Spec and it should be preserved.
		service.Spec.ClusterIP = existing.Spec.ClusterIP
		// Copy the Spec over.
		existing.Spec = service.Spec
		// Sync the metadata.
		components.ReconcileMeta(&service.ObjectMeta, &existing.ObjectMeta)

		return nil
	})
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	return reconcile.Result{}, nil
}
