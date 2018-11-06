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

package summon

import (
	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/Ridecell/ridecell-operator/pkg/controller/summon/components/configmap"
	"github.com/Ridecell/ridecell-operator/pkg/controller/summon/components/defaults"
	"github.com/Ridecell/ridecell-operator/pkg/controller/summon/components/deployment"
	"github.com/Ridecell/ridecell-operator/pkg/controller/summon/components/ingress"
	"github.com/Ridecell/ridecell-operator/pkg/controller/summon/components/migrations"
	"github.com/Ridecell/ridecell-operator/pkg/controller/summon/components/postgres"
	"github.com/Ridecell/ridecell-operator/pkg/controller/summon/components/pull_secret"
	"github.com/Ridecell/ridecell-operator/pkg/controller/summon/components/service"
	"github.com/Ridecell/ridecell-operator/pkg/controller/summon/components/statefulset"
)

// Add creates a new Summon Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this summon.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	// return &ReconcileSummon{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
	return components.NewController(mgr, &summonv1beta1.SummonPlatform{}, Templates, []components.Component{
		// Set default values.
		defaults.New(),

		// Top-level components.
		pull_secret.New(),
		postgres.New("postgres.yml.tpl"),
		configmap.New("configmap.yml.tpl"),
		migrations.New("migrations.yml.tpl"),

		// Redis components.
		deployment.New("redis/deployment.yml.tpl", false),
		service.New("redis/service.yml.tpl"),

		// Web components.
		deployment.New("web/deployment.yml.tpl", true),
		service.New("web/service.yml.tpl"),
		ingress.New("web/ingress.yml.tpl"),

		// Daphne components.
		deployment.New("daphne/deployment.yml.tpl", true),
		service.New("daphne/service.yml.tpl"),
		ingress.New("daphne/ingress.yml.tpl"),

		// Static file components.
		deployment.New("static/deployment.yml.tpl", true),
		service.New("static/service.yml.tpl"),
		ingress.New("static/ingress.yml.tpl"),

		// Celery components.
		deployment.New("celeryd/deployment.yml.tpl", true),

		// Celerybeat components.
		statefulset.New("celerybeat/statefulset.yml.tpl", true),
		service.New("celerybeat/service.yml.tpl"),

		// Channelworker components.
		deployment.New("channelworker/deployment.yml.tpl", true),
	})
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("summon-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Summon
	err = c.Watch(&source.Kind{Type: &summonv1beta1.SummonPlatform{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	for _, watchObj := range r.(*components.ComponentController).WatchTypes() {
		err = c.Watch(&source.Kind{Type: watchObj}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &summonv1beta1.SummonPlatform{},
		})
		if err != nil {
			return err
		}
	}

	return nil
}
