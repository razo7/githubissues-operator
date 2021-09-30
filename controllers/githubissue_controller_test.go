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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	// "github.com/google/go-cmp/cmp/internal/function"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	trainingv1alpha1 "github.com/razo7/githubissues-operator/api/v1alpha1"
)

var _ = Describe("GithubIssue controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		GithubIssueName      = "test-githubissue"
		GithubIssueNamespace = "default" // "githubissues-operator-system"
		JobName              = "test-job"

		timeout  = time.Second * 3
		duration = time.Second * 3
		interval = time.Millisecond * 250
	)
	var (
		githubIssue          trainingv1alpha1.GithubIssue
		githubIssueLookupKey types.NamespacedName
		ctx                  context.Context
	)
	Context("GithubIssue Four Unit Tests", func() {
		githubIssueLookupKey = types.NamespacedName{Name: GithubIssueName, Namespace: GithubIssueNamespace}
		ctx = context.Background()
		BeforeEach(func() {

			githubIssue = trainingv1alpha1.GithubIssue{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "batch.tutorial.kubebuilder.io/v1",
					Kind:       "GithubIssue",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      GithubIssueName,
					Namespace: GithubIssueNamespace,
				},
				Spec: trainingv1alpha1.GithubIssueSpec{
					Repo:        "https://github.com/razo7/githubissues-operator",
					Title:       "K8s Test Issue",
					Description: "Hi from testing K8s",
				},
				Status: trainingv1alpha1.GithubIssueStatus{
					State:               " ",
					LastUpdateTimestamp: " ",
				},
			}
			Expect(k8sClient).To(Not(BeNil()))
			err := k8sClient.Create(ctx, &githubIssue)
			Expect(err).NotTo(HaveOccurred())
		}) // BeforeEach - 1

		AfterEach(func() {
			err := k8sClient.Delete(ctx, &githubIssue)
			Expect(err).NotTo(HaveOccurred())
		})
		When("we test the github issue repo", func() {

			It("should succeed ", func() {
				By("use a real repo")
				Expect(k8sClient.Get(ctx, githubIssueLookupKey, &githubIssue)).Should(Succeed())
				// Expect(k8sClient.Create(ctx, &githubIssue)).Should(Succeed())
			}) //it - test 1

			It("should fail due to a bad repo ", func() {
				By("use a bad repo")
				badgithubIssue := &trainingv1alpha1.GithubIssue{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch.tutorial.kubebuilder.io/v1",
						Kind:       "GithubIssue",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      GithubIssueName,
						Namespace: GithubIssueNamespace,
					},
					Spec: trainingv1alpha1.GithubIssueSpec{
						Repo:        "https://github.com/razo7/githubissues-operator2",
						Title:       " ",
						Description: " ",
					},
					Status: trainingv1alpha1.GithubIssueStatus{
						State:               " ",
						LastUpdateTimestamp: " ",
						Number:              0,
					},
				}

				splitownerRepo := strings.Split(badgithubIssue.Spec.Repo, "github.com/") // extract from the repo url the repo's username and repo's name
				var ownerRepo string
				Expect(len(splitownerRepo)).Should(BeNumerically(">", 1))
				ownerRepo = splitownerRepo[1]
				issueData := GithubSend{Title: badgithubIssue.Spec.Title, Body: badgithubIssue.Spec.Description}
				//make it json
				jsonData, _ := json.Marshal(issueData)
				//creating client to set custom headers for Authorization
				client := &http.Client{}
				var apiURL string
				var req *http.Request
				apiURL = "https://api.github.com/repos/" + ownerRepo + "/issues"
				req, _ = http.NewRequest("POST", apiURL, bytes.NewReader(jsonData))
				token := "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"
				req.Header.Set("Authorization", "token "+token)
				resp, _ := client.Do(req)
				Expect(resp.StatusCode).To(Not(Equal(200))) // https://docs.github.com/en/rest/reference/issues#create-an-issue

			}) // It - 2

		}) // when - 1

		When("we test updating an issue", func() {
			It("should set state to 'closed'", func() {
				By("change Status.State")
				Eventually(func() bool {
					By("change Status.Number")
					Expect(k8sClient.Get(ctx, githubIssueLookupKey, &githubIssue)).Should(Succeed())
					githubIssue.Status.State = "closed"
					err := k8sClient.Status().Update(ctx, &githubIssue)
					if err != nil {
						return false
					}
					return true
				}, timeout, interval).Should(BeTrue())
				// Expect(githubIssue.Status.State).To(Equal("closed"))

				// Expect(k8sClient.Get(ctx, githubIssueLookupKey, &githubIssue)).Should(Succeed())
				// githubIssue.Status.State = "closed"
				// Expect(k8sClient.Status().Update(ctx, &githubIssue)).Should(Succeed())
			}) //it- test 3

			It("should set number to two ", func() {
				Eventually(func() bool {
					By("change Status.Number")
					Expect(k8sClient.Get(ctx, githubIssueLookupKey, &githubIssue)).Should(Succeed())
					githubIssue.Status.Number = 2 // set the new issue number
					err := k8sClient.Status().Update(ctx, &githubIssue)
					if err != nil {
						return false
					}
					return true
				}, timeout, interval).Should(BeTrue())
				// Expect(githubIssue.Status.Number).To(Equal(2))

				// Expect(k8sClient.Get(ctx, githubIssueLookupKey, &githubIssue)).Should(Succeed())
				// githubIssue.Status.Number = 2 // set the new issue number
				// Expect(k8sClient.Status().Update(ctx, &githubIssue)).Should(Succeed())
			}) //it- test 4
		}) // when - 2

		When("creating an issue", func() {
			It("should check if the issue isn't currently exist", func() {
				By("check if number field is larger than zero")
				Expect(k8sClient.Get(ctx, githubIssueLookupKey, &githubIssue)).Should(Succeed())
				// Expect(githubIssue.Status.Number).To(BeNumerically(">", 0)) // another option
				Expect(githubIssue.Status.Number).To(Not(Equal(0)))
			}) //it- test 5
		}) // when - 3

		// When("deleting all issues", func() {
		// 	It("should close all the issues", func() {
		// 		By("change the status to 'close' for each issue")
		// 		Expect(k8sClient.Delete(ctx, &githubIssue)).Should(Succeed())
		// 	}) // it - test 6
		// }) // when - 4

	}) //context

})
