# githubissues-operator
An exercise of creating an operator 'githubissues-operator'.
The mission is at [Google Doc](https://docs.google.com/document/d/1z1bqlnBL8GO1FecJ0B2djncFzNPukOL1jw0E5K1xpgI/) and it suggest to use [GO's Operator SDK](https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/) to create the operator.

## Finished Work
+ Initalize an Operator with the desired Spec and Status (and some optional fields). Can be seen at 'api/v1alpha1/githubissue_types.go'.
    + Spec includes 'Repo', 'Title', 'Description', and 'Lables' fields.
    + Status includes 'State', 'LastUpdateTimestamp', and 'Number' fields.
+ Reconcile loop (controllers/githubissue_controller.go)
    + fetch K8 object
    + repo's basic validation (probably no needed as the check is suposed to be in CRD level)
    + gathering Github token from environment variable (by a secret)
    + checking if this issue is new, is the Number field zero?, if so then create an API call to create it with the basic parameters it has and set it's Number for next time (by changing the Status). If this is an old issue, Number field is greater than zero, then create an API call to update it with the basic parameters.
    + update the K8s object Status.State with Status.State from the Github.com issue and reconcile again after a minute.
+ CRD validation of the Spec.Repo field by cheking it's pattern with kubebuilder.
+ Writing unit tests for the following cases (they should pass and cover):
    + failed attempt to create a real github issue
    + create if issue not exist


## Ongoing Work
+ Writing unit tests for the following cases (they should pass and cover):
    + failed attempt to update an issue ?
    + close issues on delete


## Future Work
+ implement deletion behaviour. A delete of the k8s object, triggers the github issue to be deleted. I need to address it  with finalizers (pronbably delete as well my closeIssue function from 'controllers/githubissue_controller.go').
+ Running Webhook cluster
+ Enabling to create issues with 'Lables' field, or other useful fields.