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

package secrets_test

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/DATA-DOG/go-sqlmock.v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	summonv1beta1 "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
	"github.com/Ridecell/ridecell-operator/pkg/dbpool"
	"github.com/Ridecell/ridecell-operator/pkg/test_helpers"
)

var _ = Describe("Secrets controller", func() {
	var helpers *test_helpers.PerTestHelpers

	It("Creates pullsecret", func() {
		helpers = testHelpers.SetupTest()

		pullSecret := &corev1.Secret{ObjectMeta: metav1.ObjectMeta{Name: "pull-secret", Namespace: helpers.OperatorNamespace}, Type: "kubernetes.io/dockerconfigjson", StringData: map[string]string{".dockerconfigjson": "{\"auths\": {}}"}}
		err := helpers.Client.Create(context.TODO(), pullSecret)
		Expect(err).NotTo(HaveOccurred())
	})
})
