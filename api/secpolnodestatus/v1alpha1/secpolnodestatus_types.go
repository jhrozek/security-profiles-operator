/*
Copyright 2021 The Kubernetes Authors.

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

// ProfileState defines the state that the profile is in. A profile in this context
// refers to a SeccompProfile or a SELinux profile, the states are shared between them
// as well as the management API.
type ProfileState string

const (
	// The profile is pending installation.
	ProfileStatePending ProfileState = "Pending"
	// The profile is being installed.
	ProfileStateInProgress ProfileState = "InProgress"
	// The profile was installed successfully.
	ProfileStateInstalled ProfileState = "Installed"
	// The profile is being removed and is currently terminating.
	ProfileStateTerminating ProfileState = "Terminating"
	// The profile couldn't be installed.
	ProfileStateError ProfileState = "Error"
	// When adding new statuses, remember to also adjust the LowerOfTwoStates function.
)

// Common labels of the node status objects.
const (
	// Identifies the policy by name so that the admin can list all node statuses for a certain policy.
	StatusToPolLabel = "spo.x-k8s.io/policy-name"
	// Identifies the node on which the policy is installed so that the admin can list policies per node.
	StatusToNodeLabel = "spo.x-k8s.io/node-name"
	// Allows the admin to filter out node statuses with a certain state (e.g. show me all that failed).
	StatusStateLabel = "spo.x-k8s.io/policy-state"
	// The kind of policy so that the admin can filter only e.g. all selinux policy statuses.
	StatusKindLabel = "spo.x-k8s.io/policy-kind"
)

// LowerOfTwoStates is used to figure out the "lowest common state" and is used to represent
// the overall status of a policy. The idea is that if, e.g. one in three policies is already
// installed, but the two others are pending, the overall state should be pending.
func LowerOfTwoStates(currentLowest, candidate ProfileState) ProfileState {
	orderedStates := make(map[ProfileState]int)
	orderedStates[ProfileStateError] = 0 // error must always have the lowest index
	orderedStates[ProfileStatePending] = 1
	orderedStates[ProfileStateInProgress] = 2
	orderedStates[ProfileStateInstalled] = 3
	orderedStates[ProfileStateTerminating] = 4

	if orderedStates[currentLowest] > orderedStates[candidate] {
		return candidate
	}
	return currentLowest
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecurityProfileNodeStatus is a per-node status of a security profile
// +kubebuilder:resource:shortName=spns
// +kubebuilder:printcolumn:name="Status",type=string,JSONPath=`.status`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
// +kubebuilder:printcolumn:name="Node",type=string,priority=10,JSONPath=`.nodeName`
type SecurityProfileNodeStatus struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	NodeName string       `json:"nodeName"`
	Status   ProfileState `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// SecurityProfileNodeStatusList contains a list of SecurityProfileNodeStatus.
type SecurityProfileNodeStatusList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SecurityProfileNodeStatus `json:"items"`
}

func init() { //nolint:gochecknoinits
	SchemeBuilder.Register(&SecurityProfileNodeStatus{}, &SecurityProfileNodeStatusList{})
}
