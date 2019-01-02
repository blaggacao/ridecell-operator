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
	"fmt"

	"github.com/pkg/errors"
	appsv1 "k8s.io/api/apps/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type statusComponent struct{}

func NewStatus() *statusComponent {
	return &statusComponent{}
}

func (comp *statusComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *statusComponent) IsReconcilable(_ *components.ComponentContext) bool {
	// Always ready, always waiting ...
	return true
}

func (comp *statusComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	if instance.Status.Status != summonv1beta1.StatusDeploying {
		// If the migrations component didn't already set us to Deploying, don't even bother checking.
		return reconcile.Result{}, nil
	}

	// Grab all (important) Deployments and make sure they are all ready.
	web := &appsv1.Deployment{}
	daphne := &appsv1.Deployment{}
	celeryd := &appsv1.Deployment{}
	channelworker := &appsv1.Deployment{}
	static := &appsv1.Deployment{}
	celerybeat := &appsv1.StatefulSet{}

	// Go's lack of generics can fuck right off.
	err := comp.get(ctx, "web", web)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = comp.get(ctx, "daphne", daphne)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = comp.get(ctx, "celeryd", celeryd)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = comp.get(ctx, "channelworker", channelworker)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = comp.get(ctx, "static", static)
	if err != nil {
		return reconcile.Result{}, err
	}
	err = comp.get(ctx, "celerybeat", celerybeat)
	if err != nil {
		return reconcile.Result{}, err
	}

	// The big check!
	if web.Spec.Replicas != nil && web.Status.AvailableReplicas == *web.Spec.Replicas &&
		daphne.Spec.Replicas != nil && daphne.Status.AvailableReplicas == *daphne.Spec.Replicas &&
		celeryd.Spec.Replicas != nil && celeryd.Status.AvailableReplicas == *celeryd.Spec.Replicas &&
		channelworker.Spec.Replicas != nil && channelworker.Status.AvailableReplicas == *channelworker.Spec.Replicas &&
		static.Spec.Replicas != nil && static.Status.AvailableReplicas == *static.Spec.Replicas &&
		// Note this one is different, available vs ready.
		celerybeat.Spec.Replicas != nil && celerybeat.Status.ReadyReplicas == *celerybeat.Spec.Replicas {
		// TODO: Add an actual HTTP self check in here.
		instance.Status.Status = summonv1beta1.StatusReady
		instance.Status.Message = fmt.Sprintf("Cluster %s ready", instance.Name)
	}

	return reconcile.Result{}, nil
}

// Short helper because we need to do this 6 times.
func (comp *statusComponent) get(ctx *components.ComponentContext, part string, obj runtime.Object) error {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	name := types.NamespacedName{Name: fmt.Sprintf("%s-%s", instance.Name, part), Namespace: instance.Namespace}
	err := ctx.Get(ctx.Context, name, obj)
	// If it's a NotFound error, just ignore it since we don't want to that to fail things and the zero value will fail later on.
	if err != nil && !kerrors.IsNotFound(err) {
		return errors.Wrapf(err, "status: unable to get Deployment or StatefulSet %s for %s subsystem", name, part)
	}
	return nil
}
