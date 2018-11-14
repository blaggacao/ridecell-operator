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

package components

import (
	"database/sql"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/Ridecell/ridecell-operator/pkg/dbpool"
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
	// instance := ctx.Top.(*summonv1beta1.DjangoUser)
	db, err := comp.openDatabase(ctx)
	if err != nil {
		return reconcile.Result{}, err
	}

	_, err = db.Query("SELECT version()")
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (comp *databaseComponent) openDatabase(ctx *components.ComponentContext) (*sql.DB, error) {
	instance := ctx.Top.(*summonv1beta1.DjangoUser)
	dbInfo := instance.Spec.Database
	passwordSecret := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: dbInfo.PasswordSecretRef.Name, Namespace: instance.Namespace}, passwordSecret)
	if err != nil {
		return nil, fmt.Errorf("database: Unable to load database secret %v: %v", dbInfo.PasswordSecretRef.Name, err)
	}
	dbPassword, ok := passwordSecret.Data[dbInfo.PasswordSecretRef.Key]
	if !ok {
		return nil, fmt.Errorf("database: Password key %v not found in database secret %v", dbInfo.PasswordSecretRef.Key, dbInfo.PasswordSecretRef.Name)
	}
	connStr := fmt.Sprintf("host=%s port=%v dbname=%s user=%v password='%v' sslmode=verify-full", dbInfo.Host, dbInfo.Port, dbInfo.Database, dbInfo.Username, dbPassword)
	db, err := dbpool.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("database: Unable to open database connection: %v", err)
	}
	return db, nil
}
