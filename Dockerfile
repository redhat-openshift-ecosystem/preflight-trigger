ARG quay_expiration=never
ARG release_tag=0.0.0
ARG ARCH=amd64
ARG OS=linux

FROM docker.io/golang:1.25 AS builder
ARG quay_expiration
ARG release_tag
ARG ARCH
ARG OS

# Build the preflight binary
COPY . /go/src/preflight-trigger
WORKDIR /go/src/preflight-trigger
RUN make build


FROM registry.access.redhat.com/ubi9/ubi-micro:latest
ARG quay_expiration
ARG release_tag
ARG ARCH
ARG OS

# Metadata
LABEL name="Preflight Trigger" \
      vendor="Red Hat, Inc." \
      maintainer="Red Hat OpenShift Ecosystem" \
      version="1" \
      summary="Provides the OpenShift Preflight Trigger tool." \
      description="Preflight Trigger calls prow jobs to provision OpenShift clusters." \
      url="https://github.com/redhat-openshift-ecosystem/preflight-trigger" \
      release=${release_tag}


# Define that tags should expire after 1 week. This should not apply to versioned releases.
LABEL quay.expires-after=${quay_expiration}

# Fetch the build image Architecture
LABEL ARCH=${ARCH}
LABEL OS=${OS}

# Add preflight-trigger binary
COPY --from=builder /go/src/preflight-trigger/preflight-trigger /usr/local/bin/preflight-trigger

#copy license
COPY LICENSE /licenses/LICENSE

CMD ["preflight-trigger"]
