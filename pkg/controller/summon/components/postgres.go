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
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
)

type postgresComponent struct {
	postgresTemplatePath string
	operatorTemplatePath string
}

func NewPostgres(postgresTemplatePath string, operatorTemplatePath string) *postgresComponent {
	return &postgresComponent{
		postgresTemplatePath: postgresTemplatePath,
		operatorTemplatePath: operatorTemplatePath,
	}
}

func (comp *postgresComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&postgresv1.Postgresql{},
	}
}

func (_ *postgresComponent) IsReconcilable(_ *components.ComponentContext) bool {
	return true
}

func (comp *postgresComponent) Reconcile(ctx *components.ComponentContext) (components.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	var existingOperator *dbv1beta1.PostgresOperatorDatabase
	if *instance.Spec.DatabaseSpec.SharedDatabase == true {
		res, _, err := ctx.CreateOrUpdate(comp.operatorTemplatePath, nil, func(goalObj, existingObj runtime.Object) error {
			goal := goalObj.(*dbv1beta1.PostgresOperatorDatabase)
			existingOperator = existingObj.(*dbv1beta1.PostgresOperatorDatabase)
			// Copy the Spec over.
			existingOperator.Spec = goal.Spec
			return nil
		})
		if err != nil {
			return res, errors.Wrapf(err, "postgres: failed to create or update postgresoperatordatabase object")
		}

		fetchPostgres := &postgresv1.Postgresql{}
		err = ctx.Get(ctx.Context, types.NamespacedName{Name: instance.Spec.DatabaseSpec.SharedDatabaseName, Namespace: instance.Namespace}, fetchPostgres)
		if err != nil {
			return components.Result{}, errors.Wrapf(err, "postgres: failed to get shared database object")
		}
		res.StatusModifier = func(obj runtime.Object) error {
			instance.Status.PostgresStatus = fetchPostgres.Status
			return nil
		}

		return res, nil
	}

	var existingDatabase *postgresv1.Postgresql
	res, _, err := ctx.CreateOrUpdate(comp.postgresTemplatePath, nil, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*postgresv1.Postgresql)
		existingDatabase = existingObj.(*postgresv1.Postgresql)
		// Copy the Spec over.
		existingDatabase.Spec = goal.Spec
		return nil
	})
	setPostgresStatus := func(obj runtime.Object) error {
		instance.Status.PostgresStatus = existingDatabase.Status
		return nil
	}
	if err != nil {
		res.StatusModifier = setPostgresStatus
		return res, err
	}
	if !existingDatabase.Status.Success() {
		// I honestly can't tell how this field works. I think it's just ignored by the CRD entirely. Trying to play both sides just in case.
		if existingDatabase.Error == "" {
			err = errors.Errorf("postgres: status is %s", existingDatabase.Status)
		} else {
			err = errors.Errorf("postgres: status is %s: %s", existingDatabase.Status, existingDatabase.Error)
		}
		return components.Result{StatusModifier: setPostgresStatus}, err
	}
	if existingDatabase.Status != postgresv1.ClusterStatusUnknown {
		// DB creation was started, and we already checked if something went wrong so we are at least up to initializing.
		res.StatusModifier = func(obj runtime.Object) error {
			instance := obj.(*summonv1beta1.SummonPlatform)
			instance.Status.Status = summonv1beta1.StatusInitializing
			return setPostgresStatus(obj)
		}
	}
	return res, err
}
