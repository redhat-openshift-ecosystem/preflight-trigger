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

.PHONY: vet
vet:
	go vet ./...

.PHONY: tidy
tidy:
	go mod tidy
	git diff --exit-code

.PHONY: fmt
fmt: gofumpt
	${GOFUMPT} -l -w .
	git diff --exit-code

.PHONY: lint
lint: golangci-lint ## Run golangci-lint linter checks.
	$(GOLANGCI_LINT) run

.PHONY: cover
cover:
	go test -v \
	 $$(go list ./...) \
	 -race \
	 -cover -coverprofile=coverage.out

GOFUMPT = $(shell pwd)/bin/gofumpt
gofumpt: ## Download envtest-setup locally if necessary.
	$(call go-install-tool,$(GOFUMPT),mvdan.cc/gofumpt@latest)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-install-tool
@[ -f $(1) ] || { \
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
}
endef

GOLANGCI_LINT = $(shell pwd)/bin/golangci-lint
GOLANGCI_LINT_VERSION ?= v1.52.2
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT):
	$(call go-install-tool,$(GOLANGCI_LINT),github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION))

