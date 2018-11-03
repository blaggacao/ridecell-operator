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

package migrations

import (
	"github.com/golang/glog"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type migrationComponent struct {
	templatePath string
}

func New(templatePath string) *migrationComponent {
	return &migrationComponent{templatePath: templatePath}
}

func (comp *migrationComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&batchv1.Job{},
	}
}

func (_ *migrationComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	if instance.Status.PostgresStatus == nil || *instance.Status.PostgresStatus != postgresv1.ClusterStatusRunning {
		// Database not ready yet.
		return false
	}
	if instance.Spec.Version == instance.Status.MigrateVersion {
		// Already migrated.
		return false
	}
	return true
}

func (comp *migrationComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	obj, err := ctx.GetTemplate(comp.templatePath)
	if err != nil {
		return reconcile.Result{}, err
	}
	job := obj.(*batchv1.Job)
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	existing := &batchv1.Job{}
	err = ctx.Get(ctx.Context, types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, existing)
	if err != nil && errors.IsNotFound(err) {
		glog.Infof("Creating migration Job %s/%s\n", job.Namespace, job.Name)
		err = controllerutil.SetControllerReference(instance, job, ctx.Scheme)
		if err != nil {
			return reconcile.Result{}, err
		}
		err = ctx.Create(ctx.Context, job)
		if err != nil {
			// If this fails, someone else might have started one between the Get and here, so just try again.
			return reconcile.Result{Requeue: true}, err
		}
		// Done for now.
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Check if the job succeeded.
	if existing.Status.Succeeded > 0 {
		// Success! Update the MigrateVersion (this will trigger a reconcile) and delete the job.
		instance.Status.MigrateVersion = instance.Spec.Version
		glog.Infof("Migration job %s/%s succeeded, updating MigrateVersion to %s\n", job.Namespace, job.Name, instance.Status.MigrateVersion)

		glog.V(2).Infof("Deleting migration Job %s/%s\n", existing.Namespace, existing.Name)
		err = ctx.Delete(ctx.Context, existing, client.PropagationPolicy(metav1.DeletePropagationBackground))
		if err != nil {
			return reconcile.Result{Requeue: true}, err
		}
	}

	// If the job failed, leave things for debugging.
	if existing.Status.Failed > 0 {
		glog.Errorf("Migration job for %s/%s failed, leaving job %s/%s running for debugging purposes\n", instance.Namespace, instance.Name, existing.Namespace, existing.Name)
	}

	return reconcile.Result{}, nil
}
