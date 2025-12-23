# Image URL to use all building/pushing image targets
IMG ?= dbaas-operator:latest

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

# Setting SHELL to bash allows bash commands to be executed by recipes.
SHELL = /usr/bin/env bash -o pipefail
.SHELLFLAGS = -ec

.PHONY: all
all: build

##@ General

.PHONY: help
help: ## Display this help.
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n"} /^[a-zA-Z_0-9-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

##@ Development

.PHONY: manifests
manifests: ## Generate WebhookConfiguration, ClusterRole and CustomResourceDefinition objects.
	controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases

.PHONY: generate
generate: ## Generate code containing DeepCopy, DeepCopyInto, and DeepCopyObject method implementations.
	controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."

.PHONY: fmt
fmt: ## Run go fmt against code.
	go fmt ./...

.PHONY: vet
vet: ## Run go vet against code.
	go vet ./...

.PHONY: test
test: manifests generate fmt vet ## Run tests.
	go test ./... -coverprofile cover.out

##@ Build

.PHONY: build
build: manifests generate fmt vet ## Build manager binary.
	go build -o bin/manager main.go

.PHONY: run
run: manifests generate fmt vet ## Run a controller from your host.
	go run ./main.go

.PHONY: docker-build
docker-build: test ## Build docker image with the manager.
	docker build -t ${IMG} .

.PHONY: docker-push
docker-push: ## Push docker image with the manager.
	docker push ${IMG}

##@ Deployment

.PHONY: install
install: manifests ## Install CRDs into the K8s cluster specified in ~/.kube/config.
	kubectl apply -f config/crd/bases

.PHONY: uninstall
uninstall: manifests ## Uninstall CRDs from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/crd/bases

.PHONY: deploy
deploy: manifests ## Deploy controller to the K8s cluster specified in ~/.kube/config.
	kubectl create namespace dbaas-system --dry-run=client -o yaml | kubectl apply -f -
	kubectl apply -f config/rbac/role.yaml
	kubectl apply -f config/manager/manager.yaml

.PHONY: undeploy
undeploy: ## Undeploy controller from the K8s cluster specified in ~/.kube/config.
	kubectl delete -f config/manager/manager.yaml
	kubectl delete -f config/rbac/role.yaml

.PHONY: deploy-cnpg
deploy-cnpg: ## Install CloudNativePG operator
	kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.0.yaml
	@echo "Waiting for CNPG to be ready..."
	kubectl wait --for=condition=Available --timeout=300s deployment/cnpg-controller-manager -n cnpg-system

.PHONY: deploy-all
deploy-all: deploy-cnpg install deploy ## Install CNPG, CRDs and deploy operator

.PHONY: deploy-samples
deploy-samples: ## Deploy sample resources
	kubectl apply -f config/samples/databaseengine-cnpg.yaml
	kubectl apply -f config/samples/postgresql-cluster.yaml

.PHONY: quickstart
quickstart: ## Quick start - install everything and create test cluster
	@bash quickstart.sh

.PHONY: status
status: ## Show status of deployed resources
	@echo "=== Operator Status ==="
	@kubectl get deployment -n dbaas-system 2>/dev/null || echo "Operator not deployed"
	@echo ""
	@echo "=== CNPG Operator Status ==="
	@kubectl get deployment -n cnpg-system 2>/dev/null || echo "CNPG not deployed"
	@echo ""
	@echo "=== DatabaseClusters ==="
	@kubectl get databasecluster 2>/dev/null || echo "No DatabaseClusters found"
	@echo ""
	@echo "=== CNPG Clusters ==="
	@kubectl get cluster 2>/dev/null || echo "No CNPG Clusters found"
	@echo ""
	@echo "=== OpsRequests ==="
	@kubectl get opsrequest 2>/dev/null || echo "No OpsRequests found"

.PHONY: logs
logs: ## Show operator logs
	kubectl logs -n dbaas-system deployment/dbaas-operator-controller-manager -f

.PHONY: clean
clean: ## Clean up everything
	kubectl delete databasecluster --all --ignore-not-found=true
	kubectl delete opsrequest --all --ignore-not-found=true
	kubectl delete databaseengine --all --ignore-not-found=true
	$(MAKE) undeploy
	$(MAKE) uninstall

CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
.PHONY: controller-gen
controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.14.0)

# go-get-tool will 'go install' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(lastword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go install $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
