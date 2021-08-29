# githubissues-operator
An exercise of creating an operator 'githubissues-operator', using (Operator SDK for GO) [https://sdk.operatorframework.io/docs/building-operators/golang/tutorial/]

## Introduction

## Tasks

+ implement deletion behaviour. A delete of the k8s object, triggers the github issue to be deleted.
+ tweak the resync period to every 1 minute.
+ store the github token you use in a secret and use it in the code by reading an env variable
+ add validation in the CRD level - an attempt to create a CRD with malformed 'repo' will fail
+ writing unit tests
    + documentation on how to test your reconcile code https://v0-19-x.sdk.operatorframework.io/docs/golang/legacy/unit-testing/
    + Those test cases should pass and cover:
        + failed attempt to create a real github issue
        + failed attempt to update an issue
        + create if issue not exist
        + close issues on delete
