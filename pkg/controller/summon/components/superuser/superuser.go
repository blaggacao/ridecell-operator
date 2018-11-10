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

package superuser

import (
	"database/sql"
	"fmt"

	"github.com/golang/glog"
	_ "github.com/lib/pq"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/Ridecell/ridecell-operator/pkg/dbpool"
)

type superuserComponent struct{}

func New() *superuserComponent {
	return &superuserComponent{}
}

func (comp *superuserComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{}
}

func (comp *superuserComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	// Wait for the database to be up and migrated.
	return instance.Status.PostgresStatus != nil && *instance.Status.PostgresStatus == postgresv1.ClusterStatusRunning && instance.Spec.Version == instance.Status.MigrateVersion
}

func (comp *superuserComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	db, err := openDatabase(ctx)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = db.Query("SELECT version()")
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func openDatabase(ctx *components.ComponentContext) (*sql.DB, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)
	dbSecretName := fmt.Sprintf("summon.%s-database.credentials", instance.Name)
	dbSecret := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: dbSecretName, Namespace: instance.Namespace}, dbSecret)
	if err != nil {
		return nil, fmt.Errorf("superuser: Unable to load database secret %v: %v", dbSecretName, err)
	}
	dbPassword, ok := dbSecret.Data["password"]
	if !ok {
		return nil, fmt.Errorf("superuser: Password key not found in database secret %v", dbSecretName)
	}
	glog.V(2).Infof("[%s/%s] superuser: opening DB connection: host=%s-database dbname=summon user=summon password='%s...'", instance.Namespace, instance.Name, instance.Name, dbPassword[:4])
	connStr := fmt.Sprintf("host=%s-database dbname=summon user=summon password='%s' sslmode=verify-full", instance.Name, dbPassword)
	db, err := dbpool.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("superuser: Unable to open database connection: %v", err)
	}
	return db, nil
}
