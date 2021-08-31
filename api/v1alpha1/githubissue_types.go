/*
Copyright 2021 Or Raz.

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

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GithubIssueSpec defines the desired state of GithubIssue
type GithubIssueSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Represent the github repo's URL - e.g https://github.com/rgolangh/dotfiles
	Repo string `json:"repo"`
	// The title of the issue
	Title string `json:"title"`
	// The issue's description
	Description string `json:"description"`
	// The issue's labels which are associated - array of strings or array of objects
	Labels string `json:"labels,omitempty"`
}

// GithubIssueStatus defines the observed state of GithubIssue
type GithubIssueStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// Represents the state of the real github issue. Could be open/closed or other text taken from the github API response.
	State string `json:"state"`
	// timestamp of the last time the state of the github issue was updated.
	LastUpdateTimestamp string `json:"lastUpdateTimestamp"`
	// The issue's number - used as primary key for finding if this is a new githubIssue
	Number int `json:"number,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// GithubIssue is the Schema for the githubissues API
type GithubIssue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GithubIssueSpec   `json:"spec,omitempty"`
	Status GithubIssueStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// GithubIssueList contains a list of GithubIssue
type GithubIssueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GithubIssue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&GithubIssue{}, &GithubIssueList{})
}
