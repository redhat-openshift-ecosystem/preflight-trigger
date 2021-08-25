module github.com/redhat-openshift-ecosystem/preflight-trigger

go 1.16

replace (
	github.com/Azure/go-autorest => github.com/Azure/go-autorest v14.2.0+incompatible
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.6.0

	// forked version that compiles with k8s
	github.com/bombsimon/logrusr => github.com/stevekuznetsov/logrusr v1.1.1-0.20210709145202-301b9fbb8872
	github.com/containerd/containerd => github.com/containerd/containerd v0.2.10-0.20180716142608-408d13de2fbb

	github.com/docker/docker => github.com/openshift/moby-moby v1.4.2-0.20190308215630-da810a85109d

	// Forked version that disables diff trimming
	github.com/google/go-cmp => github.com/alvaroaleman/go-cmp v0.5.7-0.20210615160450-f8688cd5aaa0

	github.com/moby/buildkit => github.com/dmcgowan/buildkit v0.0.0-20170731200553-da2b9dc7dab9
	github.com/openshift/api => github.com/openshift/api v0.0.0-20201120165435-072a4cd8ca42
	github.com/openshift/client-go => github.com/openshift/client-go v0.0.0-20200521150516-05eb9880269c
	github.com/openshift/library-go => github.com/openshift/library-go v0.0.0-20200527213645-a9b77f5402e3
	k8s.io/client-go => k8s.io/client-go v0.21.0
	k8s.io/component-base => k8s.io/component-base v0.21.0
	k8s.io/kubectl => k8s.io/kubectl v0.21.0
)

require (
	github.com/ghodss/yaml v1.0.0
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/afero v1.4.1
	k8s.io/api v0.21.2
	k8s.io/test-infra v0.0.0-20210823175823-85d839e08600
)

require github.com/openshift/ci-tools v0.0.0-20210825141814-c151c81eb844