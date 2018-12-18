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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func ReconcileMeta(target, existing *metav1.ObjectMeta) error {
	if target.Labels != nil {
		if existing.Labels == nil {
			existing.Labels = map[string]string{}
		}
		for k, v := range target.Labels {
			existing.Labels[k] = v
		}
	}
	if target.Annotations != nil {
		if existing.Annotations == nil {
			existing.Annotations = map[string]string{}
		}
		for k, v := range target.Annotations {
			existing.Annotations[k] = v
		}
	}
	return nil
}
