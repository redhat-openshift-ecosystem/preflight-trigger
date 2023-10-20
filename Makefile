shell=/usr/bin/env bash -o errexit

ifndef BIN_NAME
	override BIN_NAME = $(shell basename "$(PWD)")
endif

export CONTAINER_ENGINE ?= podman
export BIN_NAME
export REGISTRY ?= quay.io
export REGISTRY_NAMESPACE ?= opdev
export RELEASE_TAG ?= "0.0.0"


.PHONY: binary-build
binary-build:
	$(CONTAINER_ENGINE) run --rm -v $(PWD):/usr/src/$(BIN_NAME) -w /usr/src/$(BIN_NAME) -e GOOS=linux -e GOARCH=amd64 docker.io/library/golang:alpine go build -o $(BIN_NAME)

.PHONY: image-build
image-build:
	$(CONTAINER_ENGINE) build --build-arg release_tag=$(RELEASE_TAG) -t $(REGISTRY)/$(REGISTRY_NAMESPACE)/$(BIN_NAME):$(RELEASE_TAG) .

.PHONY: image-push
image-push:
	$(CONTAINER_ENGINE) push $(REGISTRY)/$(REGISTRY_NAMESPACE)/$(BIN_NAME):$(RELEASE_TAG)

.PHONY: build
build:
	CGO_ENABLED=0 go build -o $(BIN_NAME) main.go
	@ls build | grep -e '^preflight-trigger$$' &> /dev/null
