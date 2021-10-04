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
	"os"
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
		GoodGithubIssueName     = "good-githubissue"
		BadGithubIssueName      = "bad-githubissue"
		ToDeleteGithubIssueName = "Delete-githubissue"
		GithubIssueNamespace    = "default"
		JobName                 = "test-job"

		timeout  = time.Second * 3
		duration = time.Second * 3
		interval = time.Millisecond * 250
	)
	var (
		githubIssue              trainingv1alpha1.GithubIssue
		goodGithubIssueLookupKey types.NamespacedName
		ctx                      context.Context
	)
	Context("GithubIssue Four Unit Tests", func() {
		BeforeEach(func() {
			goodGithubIssueLookupKey = types.NamespacedName{Name: GoodGithubIssueName, Namespace: GithubIssueNamespace}
			githubIssue = trainingv1alpha1.GithubIssue{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "batch.tutorial.kubebuilder.io/v1",
					Kind:       "GithubIssue",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      GoodGithubIssueName,
					Namespace: GithubIssueNamespace,
				},
				Spec: trainingv1alpha1.GithubIssueSpec{
					Repo:        "https://github.com/razo7/githubissues-operator",
					Title:       "K8s good Issue",
					Description: "Hi from testing K8s",
				},
				Status: trainingv1alpha1.GithubIssueStatus{
					State:               " ",
					LastUpdateTimestamp: " ",
				},
			} // githubIssue
			Expect(k8sClient).To(Not(BeNil()))
			err := k8sClient.Create(ctx, &githubIssue)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)
			}, timeout, interval).Should(Succeed())
		}) // BeforeEach - 1

		AfterEach(func() {
			err := k8sClient.Delete(ctx, &githubIssue)
			Expect(err).NotTo(HaveOccurred())
		}) // AfterEach - 1

		When("we test the github issue repo", func() {

			It("should succeed ", func() {
				By("use a good repo")
				Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())
				// Expect(k8sClient.Create(ctx, &githubIssue)).Should(Succeed())
			}) //it - test 1

			It("should fail due to a bad repo ", func() {
				By("use a bad repo")
				badGithubIssueLookupKey := types.NamespacedName{Name: BadGithubIssueName, Namespace: GithubIssueNamespace}
				badgithubIssue := trainingv1alpha1.GithubIssue{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch.tutorial.kubebuilder.io/v1",
						Kind:       "GithubIssue",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      BadGithubIssueName,
						Namespace: GithubIssueNamespace,
					},
					Spec: trainingv1alpha1.GithubIssueSpec{
						Repo:        "https://github.com/razo7/githubissues-operator2",
						Title:       "K8s badIssue",
						Description: "Not a good issue ",
					},
					Status: trainingv1alpha1.GithubIssueStatus{
						State:               " ",
						LastUpdateTimestamp: " ",
					},
				} // badgithubIssue
				Expect(k8sClient).To(Not(BeNil()))
				err := k8sClient.Create(ctx, &badgithubIssue)
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() error {
					return k8sClient.Get(ctx, badGithubIssueLookupKey, &badgithubIssue)
				}, timeout, interval).Should(Succeed())

				Eventually(func() bool {
					Expect(k8sClient.Get(ctx, badGithubIssueLookupKey, &badgithubIssue)).Should(Succeed())
					if badgithubIssue.Status.State != "Fail repo" {
						return false
					}
					return true
				}, timeout, interval).Should(BeTrue())
			}) // It - 2

		}) // when - 1

		When("we test updating an issue", func() {

			It("should set state to 'closed'", func() {
				By("change Status.State")
				Eventually(func() bool {
					By("change Status.State")
					Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())
					githubIssue.Status.State = "closed"
					err := k8sClient.Status().Update(ctx, &githubIssue)
					if err != nil {
						return false
					}
					return true
				}, timeout, interval).Should(BeTrue())
				Expect(githubIssue.Status.State).To(Equal("closed"))
			}) //it- test 3

			It("should set number to two ", func() {
				Eventually(func() bool {
					By("change Status.Number")
					Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())
					githubIssue.Status.Number = 2 // set the new issue number
					err := k8sClient.Status().Update(ctx, &githubIssue)
					if err != nil {
						return false
					}
					return true
				}, timeout, interval).Should(BeTrue())
				Expect(githubIssue.Status.Number).To(Equal(2))
			}) //it- test 4
		}) // when - 2

		When("we test update Github.com - POST REST API problems", func() {
			token := os.Getenv("GIT_TOKEN_GI")
			It("shouldn't succed due to bad repo", func() {
				Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())
				resp, _, err := postORpatchIsuue("razo/githubissues-operator", githubIssue, token, true)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(201))
			}) // it - test 5
			It("shouldn't succed due to bad token", func() {
				Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())

				resp, _, err := postORpatchIsuue("razo7/githubissues-operator", githubIssue, token+"something", true)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(201))
			}) // it - test 6

		}) // when -3

		When("we test update Github.com - PATCH REST API problems", func() {
			token := os.Getenv("GIT_TOKEN_GI")
			It("shouldn't succed due to bad repo", func() {
				Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())
				resp, _, err := postORpatchIsuue("razo/githubissues-operator", githubIssue, token, true)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(200))
			}) // it - test 7
			It("shouldn't succed due to bad token", func() {
				Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())

				resp, _, err := postORpatchIsuue("razo7/githubissues-operator", githubIssue, token+"something", true)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(200))
			}) // it - test 8

		}) // when -4

		When("creating an issue", func() {
			It("should check if the issue isn't currently exist", func() {
				By("check if number field is larger than zero")
				Eventually(func() bool {
					Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())
					if githubIssue.Status.Number > 0 { //TODO: maybe the opposite?
						return true
					}
					return false
				}, timeout, interval).Should(BeTrue())

				// Expect(githubIssue.Status.Number).To(BeNumerically(">", 0)) // another option
				// Expect(githubIssue.Status.Number).To(Not(Equal(0)))
			}) //it- test 9
		}) // when - 5

		When("we delete an issue", func() {
			It("should change state to close for this issue", func() {
				By("change the status to 'close' for each issue")
				goodGithubIssueLookupKeyToDelete := types.NamespacedName{Name: ToDeleteGithubIssueName, Namespace: GithubIssueNamespace}
				githubIssueToDelete := trainingv1alpha1.GithubIssue{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch.tutorial.kubebuilder.io/v1",
						Kind:       "GithubIssue",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      GoodGithubIssueName,
						Namespace: GithubIssueNamespace,
					},
					Spec: trainingv1alpha1.GithubIssueSpec{
						Repo:        "https://github.com/razo7/githubissues-operator",
						Title:       "K8s good Issue",
						Description: "Hi from testing K8s",
					},
					Status: trainingv1alpha1.GithubIssueStatus{
						State:               " ",
						LastUpdateTimestamp: " ",
					},
				} // githubIssue
				Expect(k8sClient).To(Not(BeNil()))
				err := k8sClient.Create(ctx, &githubIssueToDelete)
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() error {
					return k8sClient.Get(ctx, goodGithubIssueLookupKeyToDelete, &githubIssueToDelete)
				}, timeout, interval).Should(Succeed())
				// after creating the issue, now try to delete it
				Expect(k8sClient.Delete(ctx, &githubIssueToDelete)).Should(Succeed())
				Eventually(func() bool {
					if githubIssueToDelete.Status.State == "closed" {
						return true
					}
					return false
				}, timeout, interval).Should(BeTrue())
			}) // it - test 10
		}) // when - 6

	}) //context

})
