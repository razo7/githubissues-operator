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
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	trainingv1alpha1 "github.com/razo7/githubissues-operator/api/v1alpha1"
)

var _ = Describe("GithubIssue controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		GithubIssueName      = "test-githubissue" // "test-githubissue"
		GithubIssueNamespace = "default"          // "githubissues-operator-system"
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
		BeforeEach(func() {
			ctx = context.Background()
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
		}) // BeforeEach - 1

		When("Test for OR", func() {
			It("Mini test for title spec", func() {
				By("make sure our Title string value was properly converted/handled.")
				Expect(githubIssue.Spec.Title).ShouldNot(Equal("K8s22 Test Issue"))
			}) //it - test 1
		}) // when - 1

		When("we test the github issue repo", func() {
			It("should succeed ", func() {
				By("use a real repo")
				Expect(k8sClient.Create(ctx, &githubIssue)).Should(Succeed())
			}) //it - test 2

			// It("should fail due to a bad repo ", func() {
			// 	By("use a bad repo")
			// 	badgithubIssue := &trainingv1alpha1.GithubIssue{
			// 		TypeMeta: metav1.TypeMeta{
			// 			APIVersion: "batch.tutorial.kubebuilder.io/v1",
			// 			Kind:       "GithubIssue",
			// 		},
			// 		ObjectMeta: metav1.ObjectMeta{
			// 			Name:      GithubIssueName,
			// 			Namespace: GithubIssueNamespace,
			// 		},
			// 		Spec: trainingv1alpha1.GithubIssueSpec{
			// 			Repo:        "ubissues-operator",
			// 			Title:       " ",
			// 			Description: " ",
			// 		},
			// 		Status: trainingv1alpha1.GithubIssueStatus{
			// 			State:               " ",
			// 			LastUpdateTimestamp: " ",
			// 			Number:              0,
			// 		},
			// 	}
			// 	Expect(len(strings.Split(badgithubIssue.Spec.Repo, "github.com/"))).To(BeNumerically("==", 1))
			// 	Eventually(func() error {
			// 		return k8sClient.Create(ctx, badgithubIssue)
			// 		// return k8sClient.Get(ctx, githubIssueLookupKey, badgithubIssue)
			// 	}, timeout, interval).Should(HaveOccurred())

			// 	// Expect(k8sClient.Create(ctx, badgithubIssue)).Should(Succeed())

			// }) // It - 3

		}) // when - 2

		When("we test updating an issue", func() {
			It("should set state to 'closed'", func() {
				By("change Status.State")
				Expect(k8sClient.Get(ctx, githubIssueLookupKey, &githubIssue)).Should(Succeed())
				githubIssue.Status.State = "closed"
				Expect(k8sClient.Status().Update(ctx, &githubIssue)).Should(Succeed())
			}) //it- test 4

			It("should set number to two ", func() {
				Expect(k8sClient.Get(ctx, githubIssueLookupKey, &githubIssue)).Should(Succeed())
				githubIssue.Status.Number = 2 // set the new issue number
				Expect(k8sClient.Status().Update(ctx, &githubIssue)).Should(Succeed())

			}) //it- test 5
		}) // when - 3

		When("creating an issue", func() {
			It("should check if the issue isn't currently exist", func() {
				By("check if number field is larger than zero")
				Expect(k8sClient.Get(ctx, githubIssueLookupKey, &githubIssue)).Should(Succeed())
				// Expect(githubIssue.Status.Number).To(BeNumerically(">", 0)) // another option
				Expect(githubIssue.Status.Number).To(Not(Equal(0)))
			}) //it- test 6
		}) // when - 4

		// When("deleting all issues", func() {
		// 	It("should close all the issues", func() {
		// 		By("change the status to 'close' for each issue")
		// 		Expect(k8sClient.Delete(ctx, &githubi)).Should(Succeed())
		// 	}) // it - test 7
		// }) // when - 5

	}) //context

})
