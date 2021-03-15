SHELL := /bin/bash

# Current Operator version
VERSION ?= v0.2.0
# Default bundle image tag
BUNDLE_IMG ?= controller-bundle:$(VERSION)
# Options for 'bundle-build'
ifneq ($(origin CHANNELS), undefined)
BUNDLE_CHANNELS := --channels=$(CHANNELS)
endif
ifneq ($(origin DEFAULT_CHANNEL), undefined)
BUNDLE_DEFAULT_CHANNEL := --default-channel=$(DEFAULT_CHANNEL)
endif
BUNDLE_METADATA_OPTS ?= $(BUNDLE_CHANNELS) $(BUNDLE_DEFAULT_CHANNEL)

# Image URL to use all building/pushing image targets
IMG ?= nchatsystem/consul-merge-controller:$(VERSION)
# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true,preserveUnknownFields=false"

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

all: manager

# Run tests
ENVTEST_ASSETS_DIR=$(shell pwd)/testbin
SCRIPTS_DIR=$(shell pwd)/scripts
test: generate fmt vet manifests
	mkdir -p ${ENVTEST_ASSETS_DIR}
	cp -f ${SCRIPTS_DIR}/controller-runtime/v0.7.0/setup-envtest.sh ${ENVTEST_ASSETS_DIR}/
	source ${ENVTEST_ASSETS_DIR}/setup-envtest.sh; fetch_envtest_tools $(ENVTEST_ASSETS_DIR); setup_envtest_env $(ENVTEST_ASSETS_DIR); \
		bash scripts/download-consul-bins.sh $(ENVTEST_ASSETS_DIR); \
		ENVTEST_ASSETS_DIR=$(ENVTEST_ASSETS_DIR) MAKEFILE_PATH=$(PWD)/Makefile go test ./controllers/... -coverprofile cover.out -ginkgo.trace

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

RUN_ARGS ?= ""
# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go $(RUN_ARGS)

# Install CRDs into a cluster
install: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl apply -f -

# Uninstall CRDs from a cluster
uninstall: manifests kustomize
	$(KUSTOMIZE) build config/crd | kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller=${IMG}
	$(KUSTOMIZE) build config/default | kubectl apply -f -

# UnDeploy controller from the configured Kubernetes cluster in ~/.kube/config
undeploy:
	$(KUSTOMIZE) build config/default | kubectl delete -f -

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) $(CRD_OPTIONS) rbac:roleName=manager-role webhook paths="./..." output:crd:artifacts:config=config/crd/bases

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet:
	go vet ./...

# Generate code
generate: controller-gen
	$(CONTROLLER_GEN) object:headerFile="hack/boilerplate.go.txt" paths="./..."

# Build the docker image
docker-build:
	docker build -t ${IMG} .

# Push the docker image
docker-push:
	docker push ${IMG}

# Download controller-gen locally if necessary
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
controller-gen:
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.4.1)

# Download kustomize locally if necessary
KUSTOMIZE = $(shell pwd)/bin/kustomize
kustomize:
	$(call go-get-tool,$(KUSTOMIZE),sigs.k8s.io/kustomize/kustomize/v3@v3.8.7)

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef

# Generate bundle manifests and metadata, then validate generated files.
.PHONY: bundle
bundle: manifests kustomize
	operator-sdk generate kustomize manifests -q
	cd config/manager && $(KUSTOMIZE) edit set image controller=$(IMG)
	$(KUSTOMIZE) build config/manifests | operator-sdk generate bundle -q --overwrite --version $(VERSION) $(BUNDLE_METADATA_OPTS)
	operator-sdk bundle validate ./bundle

# Build the bundle image.
.PHONY: bundle-build
bundle-build:
	docker build -f bundle.Dockerfile -t $(BUNDLE_IMG) .

# Temporary hack until consul-k8s updates its controller-runtime
GO_MOD_DEPS_DIR ?= vendor
CONSUL_K8S_VERSION ?= ""
CONSUL_K8S_DIR ?= $(GO_MOD_DEPS_DIR)/github.com/hashicorp/consul-k8s$(CONSUL_K8S_VERSION)
SED_FLAGS ?= $(shell echo $$OSTYPE | grep -q darwin && echo "-i ''" || echo "-i")
.PHONY: go-mod-vendor-hack
go-mod-vendor-hack:
	sed $(SED_FLAGS) -e 's-"k8s.io/api/admission/v1beta1"--' "$(CONSUL_K8S_DIR)/api/common/configentry_webhook.go"
	sed $(SED_FLAGS) -e 's-req.Operation == v1beta1.Create-req.Operation == "CREATE"-' "$(CONSUL_K8S_DIR)/api/common/configentry_webhook.go"
	sed $(SED_FLAGS) -e 's-"k8s.io/api/admission/v1beta1"--' "$(CONSUL_K8S_DIR)/api/v1alpha1/proxydefaults_webhook.go"
	sed $(SED_FLAGS) -e 's-req.Operation == v1beta1.Create-req.Operation == "CREATE"-' "$(CONSUL_K8S_DIR)/api/v1alpha1/proxydefaults_webhook.go"
	sed $(SED_FLAGS) -e 's-"k8s.io/api/admission/v1beta1"--' "$(CONSUL_K8S_DIR)/api/v1alpha1/serviceintentions_webhook.go"
	sed $(SED_FLAGS) -e 's-req.Operation == v1beta1.Create-req.Operation == "CREATE"-' "$(CONSUL_K8S_DIR)/api/v1alpha1/serviceintentions_webhook.go"
	sed $(SED_FLAGS) -e 's-req.Operation == v1beta1.Update-req.Operation == "UPDATE"-' "$(CONSUL_K8S_DIR)/api/v1alpha1/serviceintentions_webhook.go"

TEST_CONSUL_K8S_PATH ?= https://raw.githubusercontent.com/hashicorp/consul-k8s/v0.23.0/config/crd/bases/consul.hashicorp.com_
.PHONY: setup-local-consul-test-env
setup-local-consul-test-env:
	kubectl apply -f $(TEST_CONSUL_K8S_PATH)ingressgateways.yaml
	kubectl apply -f $(TEST_CONSUL_K8S_PATH)proxydefaults.yaml
	kubectl apply -f $(TEST_CONSUL_K8S_PATH)servicedefaults.yaml
	kubectl apply -f $(TEST_CONSUL_K8S_PATH)serviceintentions.yaml
	kubectl apply -f $(TEST_CONSUL_K8S_PATH)serviceresolvers.yaml
	kubectl apply -f $(TEST_CONSUL_K8S_PATH)servicerouters.yaml
	kubectl apply -f $(TEST_CONSUL_K8S_PATH)servicesplitters.yaml
	kubectl apply -f $(TEST_CONSUL_K8S_PATH)terminatinggateways.yaml

