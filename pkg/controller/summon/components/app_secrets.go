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
	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	dbv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/db/v1beta1"
	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type appSecretComponent struct{}

func NewAppSecret() *appSecretComponent {
	return &appSecretComponent{}
}

func (comp *appSecretComponent) WatchTypes() []runtime.Object {
	return []runtime.Object{
		&corev1.Secret{},
	}
}

func (_ *appSecretComponent) IsReconcilable(ctx *components.ComponentContext) bool {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	if instance.Status.PostgresStatus != postgresv1.ClusterStatusRunning {
		return false
	}
	return true
}

func (comp *appSecretComponent) Reconcile(ctx *components.ComponentContext) (reconcile.Result, error) {
	instance := ctx.Top.(*summonv1beta1.SummonPlatform)

	appSecrets := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: instance.Spec.Secret, Namespace: instance.Namespace}, appSecrets)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "app_secrets: Failed to get existing app secrets")
	}

	postgresSecret := &corev1.Secret{}
	err = ctx.Get(ctx.Context, types.NamespacedName{Name: fmt.Sprintf("summon.%s-database.credentials", instance.Name), Namespace: instance.Namespace}, postgresSecret)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "app_secrets: Postgres password not found")
	}
	postgresPassword, ok := postgresSecret.Data["password"]
	if !ok {
		return reconcile.Result{}, errors.New("app_secrets: Postgres password not found in secret")
	}

	appSecrets.Data["DATABASE_URL"] = []byte(fmt.Sprintf("postgis://summon:%s@%s-database/summon", postgresPassword, instance.Name))
	appSecrets.Data["OUTBOUNDSMS_URL"] = []byte(fmt.Sprintf("https://%s.prod.ridecell.io/outbound-sms", instance.Name))
	appSecrets.Data["SMS_WEBHOOK_URL"] = []byte(fmt.Sprintf("https://%s.ridecell.us/sms/receive/", instance.Name))
	appSecrets.Data["CELERY_BROKER_URL"] = []byte(fmt.Sprintf("redis://%s-redis/2", instance.Name))

	parsedYaml, err := yaml.Marshal(appSecrets.Data)
	if err != nil {
		return reconcile.Result{Requeue: true}, errors.Wrapf(err, "app_secrets: yaml.Marshal failed")
	}

	newSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("summon.%s.app-secrets", instance.Name), Namespace: instance.Namespace},
		Data:       map[string][]byte{"summon-platform.yml": parsedYaml},
	}

	fetchTarget := &corev1.Secret{}
	err = ctx.Get(ctx.Context, types.NamespacedName{Name: fmt.Sprintf("summon.%s.app-secrets", instance.Name), Namespace: instance.Namespace}, fetchTarget)
	if err != nil {
		if !k8serrors.IsNotFound(err) {
			return reconcile.Result{}, err
		}
	}

	_, err = controllerutil.CreateOrUpdate(ctx.Context, ctx, newSecret, func(fetchTarget runtime.Object) error {
		existing := fetchTarget.(*corev1.Secret)
		// Sync important fields.
		err := controllerutil.SetControllerReference(instance, existing, ctx.Scheme)
		if err != nil {
			return errors.Wrapf(err, "app_secrets: Failed to set controller reference")
		}
		existing.ObjectMeta = newSecret.ObjectMeta
		existing.Type = newSecret.Type
		existing.Data = newSecret.Data
		return nil
	})

	if err != nil {
		return reconcile.Result{}, errors.Wrapf(err, "app_secrets: Failed to update secret object")
	}

	return reconcile.Result{}, nil
}
