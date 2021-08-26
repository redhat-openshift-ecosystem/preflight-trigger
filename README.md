### preflight-trigger
Reference: _https://github.com/openshift/ci-tools/tree/master/cmd/cvp-trigger_  

Used by the hosted pipeline to trigger creation of or use of an OpenShift cluster provided by the OpenShift CI system 
managed by DPTP.  

_more details to be put here_  

preflight-trigger copies an existing periodic job which and uses it as a template to create a ProwJob resource. The ProwJob
resource is applied to the OpenShift CI cluster. The OpenShift CI cluster runs an instance of Prow
and the ProwJob is run once it has been applied to the cluster.

- [ ] Verify preflight-trigger works with cluster_claim
- [x] Verify preflight-trigger can interact with OpenShift CI
  - we need to get credentials: how-to > [documentation](https://docs.ci.openshift.org/docs/how-tos/use-registries-in-build-farm/#how-do-i-get-a-token-for-programmatic-access-to-the-central-ci-registry)
- [x] Determine where preflight-trigger code needs to be
  - not required to be in ci-tools
- [x] How do we trigger preflight-trigger
  - not "triggered" but will run as a binary getting/passing/creating vars for ProwJob
- [ ] Add env vars that preflight or other binaries that run in the tests section need
- [x] Is a Dockerfile for preflight-trigger required; based on cvp-trigger having one
  - yes
- [ ] Tekton Task for the Tekton Pipeline that runs preflight-trigger and subsequent requirements