/*
Copyright 2022 K-rays.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// WatcherSpec defines the desired state of Watcher
type WatcherSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Quantity of instances
	Exit         int32                         `json:"exit"`
	Timeout      int32                         `json:"timeout"`
	ConfigMaps   []BasicConfigurationCondition `json:"configmaps,omitempty"`
	Secrets      []BasicConfigurationCondition `json:"secrets,omitempty"`
	Services     []BasicConfigurationCondition `json:"services,omitempty"`
	Pods         []PodCondition                `json:"pods,omitempty"`
	Jobs         []JobCondition                `json:"jobs,omitempty"`
	Watchers     []WatcherCondition            `json:"watchers,omitempty"`
	Deployments  []DeploymentCondition         `json:"deployments,omitempty"`
	StatefulSets []StatefulSetCondition        `json:"statefulSets,omitempty"`
}

// BasicConfigurationCondition general match rules
type BasicConfigurationCondition struct {
	ID         int               `json:"id,omitempty"`
	Namespace  string            `json:"namespace"`
	Name       string            `json:"name"`
	APIVersion string            `json:"apiVersion,omitempty"`
	Kind       string            `json:"kind,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// PodCondition pod matching rules
type PodCondition struct {
	ID         int               `json:"id,omitempty"`
	Namespace  string            `json:"namespace"`
	Name       string            `json:"name"`
	Phase      string            `json:"phase,omitempty"`
	APIVersion string            `json:"apiVersion,omitempty"`
	Kind       string            `json:"kind,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// JobCondition pod matching rules
type JobCondition struct {
	ID         int               `json:"id,omitempty"`
	Namespace  string            `json:"namespace"`
	Name       string            `json:"name"`
	Phase      string            `json:"phase,omitempty"`
	APIVersion string            `json:"apiVersion,omitempty"`
	Kind       string            `json:"kind,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	// The number of pending and running pods.
	// +optional
	Active int32 `json:"active,omitempty"`

	// The number of pods which reached phase Succeeded.
	// +optional
	Succeeded int32 `json:"succeeded,omitempty"`

	// The number of pods which reached phase Failed.
	// +optional
	Failed int32 `json:"failed,omitempty"`
}

//WatcherCondition watcher matching rules
type WatcherCondition struct {
	ID        int               `json:"id,omitempty"`
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels,omitempty"`
	// Condition of the Watcher
	// +optional
	Status string `json:"status,omitempty"`
}

//DeploymentCondition deployment matching rules
type DeploymentCondition struct {
	ID        int               `json:"id,omitempty"`
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels,omitempty"`
	// The number of running pods.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	//Custom logic of been ready
	// +optional
	// +kubebuilder:validation:Enum=true;True;TRUE;false;False;FALSE
	Ready string `json:"ready,omitempty"`
}

//StatefulSetCondition deployment matching rules
type StatefulSetCondition struct {
	ID        int               `json:"id,omitempty"`
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Labels    map[string]string `json:"labels,omitempty"`
	// The number of running pods.
	// +optional
	Replicas int32 `json:"replicas,omitempty"`

	//Custom logic of been ready
	// +optional
	// +kubebuilder:validation:Enum=true;True;TRUE;false;False;FALSE
	Ready string `json:"ready,omitempty"`
}

// WatcherStatus defines the observed state of Watcher
type WatcherStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	Status    string `json:"status"`
	Instanced bool   `json:"instanced"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status.status`
// +kubebuilder:printcolumn:name="Age",type="date",JSONPath=".metadata.creationTimestamp"
// Watcher is the Schema for the watchers API
type Watcher struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WatcherSpec   `json:"spec,omitempty"`
	Status WatcherStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// WatcherList contains a list of Watcher
type WatcherList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Watcher `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Watcher{}, &WatcherList{})
}
