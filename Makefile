# ENVTEST_K8S_VERSION refers to the version of kubebuilder assets to be downloaded by envtest binary.
ENVTEST_K8S_VERSION = 1.23

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

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

.PHONY: manifests
manifests: controller-gen ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

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
test: manifests generate fmt vet ## Run tests.
	go test $(go list ./... | grep -v /tests/e2e) -coverprofile cover.out

.PHONY: test-e2e
test-e2e: artifacts ## Run end2end tests.
	go test ./tests/e2e -v -timeout 20m

.PHONY: test-e2e-dev
test-e2e-dev: artifacts ## Run end2end tests on dev computer using kind.
	docker build -t temporal-operator .
	docker save temporal-operator > /tmp/temporal-operator.tar
	OPERATOR_IMAGE_PATH=/tmp/temporal-operator.tar go test ./tests/e2e -v -args "--v=4"

.PHONY: ensure-license
ensure-license: go-licenser
	$(GO_LICENSER) -licensor "Alexandre VILAIN" -exclude internal/forked -exclude api -license ASL2 .

.PHONY: check-license
check-license: go-licenser
	$(GO_LICENSER) -licensor "Alexandre VILAIN" -exclude internal/forked -exclude api -license ASL2 -d .

.PHONY: deploy-dev
deploy-dev: docker-build-dev manifests kustomize ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	kind load docker-image temporal-operator
	$(KUSTOMIZE) build config/dev | kubectl apply -f -
	kubectl rollout restart deployment temporal-operator-controller-manager -n temporal-system

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
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

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

.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests -q
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --manifests --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: release
release:
	cd config/manager && $(KUSTOMIZE) edit set image ghcr.io/alexandrevilain/temporal-operator:v$(VERSION)
	$(MAKE) bundle

##@ Build Dependencies

RELEASE_PATH ?= $(shell pwd)/out/release/artifacts

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GO_LICENSER ?= $(LOCALBIN)/go-licenser
GEN_CRD_API_REFERENCE_DOCS ?= $(LOCALBIN)/gen-crd-api-reference-docs
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint

## Tool Versions
KUSTOMIZE_VERSION ?= v4.5.5
OPERATOR_SDK_VERSION ?= 1.23.0
CONTROLLER_TOOLS_VERSION ?= v0.9.0
GO_LICENSER_VERSION ?= v0.4.0
GEN_CRD_API_REFERENCE_DOCS_VERSION ?= 3f29e6853552dcf08a8e846b1225f275ed0f3e3b
GOLANGCI_LINT_VERSION ?= v1.46.2

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(KUSTOMIZE) || curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: operator-sdk
operator-sdk: $(OPERATOR_SDK)
$(OPERATOR_SDK): $(LOCALBIN)
	test -s $(OPERATOR_SDK) || curl -sLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/v${OPERATOR_SDK_VERSION}/operator-sdk_`go env GOOS`_`go env GOARCH`
	@chmod +x $(OPERATOR_SDK)

.PHONY: golangci-lint
golangci-lint: $(GOLANGCI_LINT)
$(GOLANGCI_LINT): $(GOLANGCI_LINT)
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: go-licenser
go-licenser: $(GO_LICENSER)
$(GO_LICENSER): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/elastic/go-licenser@$(GO_LICENSER_VERSION)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: gen-crd-api-reference-docs
gen-crd-api-reference-docs: $(GEN_CRD_API_REFERENCE_DOCS) ## Download gen-crd-api-reference-docs locally if necessary.
$(GEN_CRD_API_REFERENCE_DOCS): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/ahmetb/gen-crd-api-reference-docs@$(GEN_CRD_API_REFERENCE_DOCS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest
