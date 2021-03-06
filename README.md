# githubissues-operator
An operator which creats, updates and deletes Github issues using [GO's Operator SDK](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/).

The reconcile loop uses REST API (GET/POST/PATCH) calls for updating Github.com issues, and the task is from [Google Doc](https://docs.google.com/document/d/1z1bqlnBL8GO1FecJ0B2djncFzNPukOL1jw0E5K1xpgI/).
## Features
+ The Operator's Spec and Status are (api/v1alpha1/githubissue_types.go):
    + Spec includes Repo, Title, Description fields.
    + Status includes State, LastUpdateTimestamp, and Number fields.
+ CRD validation of the Spec.Repo field by cheking it's pattern with kubebuilder.
+ The reconcile loop (controllers/githubissue_controller.go):
    + fetch K8 object
    + gathers Github token from environment variable (by a secret)
    + register finalizer
    + delete the CR if it is needed
    + create CR if that's the first run of reconcile, otherwise fetch existing githubIssue from Github.com and update it's description (if it is different)
    + at the end update the status of K8s object or the reconcile object if the finalizer has been resistered/unregistered.
    + reconcile again after a minute.
+ Writing unit tests for the following cases (api/v1alpha1/githubissue_types_test.go):
    + failed attempt to create a real github issue
    + create if issue not exist
    + failed attempt to update an issue
    + close issue on delete
+ Creation/deletion of the k8s object triggers the github issue to be created/deleted.

## Ongoing Work
+ Running Webhook cluster

## Usage
+ To test the unit tests - run `make test` in the main directory.
+ To run the reconcile
    + locally - run `make install run`
    + distributly (on a cluster) - run `make deploy IMG=quay.io/oraz/githubissueimage:1.1.2`
    and then run `kubectl create secret generic mysecret --from-literal=github-token=PUBLIC_GITHUB_TOKEN -n githubissues-operator-system` where PUBLIC_GITHUB_TOKEN is the github 
+ To test creation or deletion of githubIssue CR - run oc(openshift)/kubectl(K8s) or create/delete `oc create -f config/samples/my_test_samples/ex_X.yaml` where X can be 1 to 5 with five CR samples.

