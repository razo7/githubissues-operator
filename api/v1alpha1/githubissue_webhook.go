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
	"os"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

const (
	WebhookCertDir  = "/apiserver.local.config/certificates"
	WebhookCertName = "apiserver.crt"
	WebhookKeyName  = "apiserver.key"
)

// log is for logging in this package.
var githubissuelog = logf.Log.WithName("githubissue-resource")

// NodeMaintenanceValidator validates NodeMaintenance resources. Needed because we need a client for validation
// +k8s:deepcopy-gen=false
// type NodeMaintenanceValidator struct {
// 	client client.Client
// }

// var validator *NodeMaintenanceValidator

func (r *GithubIssue) SetupWebhookWithManager(mgr ctrl.Manager) error {
	// // init the validator!
	// validator = &NodeMaintenanceValidator{
	// 	client: mgr.GetClient(),
	// }

	// check if OLM injected certs
	certs := []string{filepath.Join(WebhookCertDir, WebhookCertName), filepath.Join(WebhookCertDir, WebhookKeyName)}
	certsInjected := true
	for _, fname := range certs {
		if _, err := os.Stat(fname); err != nil {
			certsInjected = false
			break
		}
	}
	if certsInjected {
		server := mgr.GetWebhookServer()
		server.CertDir = WebhookCertDir
		server.CertName = WebhookCertName
		server.KeyName = WebhookKeyName
	} else {
		githubissuelog.Info("OLM injected certs for webhooks not found")
	}

	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!

// TODO(user): change verbs to "verbs=create;update;delete" if you want to enable deletion validation.
//+kubebuilder:webhook:path=/validate-training-githubissues-v1alpha1-githubissue,mutating=false,failurePolicy=fail,sideEffects=None,groups=training.githubissues,resources=githubissues,verbs=create;update,versions=v1alpha1,name=vgithubissue.kb.io,admissionReviewVersions={v1,v1beta1}

var _ webhook.Validator = &GithubIssue{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *GithubIssue) ValidateCreate() error {
	githubissuelog.Info("validate create", "name", r.Name)

	// TODO(user): fill in your validation logic upon object creation.
	return nil
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *GithubIssue) ValidateUpdate(old runtime.Object) error {
	githubissuelog.Info("validate update", "name", r.Name)

	// TODO(user): fill in your validation logic upon object update.
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *GithubIssue) ValidateDelete() error {
	githubissuelog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
