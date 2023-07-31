# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.26

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

UNAME=$(shell uname -s)

VERSION ?= "$(shell cat VERSION)"

# Setting SHELL to bash allows bash commands to be executed by recipes.
# This is a requirement for 'setup-envtest.sh' in the test target.
# Options are set to exit when a recipe line exits non-zero or a piped command fails.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

# The help target prints out all targets with their descriptions organized
# beneath their categories. The categories are represented by '##@' and the
# target descriptions by '##'. The awk commands is responsible for reading the
# entire set of makefiles included in this invocation, looking for lines of the
# file as xyz: ## something, and then pretty-format the target and help. Then,
# if there's a line with ##@ something, that gets pretty-printed as a category.
# More info on the usage of ANSI control characters for terminal formatting:
# https://en.wikipedia.org/wiki/ANSI_escape_code#SGR_parameters
# More info on the awk command:
# http://linuxcommand.org/lc3_adv_awk.php

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

YAML_PREFIX=spec.versions[0].schema.openAPIV3Schema.properties.spec.properties
CRD_PRESERVE=x-kubernetes-preserve-unknown-fields = true

.PHONY: manifests
manifests: controller-gen yq ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.frontend.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.frontend.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.frontend.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.history.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.history.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.history.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.matching.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.matching.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.matching.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.internalFrontend.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.internalFrontend.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.internalFrontend.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.worker.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.worker.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.worker.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).ui.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).ui.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).ui.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).admintools.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).admintools.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).admintools.properties.overrides.properties.deployment.properties.spec.properties.template.properties.spec.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml

.PHONY: generate
generate: controller-gen api-docs ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: api-docs
api-docs: gen-crd-api-reference-docs ## Generate API reference documentation
	$(GEN_CRD_API_REFERENCE_DOCS) -api-dir=./api/v1beta1 -config=./hack/api/config.json -template-dir=./hack/api/template -out-file=./docs/api/v1beta1.md

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: lint
lint: golangci-lint ## Run golang-ci-lint against code.
	$(GOLANGCI_LINT) run ./...

.PHONY: test
test: manifests generate fmt vet envtest ## Run tests.
	KUBEBUILDER_ASSETS="$(shell $(ENVTEST) use $(ENVTEST_K8S_VERSION) --bin-dir $(LOCALBIN) -p path)" go test $$(go list ./... | grep -v /tests/e2e) -coverprofile cover.out

.PHONY: test-e2e
test-e2e: artifacts ## Run end2end tests.
	go test ./tests/e2e -v -timeout 60m -args "--v=4"

.PHONY: test-e2e-dev
test-e2e-dev: artifacts ## Run end2end tests on dev computer using kind.
	docker build -t temporal-operator .
	docker save temporal-operator > /tmp/temporal-operator.tar
	docker build -t example-worker-process ./examples/worker-process/helloworld
	docker save example-worker-process > /tmp/example-worker-process.tar
	OPERATOR_IMAGE_PATH=/tmp/temporal-operator.tar WORKER_PROCESS_IMAGE_PATH=/tmp/example-worker-process.tar go test ./tests/e2e -v -timeout 60m -args "--v=4"

.PHONY: ensure-license
ensure-license: go-licenser
	$(GO_LICENSER) -licensor "Alexandre VILAIN" -exclude api -exclude pkg/version -license ASL2 .

.PHONY: check-license
check-license: go-licenser
	$(GO_LICENSER) -licensor "Alexandre VILAIN" -exclude api -exclude pkg/version -license ASL2 -d .

.PHONY: dev-cluster
dev-cluster: kind-with-registry
	$(KIND_WITH_REGISTRY)

.PHONY: clean-dev-cluster
clean-dev-cluster:
	kind delete clusters kind

.PHONY: deploy-dev
deploy-dev: dev-cluster
	tilt up

##@ Build

.PHONY: build
build: generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build-dev
docker-build-dev: ## Build docker image with the manager.
	docker build -t temporal-operator .

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: install
install: manifests kustomize ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/crd |kubectl apply --server-side -f -

.PHONY: uninstall
uninstall: manifests kustomize ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/crd | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: deploy
deploy: manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	$(KUSTOMIZE) build config/default | kubectl apply -f -

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config. Call with ignore-not-found=true to ignore resource not found errors during deletion.
	$(KUSTOMIZE) build config/default | kubectl delete --ignore-not-found=$(ignore-not-found) -f -

.PHONY: artifacts
artifacts: kustomize
	mkdir -p $(RELEASE_PATH)
	$(KUSTOMIZE) build config/crd > ${RELEASE_PATH}/temporal-operator.crds.yaml
	$(KUSTOMIZE) build config/default > ${RELEASE_PATH}/temporal-operator.yaml

.PHONY: artifacts
helm: manifests helmify
	cat ${RELEASE_PATH}/temporal-operator.crds.yaml ${RELEASE_PATH}/temporal-operator.yaml | \
	$(HELMIFY) -crd-dir -image-pull-secrets -generate-defaults charts/temporal-operator

.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests -q
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --manifests --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: prepare-release
prepare-release: kustomize
	cd config/manager && $(KUSTOMIZE) edit set image ghcr.io/alexandrevilain/temporal-operator:v$(VERSION)
	$(MAKE) bundle

##@ Build Dependencies

RELEASE_PATH ?= $(shell pwd)/out/release/artifacts

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
HELMIFY ?= $(LOCALBIN)/helmify
KUSTOMIZE ?= $(LOCALBIN)/kustomize
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
CONTROLLER_GEN ?= GOFLAGS=-mod=mod $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GO_LICENSER ?= $(LOCALBIN)/go-licenser
GEN_CRD_API_REFERENCE_DOCS ?= $(LOCALBIN)/gen-crd-api-reference-docs
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
YQ ?= $(LOCALBIN)/yq
KIND_WITH_REGISTRY ?= $(LOCALBIN)/kind-with-registry

## Tool Versions
HELMIFY_VERSION ?= v0.4.5
KUSTOMIZE_VERSION ?= v4.5.7
OPERATOR_SDK_VERSION ?= 1.26.1
CONTROLLER_TOOLS_VERSION ?=  v0.11.3
GO_LICENSER_VERSION ?= v0.4.0
GEN_CRD_API_REFERENCE_DOCS_VERSION ?= 3f29e6853552dcf08a8e846b1225f275ed0f3e3b
GOLANGCI_LINT_VERSION ?= v1.52.2
YQ_VERSION ?= v4.30.6
KIND_WITH_REGISTRY_VERSION ?= 0.17.0

.PHONY: helmify
helmify: $(HELMIFY) ## Download helmify locally if necessary. If wrong version is installed, it will be removed before downloading.
$(HELMIFY): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/helmify version | grep -q $(HELMIFY_VERSION); then \
		echo "$(LOCALBIN)/helmifyversion is not expected $(HELMIFY_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/helmify; \
	fi
	test -s $(LOCALBIN)/helmify|| GOBIN=$(LOCALBIN) go install github.com/arttor/helmify/cmd/helmify@${HELMIFY_VERSION}

.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
$(KUSTOMIZE): $(LOCALBIN)
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v4@${KUSTOMIZE_VERSION}

.PHONY: operator-sdk
operator-sdk: $(OPERATOR_SDK)
$(OPERATOR_SDK): $(LOCALBIN)
	test -s $(LOCALBIN)/operator-sdk && $(LOCALBIN)/operator-sdk version | grep -q $(OPERATOR_SDK_VERSION) || \
	curl -sLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/v${OPERATOR_SDK_VERSION}/operator-sdk_`go env GOOS`_`go env GOARCH`
	@chmod +x $(OPERATOR_SDK)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(GOLANGCI_LINT)
	test -s $(LOCALBIN)/golangci-lint && $(LOCALBIN)/golangci-lint version | grep -q $(GOLANGCI_LINT_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: go-licenser
go-licenser: $(GO_LICENSER)
$(GO_LICENSER): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/elastic/go-licenser@$(GO_LICENSER_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
$(CONTROLLER_GEN): $(LOCALBIN)
	test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: gen-crd-api-reference-docs
gen-crd-api-reference-docs: $(GEN_CRD_API_REFERENCE_DOCS) ## Download gen-crd-api-reference-docs locally if necessary.
$(GEN_CRD_API_REFERENCE_DOCS): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/ahmetb/gen-crd-api-reference-docs@$(GEN_CRD_API_REFERENCE_DOCS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: yq
yq: $(YQ)
$(YQ): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/mikefarah/yq/v4@$(YQ_VERSION)

.PHONY: kind-with-registry
kind-with-registry: $(KIND_WITH_REGISTRY)
$(KIND_WITH_REGISTRY): $(LOCALBIN)
	curl -sLo $(KIND_WITH_REGISTRY) https://raw.githubusercontent.com/kubernetes-sigs/kind/v$(KIND_WITH_REGISTRY_VERSION)/site/static/examples/kind-with-registry.sh
	chmod +x $(KIND_WITH_REGISTRY)