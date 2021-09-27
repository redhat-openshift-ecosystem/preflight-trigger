shell=/usr/bin/env bash -o errexit

ifndef BIN_NAME
	override BIN_NAME = $(shell basename "$(PWD)")
endif

export CONTAINER_ENGINE ?= podman
export BIN_NAME
export REGISTRY ?= quay.io
export REGISTRY_NAMESPACE ?= opdev
export IMAGE_TAG ?= latest

binary-build:
	$(CONTAINER_ENGINE) run --rm -v $(PWD):/usr/src/$(BIN_NAME) -w /usr/src/$(BIN_NAME) -e GOOS=linux -e GOARCH=amd64 docker.io/library/golang:alpine go build -o build/$(BIN_NAME)

image-build:
	cd build && $(CONTAINER_ENGINE) build -t $(REGISTRY)/$(REGISTRY_NAMESPACE)/$(BIN_NAME):$(IMAGE_TAG) .
image-push:
	$(CONTAINER_ENGINE) push $(REGISTRY)/$(REGISTRY_NAMESPACE)/$(BIN_NAME):$(IMAGE_TAG)
