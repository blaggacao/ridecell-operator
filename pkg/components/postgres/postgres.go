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

package postgres

import (
	"database/sql"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/components"
	"github.com/Ridecell/ridecell-operator/pkg/dbpool"
)

// Open a connection to the Postgres database as defined by a PostgresConnection object.
func Open(ctx *components.ComponentContext, dbInfo *dbv1beta1.PostgresConnection) (*sql.DB, error) {
	namespace := ctx.Top.(metav1.Object).GetNamespace()
	passwordSecret := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: dbInfo.PasswordSecretRef.Name, Namespace: namespace}, passwordSecret)
	if err != nil {
		return nil, errors.Wrapf(err, "OpenPostgres: Unable to load database secret %s/%s", namespace, dbInfo.PasswordSecretRef.Name)
	}
	dbPassword, ok := passwordSecret.Data[dbInfo.PasswordSecretRef.Key]
	if !ok {
		return nil, errors.Errorf("database: Password key %v not found in database secret %s/%s", dbInfo.PasswordSecretRef.Key, namespace, dbInfo.PasswordSecretRef.Name)
	}
	connStr := fmt.Sprintf("host=%s port=%v dbname=%s user=%v password='%s' sslmode=require", dbInfo.Host, dbInfo.Port, dbInfo.Database, dbInfo.Username, dbPassword)
	db, err := dbpool.Open("postgres", connStr)
	if err != nil {
		return nil, errors.Wrap(err, "database: Unable to open database connection")
	}
	return db, nil
}
