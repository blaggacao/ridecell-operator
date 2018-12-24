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

// TODO: This whole thing should probably be its own custom resource.

import (
	_ "github.com/lib/pq"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type superuserComponent struct{}

func NewSuperuser() *superuserComponent {
	return &superuserComponent{}
}

func (comp *superuserComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&summonv1beta1.DjangoUser{},
	}
}

func (comp *superuserComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	// Wait for the database to be up and migrated.
	if instance.Status.PostgresStatus != postgresv1.ClusterStatusRunning {
		return false
	}
	if instance.Status.PostgresExtensionStatus != summonv1beta1.StatusReady {
		return false
	}
	if instance.Status.MigrateVersion != instance.Spec.Version {
		return false
	}
	return true
}

func (comp *superuserComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	res, _, err := ctx.CreateOrUpdate("superuser.yml.tpl", nil, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*summonv1beta1.DjangoUser)
		existing := existingObj.(*summonv1beta1.DjangoUser)
		// Copy the Spec over.
		existing.Spec = goal.Spec
		return nil
	})
	return res, err
}
