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
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
)

type PostgresOperatorDatabaseComponent struct{}

func NewPostgresOperatorDB() *PostgresOperatorDatabaseComponent {
	return &PostgresOperatorDatabaseComponent{}
}

func (comp *PostgresOperatorDatabaseComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *PostgresOperatorDatabaseComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	return true
}

func (comp *PostgresOperatorDatabaseComponent) Reconcile(ctx *components.ComponentContext) (components.Result, error) {
	instance := ctx.Top.(*dbv1beta1.PostgresOperatorDatabase)
	fetchDatabase := &postgresv1.Postgresql{}
	err := ctx.Client.Get(ctx.Context, types.NamespacedName{Name: instance.Spec.DatabaseRef.Name, Namespace: instance.Namespace}, fetchDatabase)
	if err != nil {
		return components.Result{}, errors.Wrapf(err, "postgres_operatordb: Unable to get specified database")
	}

	existingRefs := fetchDatabase.GetOwnerReferences()
	if len(existingRefs) > 0 {
		return components.Result{}, errors.Errorf("postgres_operatordb: Object has owners %#v", existingRefs)
	}

	// Update Users and Databases of Postgresql object
	fetchDatabase.Spec.Users[instance.Spec.Database] = postgresv1.UserFlags{}
	fetchDatabase.Spec.Databases[instance.Spec.Database] = instance.Spec.Database

	err = ctx.Update(ctx.Context, fetchDatabase)
	if err != nil {
		return components.Result{}, errors.Wrapf(err, "postgres_operatordb: Failed to update Postgresql object")
	}

	return components.Result{StatusModifier: func(obj runtime.Object) error {
		instance := obj.(*dbv1beta1.PostgresOperatorDatabase)
		instance.Status.Status = dbv1beta1.StatusReady
		instance.Status.Message = "Ready"
		return nil
	}}, nil
}
