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
	"encoding/json"
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	trainingv1alpha1 "github.com/razo7/githubissues-operator/api/v1alpha1"
	githubApi "github.com/razo7/githubissues-operator/github"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("GithubIssue controller", func() {

	// Define utility constants for object names and testing timeouts/durations and intervals.
	const (
		GoodGithubIssueName   = "good-githubissue"
		BadGithubIssueName    = "bad-githubissue"
		DeleteGithubIssueName = "delete-githubissue"
		GithubIssueNamespace  = "default"
		JobName               = "test-job"
		RepoName              = "razo7/githubissues-operator"
		RepoURL               = "https://github.com/razo7/githubissues-operator"
		Timeout               = time.Second * 4
		Interval              = time.Millisecond * 250
	)
	var (
		githubIssue              trainingv1alpha1.GithubIssue
		goodGithubIssueLookupKey types.NamespacedName
		ctx                      context.Context
		i                        int
		// gc                       githubApi.Client
	)
	Context("GithubIssue Four Unit Tests", func() {
		i = 0
		ctx = context.Background()
		token := os.Getenv("GIT_TOKEN_GI")
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
					Repo:        RepoURL,
					Title:       "K8s good Issue",
					Description: "Hi from testing K8s",
				},
				Status: trainingv1alpha1.GithubIssueStatus{
					State:               " ",
					LastUpdateTimestamp: " ",
				},
			} // githubIssue
			Expect(k8sClient).To(Not(BeNil()))
			i++
			githubIssue.Spec.Title = "Test " + fmt.Sprint(i)
			err := k8sClient.Create(ctx, &githubIssue)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() bool {
				err = k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)
				if err == nil && githubIssue.Status.Number > 0 {
					return true
				}
				return false
			}, Timeout, Interval).Should(BeTrue())
		}) // BeforeEach - 1

		AfterEach(func() {
			Expect(k8sClient).To(Not(BeNil()))
			err := k8sClient.Delete(ctx, &githubIssue)
			Expect(err).NotTo(HaveOccurred())
			Eventually(func() error {
				return k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)
			}, Timeout, Interval).ShouldNot(Succeed())
		}) // AfterEach - 1

		When("we test the github issue repo", func() {

			It("should succeed ", func() {
				By("use a good repo")
				Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())
			}) //it - test 1

			// It("should fail due to a bad repo ", func() {
			// 	By("use a bad repo")
			// 	badGithubIssueLookupKey := types.NamespacedName{Name: BadGithubIssueName, Namespace: GithubIssueNamespace}
			// 	badgithubIssue := trainingv1alpha1.GithubIssue{
			// 		TypeMeta: metav1.TypeMeta{
			// 			APIVersion: "batch.tutorial.kubebuilder.io/v1",
			// 			Kind:       "GithubIssue",
			// 		},
			// 		ObjectMeta: metav1.ObjectMeta{
			// 			Name:      BadGithubIssueName,
			// 			Namespace: GithubIssueNamespace,
			// 		},
			// 		Spec: trainingv1alpha1.GithubIssueSpec{
			// 			Repo:        RepoURL + "2",
			// 			Title:       "K8s badIssue",
			// 			Description: "Not a good issue ",
			// 		},
			// 		Status: trainingv1alpha1.GithubIssueStatus{
			// 			State:               " ",
			// 			LastUpdateTimestamp: " ",
			// 		},
			// 	} // badgithubIssue
			// 	Expect(k8sClient).To(Not(BeNil()))
			// 	err := k8sClient.Create(ctx, &badgithubIssue)
			// 	Expect(err).NotTo(HaveOccurred())
			// 	Eventually(func() error {
			// 		return k8sClient.Get(ctx, badGithubIssueLookupKey, &badgithubIssue)
			// 	}, Timeout, Interval).Should(Succeed())

			// 	Eventually(func() bool {
			// 		if badgithubIssue.Status.State != githubApi.Fail_Repo {
			// 			return false
			// 		}
			// 		return true
			// 	}, Timeout, Interval).Should(BeTrue())

			// }) // It - 2

		}) // when - 1

		// When("we test updating an issue", func() {

		// 	It("should set state to 'closed'", func() {
		// 		By("change Status.State")
		// 		Eventually(func() bool {
		// 			By("change Status.State")
		// 			Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())
		// 			githubIssue.Status.State = "closed"
		// 			err := k8sClient.Status().Update(ctx, &githubIssue)
		// 			if err != nil {
		// 				return false
		// 			}
		// 			return true
		// 		}, Timeout, Interval).Should(BeTrue())
		// 		Expect(githubIssue.Status.State).To(Equal("closed"))
		// 	}) //it- test 3

		// 	It("should set number to two ", func() {
		// 		Eventually(func() bool {
		// 			By("change Status.Number")
		// 			Expect(k8sClient.Get(ctx, goodGithubIssueLookupKey, &githubIssue)).Should(Succeed())
		// 			githubIssue.Status.Number = 2 // set the new issue number
		// 			err := k8sClient.Status().Update(ctx, &githubIssue)
		// 			if err != nil {
		// 				return false
		// 			}
		// 			return true
		// 		}, Timeout, Interval).Should(BeTrue())
		// 		Expect(githubIssue.Status.Number).To(Equal(2))
		// 	}) //it- test 4
		// }) // when - 2

		When("we test creating and deleting - REST API", func() {
			It("Post and Close - should succeed", func() {
				var issue githubApi.GithubRecieve // Storing the github issue from Github website
				resp, body, err := githubApi.PostORpatchIsuue(RepoName, githubIssue.Spec.Title, githubIssue.Spec.Description, githubIssue.Status.Number, token, true)

				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(201))
				// Expect(json.Unmarshal(body, &issue).BeNil())
				_ = json.Unmarshal(body, &issue)
				githubIssue.Status.Number = issue.Number
				resp, err = githubApi.CloseIssue(RepoName, githubIssue.Status.Number, token)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(200))
			}) // it - test 5
		})
		When("we test update Github.com - Bad POST REST API", func() {

			It("shouldn't succeed due to a bad repo", func() {
				resp, _, err := githubApi.PostORpatchIsuue(RepoName+"1", githubIssue.Spec.Title, githubIssue.Spec.Description, githubIssue.Status.Number, token, true)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(404))
			}) // it - test 6
			It("shouldn't succeed due to a bad token", func() {

				resp, _, err := githubApi.PostORpatchIsuue(RepoName, githubIssue.Spec.Title, githubIssue.Spec.Description, githubIssue.Status.Number, token+"something", true)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(401))
			}) // it - test 7

		}) // when -3

		When("we test update Github.com - Bad PATCH REST API", func() {
			It("shouldn't succeed due to a bad repo", func() {
				resp, _, err := githubApi.PostORpatchIsuue(RepoName+"1", githubIssue.Spec.Title, githubIssue.Spec.Description, githubIssue.Status.Number, token, false)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(404))
			}) // it - test 9
			It("shouldn't succeed due to a bad token", func() {
				resp, _, err := githubApi.PostORpatchIsuue(RepoName, githubIssue.Spec.Title, githubIssue.Spec.Description, githubIssue.Status.Number, token+"something", false)
				Expect(err).To(BeNil())
				Expect(resp.StatusCode).To(Equal(401))
			}) // it - test 10

		}) // when -4

		When("we create an issue", func() {
			It("should check if the issue isn't currently exist", func() {
				By("check if number field is larger than zero")
				Expect(githubIssue.Status.Number).To(BeNumerically(">", 0))
				Expect(githubIssue.Status.Number).To(Not(Equal(0))) // another option
			}) //it- test 11
		}) // when - 5

		When("we delete an issue", func() {
			It("should change state to close for this issue", func() {
				By("change the status to 'close' for each issue")
				deleteGithubIssueLookupKey := types.NamespacedName{Name: DeleteGithubIssueName, Namespace: GithubIssueNamespace}
				deletegithubIssue := trainingv1alpha1.GithubIssue{
					TypeMeta: metav1.TypeMeta{
						APIVersion: "batch.tutorial.kubebuilder.io/v1",
						Kind:       "GithubIssue",
					},
					ObjectMeta: metav1.ObjectMeta{
						Name:      DeleteGithubIssueName,
						Namespace: GithubIssueNamespace,
					},
					Spec: trainingv1alpha1.GithubIssueSpec{
						Repo:        RepoURL,
						Title:       "K8s delete Issue",
						Description: "a delete issue",
					},
					Status: trainingv1alpha1.GithubIssueStatus{
						State:               " ",
						LastUpdateTimestamp: " ",
					},
				} // githubIssue
				Expect(k8sClient).To(Not(BeNil()))
				err := k8sClient.Create(ctx, &deletegithubIssue)
				Expect(err).NotTo(HaveOccurred())
				Eventually(func() error {
					return k8sClient.Get(ctx, deleteGithubIssueLookupKey, &deletegithubIssue)
				}, Timeout, Interval).Should(Succeed())
				// after creating the issue, now try to delete it and wait
				Expect(k8sClient.Delete(ctx, &deletegithubIssue)).Should(Succeed())
				Eventually(func() error {
					return k8sClient.Get(ctx, deleteGithubIssueLookupKey, &deletegithubIssue)
				}, Timeout, Interval).ShouldNot(Succeed())
			}) // it - test 12
		}) // when - 6

	}) //context

})
