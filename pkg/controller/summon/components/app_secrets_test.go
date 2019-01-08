/*
Copyright 2019 Ridecell, Inc.

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

package components_test

import (
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"gopkg.in/yaml.v2"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	summoncomponents "github.com/Ridecell/ridecell-operator/pkg/controller/summon/components"
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type appSecretData struct {
	FERNET_KEYS       []string `yaml:"FERNET_KEYS,omitempty"`
	DATABASE_URL      []byte   `yaml:"DATABASE_URL,omitempty"`
	OUTBOUNDSMS_URL   []byte   `yaml:"OUTBOUNDSMS_URL,omitempty"`
	SMS_WEBHOOK_URL   []byte   `yaml:"SMS_WEBHOOK_URL,omitempty"`
	CELERY_BROKER_URL []byte   `yaml:"CELERY_BROKER_URL,omitempty"`
}

var _ = Describe("app_secrets Component", func() {

	It("Unreconcilable when db not ready", func() {
		comp := summoncomponents.NewAppSecret()
		Expect(comp.IsReconcilable(ctx)).To(Equal(false))
	})

	It("Reconcilable when db is ready", func() {
		comp := summoncomponents.NewAppSecret()
		instance.Status.PostgresStatus = postgresv1.ClusterStatusRunning
		Expect(comp.IsReconcilable(ctx)).To(Equal(true))
	})

	It("Run reconcile without a postgres password", func() {
		comp := summoncomponents.NewAppSecret()
		instance.Status.PostgresStatus = postgresv1.ClusterStatusRunning
		_, err := comp.Reconcile(ctx)
		Expect(err).To(HaveOccurred())
	})

	It("Run reconcile with a blank postgres password", func() {
		comp := summoncomponents.NewAppSecret()
		instance.Status.PostgresStatus = postgresv1.ClusterStatusRunning

		appSecrets := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "testsecret", Namespace: instance.Namespace},
			Data:       map[string][]byte{},
		}

		postgresSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("summon.%s-database.credentials", instance.Name), Namespace: instance.Namespace},
			Data:       map[string][]byte{},
		}

		fernetKeys := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s.fernet-keys", instance.Name), Namespace: instance.Namespace},
			Data:       map[string][]byte{},
		}
		ctx.Client = fake.NewFakeClient(appSecrets, postgresSecret, fernetKeys)
		_, err := comp.Reconcile(ctx)
		Expect(err.Error()).To(Equal("app_secrets: Postgres password not found in secret"))
	})

	It("Sets postgres password and checks reconcile output", func() {
		comp := summoncomponents.NewAppSecret()
		//Set status so that IsReconcileable returns true
		instance.Status.PostgresStatus = postgresv1.ClusterStatusRunning

		appSecrets := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "testsecret", Namespace: instance.Namespace},
			Data:       map[string][]byte{"filler": []byte("test")},
		}

		postgresSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("summon.%s-database.credentials", instance.Name), Namespace: instance.Namespace},
			Data:       map[string][]byte{"password": []byte("postgresPassword")},
		}

		formattedTime := time.Time.Format(time.Now().UTC(), summoncomponents.CustomTimeLayout)

		fernetKeys := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s.fernet-keys", instance.Name), Namespace: instance.Namespace},
			Data:       map[string][]byte{formattedTime: []byte("lorem ipsum")},
		}

		ctx.Client = fake.NewFakeClient(appSecrets, postgresSecret, fernetKeys)
		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())

		fetchSecret := &corev1.Secret{}
		err = ctx.Client.Get(ctx.Context, types.NamespacedName{Name: fmt.Sprintf("summon.%s.app-secrets", instance.Name), Namespace: instance.Namespace}, fetchSecret)
		Expect(err).ToNot(HaveOccurred())

		byteData := fetchSecret.Data["summon-platform.yml"]
		var parsedYaml appSecretData
		err = yaml.Unmarshal(byteData, &parsedYaml)
		Expect(err).ToNot(HaveOccurred())

		Expect(string(parsedYaml.DATABASE_URL)).To(Equal("postgis://summon:postgresPassword@foo-database/summon"))
		Expect(string(parsedYaml.OUTBOUNDSMS_URL)).To(Equal("https://foo.prod.ridecell.io/outbound-sms"))
		Expect(string(parsedYaml.SMS_WEBHOOK_URL)).To(Equal("https://foo.ridecell.us/sms/receive/"))
		Expect(string(parsedYaml.CELERY_BROKER_URL)).To(Equal("redis://foo-redis/2"))
	})

	It("reconciles with existing fernet keys", func() {
		comp := summoncomponents.NewAppSecret()

		// Is there a way I could write this setup in not such a verbose way?
		now := time.Now().UTC()
		unsortedTimes := make(map[string][]byte)

		durationOne, _ := time.ParseDuration("-1h")
		durationTwo, _ := time.ParseDuration("-2h")
		durationThree, _ := time.ParseDuration("-3h")
		durationFour, _ := time.ParseDuration("-4h")
		durationFive, _ := time.ParseDuration("-5h")

		timeOne := now.Add(durationOne)
		timeTwo := now.Add(durationTwo)
		timeThree := now.Add(durationThree)
		timeFour := now.Add(durationFour)
		timeFive := now.Add(durationFive)

		unsortedTimes[time.Time.Format(timeOne, summoncomponents.CustomTimeLayout)] = []byte("1")
		unsortedTimes[time.Time.Format(timeTwo, summoncomponents.CustomTimeLayout)] = []byte("2")
		unsortedTimes[time.Time.Format(timeThree, summoncomponents.CustomTimeLayout)] = []byte("3")
		unsortedTimes[time.Time.Format(timeFour, summoncomponents.CustomTimeLayout)] = []byte("4")
		unsortedTimes[time.Time.Format(timeFive, summoncomponents.CustomTimeLayout)] = []byte("5")

		fernetKeys := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("%s.fernet-keys", instance.Name), Namespace: instance.Namespace},
			Data:       unsortedTimes,
		}

		//Set status so that IsReconcileable returns true
		instance.Status.PostgresStatus = postgresv1.ClusterStatusRunning

		postgresSecret := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: fmt.Sprintf("summon.%s-database.credentials", instance.Name), Namespace: instance.Namespace},
			Data:       map[string][]byte{"password": []byte("postgresPassword")},
		}

		appSecrets := &corev1.Secret{
			ObjectMeta: metav1.ObjectMeta{Name: "testsecret", Namespace: instance.Namespace},
			Data:       map[string][]byte{},
		}

		ctx.Client = fake.NewFakeClient(appSecrets, postgresSecret, fernetKeys)
		_, err := comp.Reconcile(ctx)
		Expect(err).ToNot(HaveOccurred())

		fetchSecret := &corev1.Secret{}
		err = ctx.Client.Get(ctx.Context, types.NamespacedName{Name: fmt.Sprintf("summon.%s.app-secrets", instance.Name), Namespace: instance.Namespace}, fetchSecret)
		Expect(err).ToNot(HaveOccurred())

		byteData := fetchSecret.Data["summon-platform.yml"]
		var parsedYaml appSecretData
		err = yaml.Unmarshal(byteData, &parsedYaml)
		Expect(err).ToNot(HaveOccurred())

		stringSlices := parsedYaml.FERNET_KEYS

		expectedSlices := []string{"1", "2", "3", "4", "5"}
		Expect(stringSlices).To(Equal(expectedSlices))

	})
})
