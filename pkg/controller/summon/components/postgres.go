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
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

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

func (comp *postgresComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	res, _, err := ctx.CreateOrUpdate(comp.templatePath, nil, func(goalObj, existingObj runtime.Object) error {
		goal := goalObj.(*postgresv1.Postgresql)
		existing := existingObj.(*postgresv1.Postgresql)
		// Store the postgres status.
		instance.Status.PostgresStatus = existing.Status
		// Copy the Spec over.
		existing.Spec = goal.Spec
		return nil
	})
	return res, err
}
