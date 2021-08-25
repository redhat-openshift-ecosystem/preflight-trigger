### preflight-trigger
Reference: _https://github.com/openshift/ci-tools/tree/master/cmd/cvp-trigger_  

Used by the hosted pipeline to trigger creation of or use of an OpenShift cluster provided by the OpenShift CI system 
managed by DPTP.  

_more details to be put here_  

preflight-trigger copies an existing periodic job which and uses it as a template to create a ProwJob resource. The ProwJob
resource is applied to the OpenShift CI cluster. The OpenShift CI cluster runs an instance of Prow
and the ProwJob is run once it has been applied to the cluster.