/*
Copyright 2025.

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PretzelAISpec defines the desired state of PretzelAI
type PretzelAISpec struct {
	// Replicas defines the desired number of PretzelAI pods to run.
	// +kubebuilder:validation:Minimum=1
	// +kubebuilder:default=1
	Replicas *int32 `json:"replicas,omitempty"`

	// Image allows overriding the container image used for PretzelAI.
	// If not specified, the operator uses the default image.
	Image string `json:"image,omitempty"`

	// ConfigMapName is the optional name of a ConfigMap containing configuration files.
	ConfigMapName string `json:"configMapName,omitempty"`

	// ServiceType specifies how to expose the application (ClusterIP, NodePort, LoadBalancer).
	// +kubebuilder:validation:Enum=ClusterIP;NodePort;LoadBalancer
	// +kubebuilder:default=ClusterIP
	ServiceType string `json:"serviceType,omitempty"`
}

// PretzelAIStatus defines the observed state of PretzelAI
type PretzelAIStatus struct {
	// ReadyReplicas is the number of pods that are currently ready and serving traffic.
	ReadyReplicas int32 `json:"readyReplicas,omitempty"`

	// AvailableReplicas is the total number of available pods.
	AvailableReplicas int32 `json:"availableReplicas,omitempty"`

	// ServiceStatus reports the current type/status of the Service.
	ServiceStatus string `json:"serviceStatus,omitempty"`

	// ConfigMapStatus confirms if the referenced ConfigMap was found and applied.
	ConfigMapStatus string `json:"configMapStatus,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:printcolumn:name="Replicas",type="integer",JSONPath=".spec.replicas",description="The desired number of replicas"
//+kubebuilder:printcolumn:name="Ready",type="integer",JSONPath=".status.readyReplicas",description="The number of ready replicas"
//+kubebuilder:printcolumn:name="Service",type="string",JSONPath=".spec.serviceType",description="The service type"
//+kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"

// PretzelAI is the Schema for the pretzelai API
type PretzelAI struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PretzelAISpec   `json:"spec,omitempty"`
	Status PretzelAIStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// PretzelAIList contains a list of PretzelAI
type PretzelAIList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PretzelAI `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PretzelAI{}, &PretzelAIList{})
}
