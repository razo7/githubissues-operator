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
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/go-logr/logr"
	trainingv1alpha1 "github.com/razo7/githubissues-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

func DeleteCR(githubi trainingv1alpha1.GithubIssue, logger logr.Logger, ownerRepo string, token string) (trainingv1alpha1.GithubIssue, error) {
	// The object is being deleted
	var err error
	if ContainsString(githubi.GetFinalizers(), FinalizerName) { // https://book.kubebuilder.io/reference/using-finalizers.html
		githubi.Status.State = "closed"
		resp, err := CloseIssue(ownerRepo, githubi.Status.Number, token) // send an API call to change the state and closing time of the Github Issue

		if err != nil {
			return githubi, err
		}
		if resp.StatusCode != Ok_Code {
			logger.Info("Not valid repo- can't close the repo", "repo", ownerRepo)
			githubi.Status.State = Fail_Repo
			githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field

		} // if -status error
		if githubi.Status.State != Fail_Repo {
			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&githubi, FinalizerName)
			githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field

		}
		logger.Info("Closing", "issue number", githubi.Status.Number)
	}
	return githubi, err
	// return result, nil // Stop reconciliation as the item is being deleted
}

// postORpatchIsuue make a REST API to Githun.com to post or patch based on isPost parameter
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

// func IsHttpError(httpCode int, expectedCode int, ownerRepo string, error string){
// 				if httpCode != expectedCode {
// 				logger.Info(error, "repo", ownerRepo)
// 				githubi.Status.State = githubApi.Fail_Repo
// 				githubi.Status.LastUpdateTimestamp = time.Now().String()        // update LastUpdateTimestamp field
// 				if err := r.Client.Status().Update(ctx, &githubi); err != nil { // Update Vs. Patch -> https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#status
// 					logger.Error(err, "Can't update the K8s status state with the 'Fail repo' after CLOSE")
// 					return result, err
// 				}
// 				return result, nil
// 			} // if -status error
// }
// Helper functions to check and remove string from a slice of string. From https://book.kubebuilder.io/reference/using-finalizers.html
func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// func removeString(slice []string, s string) (result []string) {
// 	for _, item := range slice {
// 		if item == s {
// 			continue
// 		}
// 		result = append(result, item)
// 	}
// 	return
// }
