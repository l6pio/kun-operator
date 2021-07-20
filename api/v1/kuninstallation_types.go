/*
Copyright 2021.

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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// KunInstallationSpec defines the desired state of KunInstallation
type KunInstallationSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Hub     string                 `json:"hub,omitempty"`
	Mongodb KunInstallationMongodb `json:"mongodb,omitempty"`
	Server  KunInstallationServer  `json:"server,omitempty"`
	UI      KunInstallationUI      `json:"ui,omitempty"`
}

type KunInstallationMongodb struct {
	Addr string `json:"addr,omitempty"`
	User string `json:"user,omitempty"`
	Pass string `json:"pass,omitempty"`
}

type KunInstallationServer struct {
	Replicas  *int32                      `json:"replicas,omitempty"`
	Port      *int32                      `json:"port,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Ingress   KunInstallationIngress      `json:"ingress,omitempty"`
}

type KunInstallationUI struct {
	Enabled   bool                        `json:"enabled,omitempty"`
	Replicas  *int32                      `json:"replicas,omitempty"`
	Port      *int32                      `json:"port,omitempty"`
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	Ingress   KunInstallationIngress      `json:"ingress,omitempty"`
}

type KunInstallationIngress struct {
	Enabled bool   `json:"enabled,omitempty"`
	Host    string `json:"host,omitempty"`
	Path    string `json:"path,omitempty"`
}

// KunInstallationStatus defines the observed state of KunInstallation
type KunInstallationStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// KunInstallation is the Schema for the kuninstallations API
type KunInstallation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KunInstallationSpec   `json:"spec,omitempty"`
	Status KunInstallationStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KunInstallationList contains a list of KunInstallation
type KunInstallationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KunInstallation `json:"items"`
}

func init() {
	SchemeBuilder.Register(&KunInstallation{}, &KunInstallationList{})
}
