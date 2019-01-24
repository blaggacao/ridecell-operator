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

package helpers

import (
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	"github.com/Ridecell/ridecell-operator/pkg/components"
)

type SecretRef struct {
	Name string `json:"name"`
	Key  string `json:"key,omitempty"`
}

func (s *SecretRef) Resolve(ctx *components.ComponentContext, defaultKey string) (string, error) {
	namespace := ctx.Top.(metav1.Object).GetNamespace()
	key := s.Key
	if key == "" {
		key = defaultKey
	}

	fetch := &corev1.Secret{}
	err := ctx.Get(ctx.Context, types.NamespacedName{Name: s.Name, Namespace: namespace}, fetch)
	if err != nil {
		return "", errors.Wrapf(err, "unable to fetch secret %s/%s", namespace, s.Name)
	}

	val, ok := fetch.Data[key]
	if !ok {
		return "", errors.Errorf("key %s not found in secret %s/%s", key, namespace, s.Name)
	}

	return string(val), nil
}
