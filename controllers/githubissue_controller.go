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
//  run 'kubectl create secret generic mysecret --from-literal=github-token=PUBLIC_GITHUB_TOKEN -n githubissues-operator-system'
//  after 'make deploy'
//  and 'kubectl delete secret mysecret' to delete it

import (
	"context"

	trainingv1alpha1 "github.com/razo7/githubissues-operator/api/v1alpha1"
	"github.com/razo7/githubissues-operator/github"
	githubApi "github.com/razo7/githubissues-operator/github"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	// "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"encoding/json"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "fmt"

	"os"
	"strings"
	"time"
)

// GithubIssueReconciler reconciles a GithubIssue object
type GithubIssueReconciler struct {
	client.Client
	Log          logr.Logger
	Scheme       *runtime.Scheme
	GithubClient github.Client
}

//+kubebuilder:rbac:groups=training.githubissues,resources=githubissues,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=training.githubissues,resources=githubissues/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=training.githubissues,resources=githubissues/finalizers,verbs=update
//+kubebuilder:rbac:groups=redhat.com,resources=githubissues/finalizers,verbs=get;create;update;patch;delete
// For watching the resource and implementing finalizers ->
//  https://developers.redhat.com/blog/2020/09/11/5-tips-for-developing-kubernetes-operators-with-the-new-operator-sdk#:~:text=adding%20rbac%20permissions%20with%20go

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
	firstRun := true
	var err error
	// gc githubApi.Client
	if err := r.Get(ctx, req.NamespacedName, &githubi); err != nil {
		if githubi.Status.Number == 0 { // if we can't fetch the issue after deleting it then stop it (we got here due to the last update)
			return result, nil
		}
		if apierrors.IsNotFound(err) {
			logger.Error(err, "Can't fetch Kubernetes github object", "object")
			return result, nil
		}
		return result, err
	}
	if githubi.Status.Number > 0 {
		firstRun = false
	}
	ownerRepo := strings.Split(githubi.Spec.Repo, "github.com/")[1] // extract the repo's username, and repo's name from the repo's url
	// Good link for using secrets -> https://kubernetes.io/docs/concepts/configuration/secret/#using-secrets-as-environment-variables
	token := os.Getenv("GIT_TOKEN_GI") // store the github token you use in a secret and use it in the code by reading an env variable

	// register finalizer once the CR has been created

	if githubi.Status.LastUpdateTimestamp == "" {
		if !githubApi.ContainsString(githubi.GetFinalizers(), githubApi.FinalizerName) {
			controllerutil.AddFinalizer(&githubi, githubApi.FinalizerName) // registering our finalizer.
			githubi.Status.LastUpdateTimestamp = time.Now().String()
		}
	} // if - register finalizer
	// examine DeletionTimestamp to determine if object is under deletion
	if !githubi.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is being deleted
		if githubi, err = githubApi.DeleteCR(githubi, logger, ownerRepo, token); err != nil {
			logger.Error(err, "Can't close the repo's issue- API problem")
			return result, err
		}
	} // if we need to delete

	// If my K8s GithubIssue doesn't have an ID then create a new GithubIssue and update it's ID
	// Otherwiese I have already created it earlier and it had an ID and I just update it's description
	logger.Info("After fetching K8s", "githubi.Status.Number", githubi.Status.Number, "githubi.Status.State", githubi.Status.State)
	var issue githubApi.GithubRecieve // Storing the github issue from Github website
	var jsonBody []byte               // Storing the github issue from Github website in a JSON format

	if githubi.Status.State != githubApi.Fail_Repo { // if the repo is valid

		if githubi.Status.Number == 0 { // Zero = uninitialized field
			resp, body, err := githubApi.PostORpatchIsuue(ownerRepo, githubi.Spec.Title, githubi.Spec.Description, githubi.Status.Number, token, true)
			jsonBody = body
			if err != nil {
				logger.Error(err, "Can't create new repo's issue")
				return result, err
			}
			if resp.StatusCode != githubApi.Created_Code { // https://docs.github.com/en/rest/reference/issues#create-an-issue
				logger.Info("Not a valid repo- changing the state", "repo", ownerRepo)
				githubi.Status.State = githubApi.Fail_Repo
				githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field
			} // if -status error
			if githubi.Status.State != githubApi.Fail_Repo { // if the repo is valid
				if err := json.Unmarshal(jsonBody, &issue); err != nil {
					logger.Error(err, "Can't parse the githubIssue - json.Unmarshal error - after post")
					return result, err
				}

				githubi.Status.Number = issue.Number // set the new issue number
				githubi.Status.State = issue.State
				githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field
				logger.Info("Get K8s YAML ID", "githubi.Status.Number", githubi.Status.Number, "githubi.Status.State", githubi.Status.State)
			}
		} else { // update the description (if needed).
			resp, body, err := githubApi.PostORpatchIsuue(ownerRepo, githubi.Spec.Title, githubi.Spec.Description, githubi.Status.Number, token, false)
			if err != nil {
				logger.Error(err, "Can't update the description in repo's issue")
				return result, err
			}
			if resp.StatusCode != githubApi.Ok_Code {
				logger.Info("Bad repo, there is no repo -", ownerRepo, " in github.com")
				githubi.Status.State = githubApi.Fail_Repo
				githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field
			} // if -status error
			jsonBody = body
		} // else
		if githubi.Status.State != githubApi.Fail_Repo { // if the repo is valid

			if err := json.Unmarshal(jsonBody, &issue); err != nil {
				logger.Error(err, "Can't parse the githubIssue - json.Unmarshal error - after post/patch")
				return result, err
			}
			if githubi.Spec.Description != issue.Description { // Is there a change in the description?
				githubi.Spec.Description = issue.Description
				githubi.Status.State = issue.State
				githubi.Status.LastUpdateTimestamp = time.Now().String() // update LastUpdateTimestamp field
			}
		}
	} else {
		// remove our finalizer from the list and update it.
		controllerutil.RemoveFinalizer(&githubi, githubApi.FinalizerName)
		// return result, nil
	} // if -Fail repo
	if githubi.Status.State == "open" {
		if err := r.Client.Status().Update(ctx, &githubi); err != nil { // Update Vs. Patch -> https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#status
			logger.Error(err, "Can't status")
			return result, err
		}
	}
	if firstRun || githubi.Status.State == "closed" {
		if err := r.Update(ctx, &githubi); err != nil {
			logger.Error(err, "Can't update the K8s status state with the real github issue, maybe because the github issue has already been closed")
			return result, err
		}
	}

	logger.Info("End", "githubi.Status.Number", githubi.Status.Number, "githubi.Status.State", githubi.Status.State)
	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil // tweak the resync period to every 1 minute.
} // Reconcile

// SetupWithManager sets up the controller with the Manager.
func (r *GithubIssueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&trainingv1alpha1.GithubIssue{}).
		Complete(r)
}
