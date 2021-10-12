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
	githubApi "github.com/razo7/githubissues-operator/github"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	// "sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	// metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	// "fmt"

	"strings"
	"time"
)

// GithubIssueReconciler reconciles a GithubIssue object
type GithubIssueReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
	// GithubClient github.Client
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
	var sucess bool

	if err := r.Get(ctx, req.NamespacedName, &githubi); err != nil {
		if githubi.Status.Number == 0 { // if we can't fetch the issue after deleting it then stop reconcile
			return result, nil
		}
		if apierrors.IsNotFound(err) {
			logger.Error(err, "Can't fetch Kubernetes github object")
			return result, nil
		}
		return result, err
	}
	if githubi.Status.Number > 0 {
		firstRun = false // chnaged into false once it has a number (ID)
	}
	ownerRepo := strings.Split(githubi.Spec.Repo, "github.com/")[1] // extract the repo's username, and repo's name from the repo's url
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
		if githubi, err = githubApi.DeleteIssue(githubi, ownerRepo); err != nil {
			logger.Error(err, "Closing issue")
			return result, err
		}
		logger.Info("Successful close", "number", githubi.Status.Number)
	} // if we need to delete the issue

	// If my K8s GithubIssue doesn't have an ID then create a new GithubIssue and update it's ID
	// Otherwiese I have already created it earlier and it had an ID and I just update it's description
	logger.Info("After fetching K8s issue", "number", githubi.Status.Number, "state", githubi.Status.State)

	if githubi.Status.State != githubApi.Fail_Repo { // if the repo is valid

		if githubi.Status.Number == 0 { // Zero = uninitialized field
			if githubi, err, _ = githubApi.GetIssue(githubi, ownerRepo, "POST"); err != nil {
				logger.Error(err, "Creating Issue")
				return result, err
			}
			logger.Info("Successful creation", "number", githubi.Status.Number, "state", githubi.Status.State)

		} else {
			// if githubi.Spec.Description != issue.Description { // update the description (if needed).
			if githubi, err, sucess = githubApi.GetIssue(githubi, ownerRepo, "GET"); err != nil {
				logger.Error(err, "Updating Issue")
				return result, err
			}
			if sucess {
				logger.Info("Successful update", "number", githubi.Status.Number, "description", githubi.Spec.Description)
			}
		} // else
	} else {
		// remove our finalizer from the list and update it.
		controllerutil.RemoveFinalizer(&githubi, githubApi.FinalizerName)
	}

	// Update the client status or the whole client (for register/unregister finalizer)
	if firstRun {
		if err := r.Client.Status().Update(ctx, &githubi); err != nil { // Update Vs. Patch -> https://sdk.operatorframework.io/docs/building-operators/golang/references/client/#status
			logger.Error(err, "Can't update Client's status")
			return result, err
		}
	}
	if firstRun || githubi.Status.State == "closed" {
		if err := r.Update(ctx, &githubi); err != nil {
			logger.Error(err, "Can't update reconcile - for register/unregister finalizer")
			return result, err
		}
	}

	logger.Info("End reconcile", "number", githubi.Status.Number, "state", githubi.Status.State)
	return ctrl.Result{RequeueAfter: 60 * time.Second}, nil // tweak the resync period to every 1 minute.
} // Reconcile

// SetupWithManager sets up the controller with the Manager.
func (r *GithubIssueReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&trainingv1alpha1.GithubIssue{}).
		Complete(r)
}
