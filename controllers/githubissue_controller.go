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

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	// log "sigs.k8s.io/controller-runtime/pkg/log"
	trainingv1alpha1 "github.com/razo7/githubissues-operator/api/v1alpha1"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors" 

	"bytes"
	"encoding/json"
	"net/http"
	
	"os"
	"io/ioutil"
)

// A Response struct to map the Entire Response
type Response struct {

	Repo                string      `json:"repo"`
	Title               string      `json:"title"`
	Description         string      `json:"description"`	
	// State               string      `json:"state,omitempty"`
	// LastUpdateTimestamp string      `json:"updated_at"`	
    // Name    string    `json:"name"`
}

// GithubIssueReconciler reconciles a GithubIssue object
type GithubIssueReconciler struct {
	client.Client
	Log           logr.Logger
	Scheme *runtime.Scheme	
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
	// _ = log.FromContext(ctx) // DO I need it?
	// your logic here
	// fetch GithubIssue - inspired by NHC controller
	log := r.Log.WithValues("GithubIssue", req.NamespacedName)
	githubi := trainingv1alpha1.GithubIssue{}
	result := ctrl.Result{}
	if err := r.Get(ctx, req.NamespacedName, &githubi) ; err != nil {
		log.Error(err, "failed fetching Kubernetes github object", "object", githubi)
		if apierrors.IsNotFound(err) {
			return result, nil
		}
		return result, err
	} // Update	githubi with the Kubernetes github object
	
	ownerRepo := "razo7/githubissues-operator"
	token := "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"
	token = os.Getenv("GIT_TOKEN_GI")
	resp, body, err := r.getIsuues(ownerRepo, token)
	if resp.StatusCode != http.StatusCreated && err != nil{
		log.Error(err, "unable to list issues due to error %d\n", resp.StatusCode)
		if resp.StatusCode == 404 {
			return result, nil
		} else {
			return result, err // not a NotExist error
		}
	}// if error

	var issues []Response // Inspariation -> https://tutorialedge.net/golang/consuming-restful-api-with-go/
    json.Unmarshal(body, &issues)
	// Find the desired github issue
	var isExist bool  // default value is false
	for _, issue := range issues{
	// for i := 0; i < len(issues.respIssues); i++ {
		if githubi.Spec.Title == issue.Title {
			if githubi.Spec.Description != issue.Description { // update the description (if needed).
				// i.spec.description = githubi.Spec.Description 
				_, err = patchIsuue(ownerRepo, githubi.Spec.Title, githubi.Spec.Description, token)
				if err == nil {
					log.Error(err, "failed updating the description in repo's issue")
				}
			}
			isExist = true
			break
		}
	}// for
	if ! isExist{ //no matching issue, then create an issue 
		resp, err = postIsuue(ownerRepo, githubi.Spec.Title, githubi.Spec.Description, token)
		if  err != nil{
			issues = append(issues, Response{ Repo: ownerRepo, Title: githubi.Spec.Title, Description: githubi.Spec.Description})
		} else {
			log.Error(err, "failed creating new repo's issue")
		}
	}
	return result, nil
}
func (r *GithubIssueReconciler) getIsuues(ownerRepo string, token string) (*http.Response, []byte, error){
	apiURL := "https://api.github.com/repos/" + ownerRepo + "/issues"
	client := &http.Client{}
	req, _ := http.NewRequest("GET", apiURL, nil)
	req.Header.Set("Authorization", "token "+token)
	resp, err := client.Do(req)
	if err != nil {
		r.Log.Error(err, "No Repo was found")
	}
	if resp.Body != nil {
		defer resp.Body.Close()
	}

	body, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		r.Log.Error(readErr, "Can't read repp's issues")
	}

	return resp, body, err
}// Fetch all github issues

// MyNewIssue - specify data fields for new github issue submission
type MyNewIssue struct {
	Title string `json:"title"`
	Description  string `json:"description"`
}

func postIsuue(ownerRepo string, title string, description string, token string) (*http.Response, error) {
	apiURL := "https://api.github.com/repos/" + ownerRepo + "/issues"
	issueData := MyNewIssue{Title: title, Description: description}
	//make it json
	jsonData, _ := json.Marshal(issueData)
	//creating client to set custom headers for Authorization
	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(jsonData))
	req.Header.Set("Authorization", "token "+token)
	resp, err := client.Do(req)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	return resp, err
}// Create a github issue

func patchIsuue(ownerRepo string, title string, description string, token string) (*http.Response, error) {
	apiURL := "https://api.github.com/repos/" + ownerRepo + "/issues"
	issueData := MyNewIssue{Title: title, Description: description}
	//make it json
	jsonData, _ := json.Marshal(issueData)
	//creating client to set custom headers for Authorization
	client := &http.Client{}
	req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(jsonData))
	req.Header.Set("Authorization", "token "+token)
	resp, err := client.Do(req)
	if resp.Body != nil {
		defer resp.Body.Close()
	}
	return resp, err
}// Create a github issue




// SetupWithManager sets up the controller with the Manager.
func (r *GithubIssueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&trainingv1alpha1.GithubIssue{}).
		Complete(r)
}