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

package v1beta1

import (
	"github.com/Ridecell/ridecell-operator/pkg/components"
)

func (s *PullSecret) GetStatus() components.Status {
	return s.Status
}

func (s *PullSecret) SetStatus(status components.Status) {
	s.Status = status.(PullSecretStatus)
}

func (s *PullSecret) SetErrorStatus(errorMsg string) {
	s.Status.Status = StatusError
	s.Status.Message = errorMsg
}

func (es *EncryptedSecret) GetStatus() components.Status {
	return es.Status
}

func (es *EncryptedSecret) SetStatus(status components.Status) {
	es.Status = status.(EncryptedSecretStatus)
}

func (es *EncryptedSecret) SetErrorStatus(errorMsg string) {
	es.Status.Status = StatusError
	es.Status.Message = errorMsg
}
