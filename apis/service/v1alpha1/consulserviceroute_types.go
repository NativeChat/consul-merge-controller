/*
Copyright 2021 Progress Software Corporation.

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
	consulk8s "github.com/hashicorp/consul-k8s/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConsulServiceRouteSpec defines the desired state of ConsulServiceRoute
type ConsulServiceRouteSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Route consulk8s.ServiceRoute `json:"route"`
}

// ConsulServiceRouteStatus defines the observed state of ConsulServiceRoute
type ConsulServiceRouteStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	UpdatedAt string `json:"updatedAt,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// ConsulServiceRoute is the Schema for the consulserviceroutes API
type ConsulServiceRoute struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConsulServiceRouteSpec   `json:"spec,omitempty"`
	Status ConsulServiceRouteStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConsulServiceRouteList contains a list of ConsulServiceRoute
type ConsulServiceRouteList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ConsulServiceRoute `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ConsulServiceRoute{}, &ConsulServiceRouteList{})
}
