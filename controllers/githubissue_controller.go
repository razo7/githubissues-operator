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

package controllers

//	How to create this repo? Follow the next two lines
//	operator-sdk init --domain githubissues --repo github.com/razo7/githubissues-operator --owner "Or Raz"
//	operator-sdk create api --group training --version v1alpha1 --kind GithubIssue --resource --controller
//  run 'kubectl create secret generic mysecret --from-literal=github-token=PUBLIC_GITHUB_TOKEN' after 'make deploy'
//  and 'kubectl delete secret mysecret' to delete it

import (
	"context"

	trainingv1alpha1 "github.com/razo7/githubissues-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	// "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"bytes"
	"encoding/json"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"

	// "fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

// GithubIssueReconciler reconciles a GithubIssue object
type GithubIssueReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

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

//+kubebuilder:rbac:groups=training.githubissues,resources=githubissues,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=training.githubissues,resources=githubissues/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=training.githubissues,resources=githubissues/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the GithubIssue object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.9.2/pkg/reconcile
func (r *GithubIssueReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// fetch K8s GithubIssue - inspired by NHC controller
	logger := r.Log.WithValues("githubssue", req.NamespacedName)
	githubi := trainingv1alpha1.GithubIssue{} // Empty GithubIssue
	result := ctrl.Result{}                   // Empty Result
	if err := r.Get(ctx, req.NamespacedName, &githubi); err != nil {
		if githubi.Status.Number == 0 { // if we can't fetch the issue after deleting it then stop it (we got here due to the last update)
			return result, nil
		}
		if apierrors.IsNotFound(err) {
			logger.Error(err, "Can't fetch Kubernetes github object", "object") //TODO: Do I reach to here?
			return result, nil
		}
		return result, err
	}

	ownerRepo := strings.Split(githubi.Spec.Repo, "github.com/")[1] // extract the repo's username, and repo's name from the repo's url
	// Good link for using secrets -> https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables
	token := os.Getenv("GIT_TOKEN_GI") // store the github token you use in a secret and use it in the code by reading an env variable

	// register finalizer once the CR enters the reconcile
	myFinalizerName := "batch.tutorial.kubebuilder.io/finalizer"
	if githubi.Status.LastUpdateTimestamp == "" {
		if !containsString(githubi.GetFinalizers(), myFinalizerName) {
			controllerutil.AddFinalizer(&githubi, myFinalizerName) // registering our finalizer.
			if err := r.Update(ctx, &githubi); err != nil {
				return result, err
			}
		}
	}
	// examine DeletionTimestamp to determine if object is under deletion
	if !githubi.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if containsString(githubi.GetFinalizers(), myFinalizerName) { // https://book.kubebuilder.io/reference/using-finalizers.html
			err := closeIssue(ownerRepo, githubi, token) // send an API call to change the state and closing time of the Github Issue
			if err != nil {
				logger.Error(err, "Can't close the repo's issue- API problem")
				return result, err
			}

			// remove our finalizer from the list and update it.
			controllerutil.RemoveFinalizer(&githubi, myFinalizerName)
			if err := r.Update(ctx, &githubi); err != nil {
				logger.Error(err, "Can't close the repo's issue - update problem")
				return result, err
			}
		}
		return result, nil // Stop reconciliation as the item is being deleted
	}

	// If my K8s GithubIssue doesn't have an ID then create a new GithubIssue and update it's ID
	// Otherwiese I have already created it earlier and it had an ID and I just update it's description
	logger.Info("Looking for K8s YAML ID", "githubi.Status.Number", githubi.Status.Number, "githubi.Status.State", githubi.Status.State)
	var issue GithubRecieve // Storing the github issue from Github website
	var jsonBody []byte     // Storing the github issue from Github website in a JSON format

	if githubi.Status.State != "Fail repo" { // if the repo is valid

		if githubi.Status.Number == 0 { // Zero = uninitialized field
			resp, body, err := postORpatchIsuue(ownerRepo, githubi, token, true)
			jsonBody = body
			if err != nil {
				logger.Error(err, "Can't create new repo's issue")
				return result, err
			}
			if resp.StatusCode != 201 { // https://docs.github.com/en/rest/reference/issues#create-an-issue
				logger.Info("Not valid repo- change the state", "repo", ownerRepo)
				githubi.Status.State = "Fail repo"
				githubi.Status.LastUpdateTimestamp = time.Now().String()        // update LastUpdateTimestamp field
				if err := r.Client.Status().Update(ctx, &githubi); err != nil { // Update Vs. Patch -> https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#status
					logger.Error(err, "Can't update the K8s status state with the 'Fail repo' after POST")
					return result, err
				}
				return result, nil
			} // if -status error
			if err := json.Unmarshal(jsonBody, &issue); err != nil {
				logger.Error(err, "Can't parse the githubIssue - json.Unmarshal error - after post")
				return result, err
			}

			githubi.Status.Number = issue.Number                     // set the new issue number
			githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field
			logger.Info("Get K8s YAML ID", "githubi.Status.Number", githubi.Status.Number, "githubi.Status.State", githubi.Status.State)
			if err := r.Client.Status().Update(ctx, &githubi); err != nil {
				logger.Error(err, "Can't update the K8s github issue number from the issue in Github.com")
				return result, err
			}

		} else { // update the description (if needed).
			resp, body, err := postORpatchIsuue(ownerRepo, githubi, token, false)
			if err != nil {
				logger.Error(err, "Can't update the description in repo's issue")
				return result, err
			}
			if resp.StatusCode != 200 {
				logger.Info("Bad repo, there is no repo -", ownerRepo, " in github.com")
				githubi.Status.State = "Fail repo"
				if err := r.Client.Status().Update(ctx, &githubi); err != nil {
					logger.Error(err, "Can't update the K8s status state with the 'Fail repo' after PATCH")
					return result, err
				}
			} // if -status error
			jsonBody = body
		}

		if err := json.Unmarshal(jsonBody, &issue); err != nil {
			logger.Error(err, "Can't parse the githubIssue - json.Unmarshal error - after post/patch")
			return result, err
		}
		if githubi.Spec.Description != issue.Description { // Is there a change in the description?
			githubi.Status.State = issue.State
			githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field
			if err := r.Client.Status().Update(ctx, &githubi); err != nil {
				logger.Error(err, "Can't update the K8s status state with the real github issue, maybe because the github issue has already been closed")
				return result, err
			}
		}

		logger.Info("End", "githubi.Status.Number", githubi.Status.Number, "githubi.Status.State", githubi.Status.State)
		return ctrl.Result{RequeueAfter: 60 * time.Second}, nil // tweak the resync period to every 1 minute.
	} else {
		// TODO: should I delete the resource
		return result, nil
	} // if -Fail repo
} // Reconcile

// Helper functions to check and remove string from a slice of string. From https://book.kubebuilder.io/reference/using-finalizers.html
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func removeString(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

// postORpatchIsuue make a REST API to Githun.com to post or patch based on isPost parameter
func postORpatchIsuue(ownerRepo string, gituhubi trainingv1alpha1.GithubIssue, token string, isPost bool) (*http.Response, []byte, error) {
	issueData := GithubSend{Title: gituhubi.Spec.Title, Body: gituhubi.Spec.Description}
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
		apiURL = "https://api.github.com/repos/" + ownerRepo + "/issues/" + strconv.Itoa(gituhubi.Status.Number)
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

func closeIssue(ownerRepo string, gituhubi trainingv1alpha1.GithubIssue, token string) error {
	apiURL := "https://api.github.com/repos/" + ownerRepo + "/issues/" + strconv.Itoa(gituhubi.Status.Number)
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
	return err
} // closeIssue

// SetupWithManager sets up the controller with the Manager.
func (r *GithubIssueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&trainingv1alpha1.GithubIssue{}).
		Complete(r)
}

// layout := "2006-01-02 15:04:05 +0300 IDT m=+5"
// lastUpdateTime, err := time.Parse(layout, githubi.Status.LastUpdateTimestamp)
// if err != nil {
// 	logger.Error(err, "Can't parse the time")
// 	return result, err
// }
// currTime := time.Now()
// fmt.Println("Hi print 1")
// fmt.Println("curr time = ", currTime, "\nlast time = ", githubi.Status.LastUpdateTimestamp, "\ndifference = ", currTime.Sub(lastUpdateTime))
// fmt.Println("diff = ", currTime.Sub(lastUpdateTime), "\ndiff mili= ", currTime.Sub(lastUpdateTime).Milliseconds(), "\ndifference nano = ", currTime.Sub(lastUpdateTime).Nanoseconds())
// if (currTime.Sub(lastUpdateTime)).Seconds() < 15500 { // don't update the description if it the last update was less than a second, otherwise it results in an endless loop
// 	return result, nil
// }
