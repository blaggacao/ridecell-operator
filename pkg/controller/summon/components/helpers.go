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

package components

import (
  summonv1beta "github.com/Ridecell/ridecell-operator/pkg/apis/summon/v1beta1"
)

// Helper function for use as a StatusModifier which just sets the main status.
func setStatus(status string) components.StatusModifier {
  return func(obj runtime.Object) error {
    instance := obj.(*summonv1beta.SummonPlatform)
    instance.Status.Status = status
    return nil
  }
}
