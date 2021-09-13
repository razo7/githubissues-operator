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

import (
	"context"
	"fmt"

	trainingv1alpha1 "github.com/razo7/githubissues-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
	Number      int    `json:"number,omitempty"` // TODO can it be number or integer?  -> https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.0.md#schemaObject
	// LastUpdateTimestamp string      `json:"updated_at"`
	// Name    string    `json:"name"`
}

// GithubSend - specify data fields for new github issue submission
type GithubSend struct {
	Title       string `json:"title,omitempty"`
	Body        string `json:"body,omitempty"`
	State       string `json:"state,omitempty"`
	ClosingTime string `json:"closed_at,omitempty"`
	// Labels 	string 	`json:"labels` /// TODO: add label functionality

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
	ownerRepo := "razo7/githubissues-operator" // TODO: Should be -> ownerRepo := githubi.Spec.Repo
	// Good link for using secrets -> https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables
	// Run 'kubectl create secret generic mysecret --from-literal=github-token='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'' after 'make deploy'
	token := os.Getenv("GIT_TOKEN_GI") // store the github token you use in a secret and use it in the code by reading an env variable
	var myBody []byte

	githubi := trainingv1alpha1.GithubIssue{}
	result := ctrl.Result{}
	if err := r.Get(ctx, req.NamespacedName, &githubi); err != nil {
		// Can get it because of delete or something else?

		logger.Info("Hi new Issue!", "object", githubi, "req", req.NamespacedName, "githubi.ObjectMeta", githubi.ObjectMeta)
		logger.Info("Hi 2!", "githubi.TypeMeta.String()", githubi.TypeMeta.String(), "githubi.ObjectMeta.GetDeletionTimestamp()", githubi.ObjectMeta.GetDeletionTimestamp())
		logger.Info("Hi 3!", "githubi.ObjectMeta.GetCreationTimestamp()", githubi.ObjectMeta.GetCreationTimestamp(), "githubi.ObjectMeta.GetClusterName()", githubi.ObjectMeta.GetClusterName())

		if apierrors.IsNotFound(err) {
			logger.Error(err, "Can't fetch Kubernetes github object", "object", githubi, "githubi.Status.Number", githubi.Status.Number)
			return result, nil // tweak the resync period to every 1 minute.
		}
		// if githubi.ObjectMeta.creationTimestamp == nil {
		// if the creationTimestamp is null then we should delete this issue from the website
		logger.Info("before close issue")
		body, err := closeIssue(ownerRepo, githubi, token)
		myBody = body
		if err != nil {
			logger.Error(err, "Can't close the repo's issue")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err // tweak the resync period to every 1 minute.
		}
		// }
		logger.Info("I closed the repo's issue")
		return ctrl.Result{RequeueAfter: 60 * time.Second}, nil // tweak the resync period to every 1 minute.
	} // Update	githubi with the Kubernetes github object

	// If my K8s GithubIssue doesn't has an ID then create a new GithubIssue and update it's ID
	// Otherwiese I have already created it earlier and it had an ID and I just update it's description
	logger.Info("Looking for K8s YAML ID", "githubi.Status.Number", githubi.Status.Number, "githubi.Status.State", githubi.Status.State)
	var issue GithubRecieve
	if githubi.Status.Number == 0 { // Zero = uninitialized field
		body, err := postIsuue(ownerRepo, githubi, token)
		myBody = body
		if err != nil {
			logger.Error(err, "Can't create new repo's issue")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}

		if err := json.Unmarshal(myBody, &issue); err != nil {
			logger.Error(err, "Can't parse the githubIssue - json.Unmarshal error")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}

		githubi.Status.Number = issue.Number // Get the new issue number
		logger.Info("Get K8s YAML ID", "githubi.Status.Number", githubi.Status.Number, "githubi.Spec.Title", githubi.Spec.Title, "githubi.Status.State", githubi.Status.State, "githubi.Spec.Repo", githubi.Spec.Repo)
		if err := r.Client.Status().Update(ctx, &githubi); err != nil { // Update Vs. Patch -> https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#status
			logger.Error(err, "Can't update the K8s github issue number from website github issue")
			return ctrl.Result{RequeueAfter: 10 * time.Second}, err
		}
		// update ID and close end reconcile
		return result, err // tweak the resync period to every 1 minute.
	} else { // update the description (if needed).
		body, err := patchIsuue(ownerRepo, githubi, token)
		myBody = body
		if err != nil {
			logger.Error(err, "Can't update the description in repo's issue")
			return result, err
		}
	}

	if err := json.Unmarshal(myBody, &issue); err != nil {
		logger.Error(err, "Can't parse the githubIssue - json.Unmarshal error")
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}
	githubi.Status.State = issue.State                              // TODO: Is it the right place to update the issue?
	if err := r.Client.Status().Update(ctx, &githubi); err != nil { // Update Vs. Patch -> https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#status
		logger.Error(err, "Can't update the K8s status state with the real github issue, maybe because the github issue has already been closed")
		return result, err
	}

	logger.Info("End", "githubi.Status.Number", githubi.Status.Number, "githubi.Status.State", githubi.Status.State)
	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil // tweak the resync period to every 1 minute.
}

func postIsuue(ownerRepo string, gituhubi trainingv1alpha1.GithubIssue, token string) ([]byte, error) {
	apiURL := "https://api.github.com/repos/" + ownerRepo + "/issues"
	issueData := GithubSend{Title: gituhubi.Spec.Title, Body: gituhubi.Spec.Description}
	//make it json
	jsonData, _ := json.Marshal(issueData)
	//creating client to set custom headers for Authorization
	client := &http.Client{}
	fmt.Println("issueData ", issueData, ", jsonData", jsonData, "token ", token)
	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(jsonData))
	req.Header.Set("Authorization", "token "+token)
	resp, err := client.Do(req)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("fmt - Hello from postIsuue, status = ", resp.StatusCode, " and http.StatusCreated = ", http.StatusCreated, " and err = ", err) // fmt option
	return body, err
} // Create a github issue

func patchIsuue(ownerRepo string, gituhubi trainingv1alpha1.GithubIssue, token string) ([]byte, error) {
	apiURL := "https://api.github.com/repos/" + ownerRepo + "/issues/" + strconv.Itoa(gituhubi.Status.Number)
	issueData := GithubSend{Title: gituhubi.Spec.Title, Body: gituhubi.Spec.Description}
	//make it json
	jsonData, _ := json.Marshal(issueData)
	//creating client to set custom headers for Authorization
	client := &http.Client{}
	req, _ := http.NewRequest("PATCH", apiURL, bytes.NewReader(jsonData))
	req.Header.Set("Authorization", "token "+token)
	resp, err := client.Do(req)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("fmt - Hello from patchIsuue, status = ", resp.StatusCode, " and http.StatusCreated = ", http.StatusCreated, " and err = ", err, "body = ", resp.Body) // fmt option
	return body, err
} // Create a github issue

func closeIssue(ownerRepo string, gituhubi trainingv1alpha1.GithubIssue, token string) ([]byte, error) {
	apiURL := "https://api.github.com/repos/" + ownerRepo + "/issues/" + strconv.Itoa(gituhubi.Status.Number)
	issueData := GithubSend{State: "closed", ClosingTime: time.Now().Format("2006-01-02 15:04:05")} // TODO: Maybe it is "close", formating time -> https://stackoverflow.com/questions/33119748/convert-time-time-to-string
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
	body, _ := ioutil.ReadAll(resp.Body)
	// fmt.Println("fmt - Hello from closeIsuue, status = ", resp.StatusCode, " and http.StatusCreated = ", http.StatusCreated, " and err = ", err, "body = ", resp.Body) // fmt option
	return body, err
} // Create a github issue

// SetupWithManager sets up the controller with the Manager.
func (r *GithubIssueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&trainingv1alpha1.GithubIssue{}).
		Complete(r)
}
