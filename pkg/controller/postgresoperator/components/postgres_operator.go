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
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
)

type PostgresOperatorComponent struct{}

func NewOperatorDatabase() *PostgresOperatorComponent {
	return &PostgresOperatorComponent{}
}

func (comp *PostgresOperatorComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *PostgresOperatorComponent) IsReconcilable(_ *components.ComponentContext) bool {
	return true
}

func (comp *PostgresOperatorComponent) Reconcile(ctx *components.ComponentContext) (components.Result, error) {
	instance := ctx.Top.(*dbv1beta1.PostgresOperator)
	fetchDatabase := &postgresv1.Postgresql{}
	err := ctx.Client.Get(ctx.Context, types.NamespacedName{Name: instance.Spec.DatabaseRef.Name, Namespace: instance.Spec.DatabaseRef.Namespace}, fetchDatabase)
	if err != nil {
		return components.Result{StatusModifier: func(obj runtime.Object) error {
			instance := obj.(*dbv1beta1.PostgresOperator)
			instance.Status.Status = dbv1beta1.StatusError
			instance.Status.Message = fmt.Sprintf("postgres_operator: Failed to get specified Postgresql object Name: %s, Namespace %s", instance.Spec.DatabaseRef.Name, instance.Spec.DatabaseRef.Namespace)
			return nil
		}}, errors.Wrapf(err, "postgres_operator: Unable to get specified database")
	}

	for user, userFlags := range instance.Spec.Users {
		existingUserFlags, ok := fetchDatabase.Spec.Users[user]
		if !ok {
			fetchDatabase.Spec.Users[user] = userFlags
		} else {
			for _, userFlag := range userFlags {
				foundUserFlag := false
				for _, existingFlag := range existingUserFlags {
					if userFlag == existingFlag {
						foundUserFlag = true
					}
				}
				if !foundUserFlag {
					existingUserFlags = append(existingUserFlags, userFlag)
					fetchDatabase.Spec.Users[user] = existingUserFlags
				}
			}
		}
	}

	for database, user := range instance.Spec.Databases {
		_, ok := fetchDatabase.Spec.Databases[database]
		if !ok {
			fetchDatabase.Spec.Databases[database] = user
		}
	}

	_, err = controllerutil.CreateOrUpdate(ctx.Context, ctx, fetchDatabase.DeepCopyObject(), func(existingObj runtime.Object) error {
		existing := existingObj.(*postgresv1.Postgresql)
		existing.Labels = fetchDatabase.Labels
		existing.Annotations = fetchDatabase.Annotations
		existing.Spec = fetchDatabase.Spec

		return nil
	})
	if err != nil {
		return components.Result{StatusModifier: func(obj runtime.Object) error {
			instance := obj.(*dbv1beta1.PostgresOperator)
			instance.Status.Status = dbv1beta1.StatusError
			instance.Status.Message = "postgres_operator: Failed to update Postgresql object"
			return nil
		}}, errors.Wrapf(err, "postgres_operator: Failed to update Postgresql object")
	}

	return components.Result{StatusModifier: func(obj runtime.Object) error {
		instance := obj.(*dbv1beta1.PostgresOperator)
		instance.Status.Status = dbv1beta1.StatusReady
		instance.Status.Message = "Ready"
		return nil
	}}, nil
}
