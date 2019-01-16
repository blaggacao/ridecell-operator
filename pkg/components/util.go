/*
Copyright 2018-2019 Ridecell, Inc.

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

	meta "k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
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

func GetOwnerNames(obj runtime.Object, kind string) (sets.String, error) {
	accessor, err := meta.Accessor(obj)
	if err != nil {
		return nil, err
	}

	names := sets.NewString()
	for _, ownerReference := range accessor.GetOwnerReferences() {
		if ownerReference.Kind == kind {
			names.Insert(ownerReference.Name)
		}
	}
	return names, nil
}

func GetOwnerName(obj runtime.Object, kind string) (*string, error) {
	ss, err := GetOwnerNames(obj, kind)
	if err != nil {
		return nil, err
	}
	if ss.Len() > 1 {
		err := fmt.Errorf("more than one matching %v instance found", kind)
		return nil, err
	} else if ss.Len() < 1 {
		err := fmt.Errorf("no matching %v instance found", kind)
		return nil, err
	}
	return &ss.List()[0], nil
}
