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

package github

import "net/http"

const (
	Fail_Repo     = "Fail repo"
	Created_Code  = 201 // https://docs.github.com/en/rest/reference/issues#create-an-issue
	Ok_Code       = 200
	FinalizerName = "batch.tutorial.kubebuilder.io/finalizer"
)

// A GithubRecieve struct to map the entire Response
type GithubRecieve struct {
	Repo        string `json:"url"` // or `json:"html_url"`
	Title       string `json:"title"`
	Description string `json:"body"` // It is called 'body' in the json file
	State       string `json:"state,omitempty"`
	Number      int    `json:"number,omitempty"`
}

// GithubSend - specify data fields for new github issue submission
type GithubSend struct {
	Title       string `json:"title,omitempty"`
	Body        string `json:"body,omitempty"`
	State       string `json:"state,omitempty"`
	ClosingTime string `json:"closed_at,omitempty"`
	// Labels      string `json:"labels` /// TODO: add label functionality

}

type Client interface {
	PostORpatchIsuue(ownerRepo string, title string, description string, number int, token string, isPost bool) (*http.Response, []byte, error)
	CloseIssue(ownerRepo string, issueNumber int, token string) (*http.Response, error)
}