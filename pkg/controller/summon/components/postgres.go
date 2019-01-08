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
	"github.com/pkg/errors"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"k8s.io/apimachinery/pkg/runtime"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type postgresComponent struct {
	templatePath string
}

func NewPostgres(templatePath string) *postgresComponent {
	return &postgresComponent{templatePath: templatePath}
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
	var existing *postgresv1.Postgresql
	res, _, err := ctx.CreateOrUpdate(comp.templatePath, nil, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*postgresv1.Postgresql)
		existing = existingObj.(*postgresv1.Postgresql)
		// Copy the Spec over.
		existing.Spec = goal.Spec
		return nil
	})
	setPostgresStatus := func(obj runtime.Object) error {
		instance := obj.(*summonv1beta1.SummonPlatform)
		instance.Status.PostgresStatus = existing.Status
		return nil
	}
	if err != nil {
		res.StatusModifier = setPostgresStatus
		return res, err
	}
	if !existing.Status.Success() {
		// I honestly can't tell how this field works. I think it's just ignored by the CRD entirely. Trying to play both sides just in case.
		if existing.Error == "" {
			err = errors.Errorf("postgres: status is %s", existing.Status)
		} else {
			err = errors.Errorf("postgres: status is %s: %s", existing.Status, existing.Error)
		}
		return components.Result{StatusModifier: setPostgresStatus}, err
	}
	if existing.Status != postgresv1.ClusterStatusUnknown {
		// DB creation was started, and we already checked if something went wrong so we are at least up to initializing.
		res.StatusModifier = func(obj runtime.Object) error {
			instance := obj.(*summonv1beta1.SummonPlatform)
			instance.Status.Status = summonv1beta1.StatusInitializing
			return setPostgresStatus(obj)
		}
	}
	return res, err
}
