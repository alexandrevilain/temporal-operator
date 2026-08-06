# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

UNAME=$(shell uname -s)

# GNU vs BSD sed incompatibility adaption
ifeq ($(UNAME), Darwin)
  SEDI = sed -i ''
else
  SEDI = sed -i
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

YAML_PREFIX=spec.versions[0].schema.openAPIV3Schema.properties.spec.properties
CRD_PRESERVE=x-kubernetes-preserve-unknown-fields = true

.PHONY: manifests
manifests: controller-gen yq ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	$(CONTROLLER_GEN) rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.frontend.properties.initContainers.items.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.frontend.properties.initContainers.items.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.frontend.properties.initContainers.items.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.history.properties.initContainers.items.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.history.properties.initContainers.items.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.history.properties.initContainers.items.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.matching.properties.initContainers.items.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.matching.properties.initContainers.items.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.matching.properties.initContainers.items.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.internalFrontend.properties.initContainers.items.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.internalFrontend.properties.initContainers.items.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.internalFrontend.properties.initContainers.items.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.worker.properties.initContainers.items.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).services.properties.worker.properties.initContainers.items.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).services.properties.worker.properties.initContainers.items.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).jobInitContainers.items.properties)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i 'del(.$(YAML_PREFIX).jobInitContainers.items.required)' ./config/crd/bases/temporal.io_temporalclusters.yaml
	$(YQ) -i '.$(YAML_PREFIX).jobInitContainers.items.$(CRD_PRESERVE)' ./config/crd/bases/temporal.io_temporalclusters.yaml

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
	OPERATOR_IMAGE_PATH=/tmp/temporal-operator.tar go test ./tests/e2e -v -timeout 60m -args "-v=4"

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

.PHONY: helm
helm: helm-docs manifests artifacts
	$(SEDI) 's/^appVersion: ".*"/appVersion: "v$(shell cat VERSION)"/' charts/temporal-operator/Chart.yaml
	cp ${RELEASE_PATH}/temporal-operator.crds.yaml charts/temporal-operator/crds
	$(HELM_DOCS) --chart-search-root=charts/temporal-operator --template-files=hack/helm/template/README.md.gotmpl

.PHONY: bundle
bundle: manifests kustomize operator-sdk ## Generate bundle manifests and metadata, then validate generated files.
	$(OPERATOR_SDK) generate kustomize manifests -q
	$(KUSTOMIZE) build config/manifests | $(OPERATOR_SDK) generate bundle -q --overwrite --manifests --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	$(OPERATOR_SDK) bundle validate ./bundle

.PHONY: prepare-release
prepare-release: kustomize
	$(eval OLD_VERSION := $(shell curl -s https://api.github.com/repos/alexandrevilain/temporal-operator/releases | jq -r '.[0].tag_name'))
	cd config/manager && $(KUSTOMIZE) edit set image ghcr.io/alexandrevilain/temporal-operator:v$(VERSION)
	sed -i'' -e 's/replaces: temporal-operator.v.*/replaces: temporal-operator.$(OLD_VERSION)/' config/manifests/bases/temporal-operator.clusterserviceversion.yaml
	$(MAKE) bundle

OPERATOR_HUB_FORK_REPOSITORY ?= git@github.com:alexandrevilain/community-operators.git
.PHONY: operatorhub
operatorhub:
	VERSION=$(VERSION) FORK_REPOSITORY=$(OPERATOR_HUB_FORK_REPOSITORY) ./hack/operatorhub.sh

##@ Build Dependencies

RELEASE_PATH ?= $(shell pwd)/out/release/artifacts

## Location to install dependencies to
LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
	mkdir -p $(LOCALBIN)

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
OPERATOR_SDK ?= $(LOCALBIN)/operator-sdk
CONTROLLER_GEN ?= GOFLAGS=-mod=mod $(LOCALBIN)/controller-gen
ENVTEST ?= $(LOCALBIN)/setup-envtest
GO_LICENSER ?= $(LOCALBIN)/go-licenser
GEN_CRD_API_REFERENCE_DOCS ?= $(LOCALBIN)/gen-crd-api-reference-docs
GOLANGCI_LINT ?= $(LOCALBIN)/golangci-lint
YQ ?= $(LOCALBIN)/yq
KIND_WITH_REGISTRY ?= $(LOCALBIN)/kind-with-registry
HELM_DOCS ?= $(LOCALBIN)/helm-docs

## Tool Versions
KUSTOMIZE_VERSION ?= v5.8.1
OPERATOR_SDK_VERSION ?= 1.42.3
CONTROLLER_TOOLS_VERSION ?= v0.21.0
GO_LICENSER_VERSION ?= v0.4.2
GEN_CRD_API_REFERENCE_DOCS_VERSION ?= fca9c57bb1b2075a32100ca24637513a7b1d9dfe
GOLANGCI_LINT_VERSION ?= v2.12.2
YQ_VERSION ?= v4.53.3
KIND_WITH_REGISTRY_VERSION ?= 0.32.0
HELM_DOCS_VERSION ?= v1.14.2
# GO_VERSION is the Go version this module targets, taken from go.mod. It is used to
# pin GOTOOLCHAIN when building golangci-lint: golangci-lint's own go.mod requires a
# lower Go version than ours, so a plain `go install` builds it with that lower
# toolchain, and it then refuses to run with "the Go language version used to build
# golangci-lint is lower than the targeted Go version".
GO_VERSION ?= $(shell go list -m -f '{{.GoVersion}}')
#ENVTEST_VERSION is the version of controller-runtime release branch to fetch the envtest setup script (i.e. release-0.20)
ENVTEST_VERSION ?= $(shell v='$(call gomodver,sigs.k8s.io/controller-runtime)'; \
  [ -n "$$v" ] || { echo "Set ENVTEST_VERSION manually (controller-runtime replace has no tag)" >&2; exit 1; }; \
  printf '%s\n' "$$v" | sed -E 's/^v?([0-9]+)\.([0-9]+).*/release-\1.\2/')
#ENVTEST_K8S_VERSION is the version of Kubernetes to use for setting up ENVTEST binaries (i.e. 1.31)
ENVTEST_K8S_VERSION ?= $(shell v='$(call gomodver,k8s.io/api)'; \
  [ -n "$$v" ] || { echo "Set ENVTEST_K8S_VERSION manually (k8s.io/api replace has no tag)" >&2; exit 1; }; \
  printf '%s\n' "$$v" | sed -E 's/^v?[0-9]+\.([0-9]+).*/1.\1/')

# The version-checked tools below hang their recipe off the .PHONY target rather
# than off $(LOCALBIN)/<tool>. As a file target the recipe is skipped whenever the
# binary already exists, so the version check never runs and a bump to the pins
# above silently has no effect on anyone with a warm bin/ -- which is how a v4
# kustomize survived the v5 bump and emitted unsubstituted placeholders.
.PHONY: kustomize
kustomize: $(LOCALBIN) ## Download kustomize locally if necessary. If wrong version is installed, it will be removed before downloading.
	@if test -x $(LOCALBIN)/kustomize && ! $(LOCALBIN)/kustomize version | grep -q $(KUSTOMIZE_VERSION); then \
		echo "$(LOCALBIN)/kustomize version is not expected $(KUSTOMIZE_VERSION). Removing it before installing."; \
		rm -rf $(LOCALBIN)/kustomize; \
	fi
	@test -s $(LOCALBIN)/kustomize || GOBIN=$(LOCALBIN) go install sigs.k8s.io/kustomize/kustomize/v5@${KUSTOMIZE_VERSION}

.PHONY: operator-sdk
operator-sdk: $(LOCALBIN)
	@test -s $(LOCALBIN)/operator-sdk && $(LOCALBIN)/operator-sdk version | grep -q $(OPERATOR_SDK_VERSION) || \
	{ curl -sLo $(OPERATOR_SDK) https://github.com/operator-framework/operator-sdk/releases/download/v${OPERATOR_SDK_VERSION}/operator-sdk_`go env GOOS`_`go env GOARCH` && chmod +x $(OPERATOR_SDK); }

.PHONY: golangci-lint
golangci-lint: $(LOCALBIN)
	@test -s $(LOCALBIN)/golangci-lint && $(LOCALBIN)/golangci-lint version | grep -q $(GOLANGCI_LINT_VERSION) || \
	GOTOOLCHAIN=go$(GO_VERSION) GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@$(GOLANGCI_LINT_VERSION)

.PHONY: go-licenser
go-licenser: $(GO_LICENSER)
$(GO_LICENSER): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/elastic/go-licenser@$(GO_LICENSER_VERSION)

.PHONY: controller-gen
controller-gen: $(LOCALBIN) ## Download controller-gen locally if necessary. If wrong version is installed, it will be overwritten.
	@test -s $(LOCALBIN)/controller-gen && $(LOCALBIN)/controller-gen --version | grep -q $(CONTROLLER_TOOLS_VERSION) || \
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: gen-crd-api-reference-docs
gen-crd-api-reference-docs: $(GEN_CRD_API_REFERENCE_DOCS) ## Download gen-crd-api-reference-docs locally if necessary.
$(GEN_CRD_API_REFERENCE_DOCS): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/ahmetb/gen-crd-api-reference-docs@$(GEN_CRD_API_REFERENCE_DOCS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	test -s $(LOCALBIN)/setup-envtest || GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@$(ENVTEST_VERSION)

# yq is version-checked because `make manifests` and `make helm` run it over the
# generated CRDs; a stale yq can change emitted YAML formatting.
.PHONY: yq
yq: $(LOCALBIN)
	@test -s $(LOCALBIN)/yq && $(LOCALBIN)/yq --version | grep -q $(YQ_VERSION) || \
	GOBIN=$(LOCALBIN) go install github.com/mikefarah/yq/v4@$(YQ_VERSION)

.PHONY: kind-with-registry
kind-with-registry: $(KIND_WITH_REGISTRY)
$(KIND_WITH_REGISTRY): $(LOCALBIN)
	curl -sLo $(KIND_WITH_REGISTRY) https://raw.githubusercontent.com/kubernetes-sigs/kind/v$(KIND_WITH_REGISTRY_VERSION)/site/static/examples/kind-with-registry.sh
	chmod +x $(KIND_WITH_REGISTRY)

.PHONY: helm-docs
helm-docs: $(HELM_DOCS) ## Generate helm documentation
$(HELM_DOCS): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/norwoodj/helm-docs/cmd/helm-docs@$(HELM_DOCS_VERSION)


define gomodver
$(shell go list -m -f '{{if .Replace}}{{.Replace.Version}}{{else}}{{.Version}}{{end}}' $(1) 2>/dev/null)
endef