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

	"github.com/lib/pq"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/Ridecell/ridecell-operator/pkg/components/postgres"
)

type databaseComponent struct{}

func NewDatabase() *databaseComponent {
	return &databaseComponent{}
}

func (_ *databaseComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (_ *databaseComponent) IsReconcilable(_ *components.ComponentContext) bool {
	return true
}

func (comp *databaseComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*dbv1beta1.PostgresExtension)

	// Connect to the database.
	db, err := postgres.Open(ctx, &instance.Spec.Database)
	if err != nil {
		return reconcile.Result{Requeue: true}, err
	}

	// Two codepaths because both queries look very different depending on if we have a version or not.
	if instance.Spec.Version == "" {
		// Create the extension if it doesn't exist already.
		_, err = db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s", pq.QuoteIdentifier(instance.Spec.ExtensionName)))
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "database: Error running CREATE EXTENSION")
		}

		// Upgrade the extension if it did exist.
		_, err = db.Exec(fmt.Sprintf("ALTER EXTENSION %s UPDATE", pq.QuoteIdentifier(instance.Spec.ExtensionName)))
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "database: Error running ALTER EXTENSION")
		}
	} else {
		// Create the extension if it doesn't exist already.
		_, err = db.Exec(fmt.Sprintf("CREATE EXTENSION IF NOT EXISTS %s WITH VERSION %s", pq.QuoteIdentifier(instance.Spec.ExtensionName), pq.QuoteIdentifier(instance.Spec.Version)))
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "database: Error running CREATE EXTENSION")
		}

		// Upgrade the extension if it did exist.
		_, err = db.Exec(fmt.Sprintf("ALTER EXTENSION %s UPDATE TO %s", pq.QuoteIdentifier(instance.Spec.ExtensionName), pq.QuoteIdentifier(instance.Spec.Version)))
		if err != nil {
			return reconcile.Result{}, errors.Wrap(err, "database: Error running ALTER EXTENSION")
		}
	}

	// Success!
	instance.Status.Status = dbv1beta1.StatusReady
	instance.Status.Message = fmt.Sprintf("Extension %v created", instance.Spec.ExtensionName)

	return reconcile.Result{}, nil
}
