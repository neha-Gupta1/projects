/*
Copyright 2024.

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

package v1

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VolrestoreSpec defines the desired state of Volrestore
type VolrestoreSpec struct {
	// VolumeName is the volsnap name whose PV data is to be restored
	VolSnapName string `json:"volume_snap_name,omitempty"`
}

// VolrestoreStatus defines the observed state of Volrestore
type VolrestoreStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Phase     v1.PersistentVolumeClaimPhase       `json:"phase,omitempty"`
	Condition []v1.PersistentVolumeClaimCondition `json:"condition,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Volrestore is the Schema for the volrestores API
type Volrestore struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VolrestoreSpec   `json:"spec,omitempty"`
	Status VolrestoreStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VolrestoreList contains a list of Volrestore
type VolrestoreList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Volrestore `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Volrestore{}, &VolrestoreList{})
}
