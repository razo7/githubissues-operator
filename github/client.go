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

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	trainingv1alpha1 "github.com/razo7/githubissues-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

////////////////////////////////////////////////////////////////  Client FUNCTIONS  ////////////////////////////////////////////////////////////////

// HttpHandler check for a mismatch between httpCode and the expected code, and update the Stauts accordingly
func HttpHandler(githubi trainingv1alpha1.GithubIssue, logger logr.Logger, httpCode int, expectedCode int, ownerRepo string) (trainingv1alpha1.GithubIssue, error) {
	var err error
	var errName string
	switch httpCode {
	case 404:
		errName = ", Not Found"
	case 401:
		errName = ", Unauthorized Client"
	default:
		errName = ""
	}
	if httpCode != expectedCode {
		err = fmt.Errorf("Not valid repo - %s, received bad HTTP response code %d%s", ownerRepo, httpCode, errName)
		githubi.Status.State = Fail_Repo
		githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field
	} // if -status error
	return githubi, err
}

// DeleteCR check if FinalizerName has been registered, make a REST API call to close the Issue, check http response and eventually unregister FinalizerName
func DeleteIssue(githubi trainingv1alpha1.GithubIssue, logger logr.Logger, ownerRepo string, token string) (trainingv1alpha1.GithubIssue, error, string) {
	var err error
	if ContainsString(githubi.GetFinalizers(), FinalizerName) { // https://book.kubebuilder.io/reference/using-finalizers.html
		githubi.Status.State = "closed"
		// send an API call to change the state and closing time of the Github Issue
		resp, err := CloseIssue(ownerRepo, githubi.Status.Number, token)
		if err != nil {
			return githubi, err, "REST"
		}
		if githubi, err = HttpHandler(githubi, logger, resp.StatusCode, Ok_Code, ownerRepo); err != nil {
			return githubi, err, "TOKEN"
		} else {
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&githubi, FinalizerName)
			githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field
		}
	}
	return githubi, err, ""
	// return result, nil // Stop reconciliation as the item is being deleted
}

// CreateIssue creates a githubissue and chcecks for errors of REST, bad token/repo or JSON and eventually update the K8s object
func CreateIssue(githubi trainingv1alpha1.GithubIssue, logger logr.Logger, ownerRepo string, token string) (trainingv1alpha1.GithubIssue, error, string) {
	resp, body, err := PostORpatchIsuue(ownerRepo, githubi.Spec.Title, githubi.Spec.Description, githubi.Status.Number, token, true)
	if err != nil {
		return githubi, err, "REST"
	}
	if githubi, err = HttpHandler(githubi, logger, resp.StatusCode, Created_Code, ownerRepo); err != nil {
		return githubi, err, "TOKEN"
	}
	if err := json.Unmarshal(body, &issue); err != nil {
		return githubi, err, "JSON"
	}
	githubi.Status.Number = issue.Number // set the new issue number
	githubi.Status.State = issue.State
	githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field
	return githubi, err, ""
}

// UpdateIssue updates the githubissue and chcecks for errors of REST, bad token/repo
func UpdateIssue(githubi trainingv1alpha1.GithubIssue, logger logr.Logger, ownerRepo string, token string) (trainingv1alpha1.GithubIssue, error, string) {
	resp, _, err := PostORpatchIsuue(ownerRepo, githubi.Spec.Title, githubi.Spec.Description, githubi.Status.Number, token, false)
	if err != nil {
		return githubi, err, "REST"
	}
	if githubi, err = HttpHandler(githubi, logger, resp.StatusCode, Ok_Code, ownerRepo); err != nil {
		return githubi, err, "TOKEN"
	}
	return githubi, err, "" //empty string = no errors
}

////////////////////////////////////////////////////////////////  REST API FUNCTIONS  ////////////////////////////////////////////////////////////////

// postORpatchIsuue make a Post or Patch REST API call to Github.com
func PostORpatchIsuue(ownerRepo string, title string, description string, number int, token string, isPost bool) (*http.Response, []byte, error) {
	issueData := GithubSend{Title: title, Body: description}
	//make it json
	jsonData, _ := json.Marshal(issueData)
	//creating client to set custom headers for Authorization
	client := &http.Client{}
	var apiURL string
	var req *http.Request
	if isPost {
		apiURL = "https://api.github.com/repos/" + ownerRepo + "/issues"
		req, _ = http.NewRequest("POST", apiURL, bytes.NewReader(jsonData))
	} else {
		apiURL = "https://api.github.com/repos/" + ownerRepo + "/issues/" + strconv.Itoa(number)
		req, _ = http.NewRequest("PATCH", apiURL, bytes.NewReader(jsonData))
	}
	req.Header.Set("Authorization", "token "+token)
	resp, err := client.Do(req)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("fmt - Hello from postORpatchIsuue, isPost =", isPost, ", status = ", resp.StatusCode, ", http.StatusCreated = ", http.StatusCreated, " and err = ", err) // fmt option
	return resp, body, err
} // postORpatchIsuue

// CloseIssue make a Patch REST API call to Github.com to change the state of the githubissue which will close it
func CloseIssue(ownerRepo string, issueNumber int, token string) (*http.Response, error) {
	apiURL := "https://api.github.com/repos/" + ownerRepo + "/issues/" + strconv.Itoa(issueNumber)
	issueData := GithubSend{State: "closed", ClosingTime: time.Now().Format("2006-01-02 15:04:05")} // formating time -> https://stackoverflow.com/questions/33119748/convert-time-time-to-string
	//make it json
	jsonData, _ := json.Marshal(issueData)
	//creating client to set custom headers for Authorization
	client := &http.Client{}
	// fmt.Println("issueData ", issueData, ", jsonData", jsonData)
	req, _ := http.NewRequest("PATCH", apiURL, bytes.NewReader(jsonData))
	req.Header.Set("Authorization", "token "+token)
	resp, err := client.Do(req)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	return resp, err
} // closeIssue

////////////////////////////////////////////////////////////////  Other FUNCTIONS  ////////////////////////////////////////////////////////////////

// Helper functions to check and remove string from a slice of string. From https://book.kubebuilder.io/reference/using-finalizers.html
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}
