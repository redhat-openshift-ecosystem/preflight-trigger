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

```shell
#latest run Thu Aug 26 17:08:16 CDT 2021

ovpn-113-106:preflight-trigger mrhillsman$ go run main.go -job-config-path hcp-dptp-hackery/jobs/redhat-openshift-ecosystem/preflight/redhat-openshift-ecosystem-preflight-ocp-4.8-periodics.yaml -job-name period
ic-ci-redhat-openshift-ecosystem-preflight-ocp-4.8-preflight-cluster-claim-aws -prow-config-path hcp-dptp-hackery/core-services/prow/02_config/_config.yaml -ocp-version 4.8.7 -output-path . -dry-run
INFO[0000] Cannot use full value, will truncate.         id= job=periodic-ci-redhat-openshift-ecosystem-preflight-ocp-4.8-preflight-cluster-claim-aws key=prow.k8s.io/job maybeTruncated=periodic-ci-redhat-openshift-ecosystem-preflight-ocp-4.8-prefli value=periodic-ci-redhat-openshift-ecosystem-preflight-ocp-4.8-preflight-cluster-claim-aws
apiVersion: prow.k8s.io/v1
kind: ProwJob
metadata:
  annotations:
    prow.k8s.io/context: ""
    prow.k8s.io/job: periodic-ci-redhat-openshift-ecosystem-preflight-ocp-4.8-preflight-cluster-claim-aws
  creationTimestamp: null
  labels:
    created-by-prow: "true"
    prow.k8s.io/context: ""
    prow.k8s.io/job: periodic-ci-redhat-openshift-ecosystem-preflight-ocp-4.8-prefli
    prow.k8s.io/type: periodic
  name: ab2f7958-06b9-11ec-ae4f-dca9048abc0d
spec:
  agent: kubernetes
  cluster: build02
  decoration_config:
    censor_secrets: true
    gcs_configuration:
      bucket: origin-ci-test
      default_org: openshift
      default_repo: origin
      mediaTypes:
        log: text/plain
      path_strategy: single
    gcs_credentials_secret: gce-sa-credentials-gcs-publisher
    grace_period: 1h0m0s
    resources:
      clonerefs:
        limits:
          memory: 3Gi
        requests:
          cpu: 100m
          memory: 500Mi
      initupload:
        limits:
          memory: 200Mi
        requests:
          cpu: 100m
          memory: 50Mi
      place_entrypoint:
        limits:
          memory: 100Mi
        requests:
          cpu: 100m
          memory: 25Mi
      sidecar:
        limits:
          memory: 2Gi
        requests:
          cpu: 100m
          memory: 250Mi
    skip_cloning: true
    timeout: 4h0m0s
    utility_images:
      clonerefs: gcr.io/k8s-prow/clonerefs:v20210825-bc8cae85fb
      entrypoint: gcr.io/k8s-prow/entrypoint:v20210825-bc8cae85fb
      initupload: gcr.io/k8s-prow/initupload:v20210825-bc8cae85fb
      sidecar: gcr.io/k8s-prow/sidecar:v20210825-bc8cae85fb
  job: periodic-ci-redhat-openshift-ecosystem-preflight-ocp-4.8-preflight-cluster-claim-aws
  namespace: ci
  pod_spec:
    containers:
    - args:
      - --gcs-upload-secret=/secrets/gcs/service-account.json
      - --hive-kubeconfig=/secrets/hive-hive-credentials/kubeconfig
      - --image-import-pull-secret=/etc/pull-secret/.dockerconfigjson
      - --report-credentials-file=/etc/report/credentials
      - --secret-dir=/secrets/ci-pull-credentials
      - --target=preflight-cluster-claim-aws
      - --input-hash=CLUSTER_TYPEawsOCP_VERSION4.8.7
      command:
      - ci-operator
      env:
      - name: CLUSTER_TYPE
        value: aws
      - name: OCP_VERSION
        value: 4.8.7
      image: ci-operator:latest
      imagePullPolicy: Always
      name: ""
      resources:
        requests:
          cpu: 10m
      volumeMounts:
      - mountPath: /secrets/ci-pull-credentials
        name: ci-pull-credentials
        readOnly: true
      - mountPath: /secrets/gcs
        name: gcs-credentials
        readOnly: true
      - mountPath: /secrets/hive-hive-credentials
        name: hive-hive-credentials
        readOnly: true
      - mountPath: /etc/pull-secret
        name: pull-secret
        readOnly: true
      - mountPath: /etc/report
        name: result-aggregator
        readOnly: true
    serviceAccountName: ci-operator
    volumes:
    - name: ci-pull-credentials
      secret:
        secretName: ci-pull-credentials
    - name: hive-hive-credentials
      secret:
        secretName: hive-hive-credentials
    - name: pull-secret
      secret:
        secretName: registry-pull-credentials
    - name: result-aggregator
      secret:
        secretName: result-aggregator
  report: true
  type: periodic
status:
  startTime: "2021-08-26T22:05:05Z"
  state: triggered

ovpn-113-106:preflight-trigger mrhillsman$ 
```

- [ ] What does a successful ProwJob look like?