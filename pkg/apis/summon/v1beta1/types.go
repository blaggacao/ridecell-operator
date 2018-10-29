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

package v1beta1

import (
	postgresv1 "github.com/zalando-incubator/postgres-operator/pkg/apis/acid.zalan.do/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SummonPlatformSpec defines the desired state of SummonPlatform
type SummonPlatformSpec struct {
	// Important: Run "make" to regenerate code after modifying this file

	// Hostname to use for the instance. Defaults to $NAME.ridecell.us.
	// +optional
	Hostname *string `json:"hostname,omitempty"`
	// Summon image version to deploy.
	Version string `json:"version"`
	// Name of the secret to use for secret values.
	Secret string `json:"secret"`

	// Number of gunicorn pods to run. Defaults to 1.
	// +optional
	WebReplicas *int32 `json:"web_replicas,omitempty"`
	// Number of daphne pods to run. Defaults to 1.
	// +optional
	DaphneReplicas *int32 `json:"daphne_replicas,omitempty"`
	// Number of celeryd pods to run. Defaults to 1.
	// +optional
	WorkerReplicas *int32 `json:"worker_replicas,omitempty"`
	// Number of channelworker pods to run. Defaults to 1.
	// +optional
	ChannelWorkerReplicas *int32 `json:"channel_worker_replicas,omitempty"`
	// Number of caddy pods to run. Defaults to 1.
	// +optional
	StaticReplicas *int32 `json:"static_replicas,omitempty"`
}

// SummonPlatformStatus defines the observed state of SummonPlatform
type SummonPlatformStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Current Postgresql status if one exists.
	PostgresStatus *postgresv1.PostgresStatus `json:"postgresStatus,omitempty"`

	// Previous version for which migrations ran successfully.
	// +optional
	MigrateVersion string `json:"migrateVersion,omitempty"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SummonPlatform is the Schema for the summonplatforms API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type SummonPlatform struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SummonPlatformSpec   `json:"spec,omitempty"`
	Status SummonPlatformStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SummonPlatformList contains a list of SummonPlatform
type SummonPlatformList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SummonPlatform `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SummonPlatform{}, &SummonPlatformList{})
}
