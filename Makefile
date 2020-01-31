
# Image URL to use all building/pushing image targets
IMG ?= controller:latest

all: manager

# Run tests
test: generate manifests
	go test ./... -coverprofile cover.out

# Build manager binary
manager: generate fmt vet
	go build -o bin/manager main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
run: generate fmt vet manifests
	go run ./main.go

# Run linters against all files
lint: \
	go-lint \
	md-lint

# Install all dev dependencies
dev-tools: \
	controller-gen \
	golangci-lint \
	gomplate \
	helm \
	kubebuilder \
	kustomize \
	markdownlint \
	pkger

#####################
#     Packaging     #
#####################

# Produce CRDs that work back to Kubernetes 1.11 (no version conversion)
CRD_OPTIONS ?= "crd:trivialVersions=true"

# Generate manifests e.g. CRD, RBAC etc.
manifests: controller-gen
	$(CONTROLLER_GEN) \
		$(CRD_OPTIONS) \
		rbac:roleName="manager-role" \
		output:crd:artifacts:config="config/crd/bases" \
		webhook \
		paths="./..."

# Generate code
generate: controller-gen pkged.go
	$(CONTROLLER_GEN) \
		object:headerFile="./hack/boilerplate.go.txt" \
		paths="./..."

ASSETS := $(wildcard assets/*)

pkged.go: pkger $(ASSETS)
	$(PKGER) parse ; $(PKGER)

# Build the docker image
docker-build:
	docker build . -t "$(IMG)"

# Push the docker image
docker-push:
	docker push "$(IMG)"

#####################
#      Linters      #
#####################

# Run go linters
go-lint: golangci-lint fmt vet generate
	$(GOLANGCI_LINT) run --verbose

# Run go fmt against code
fmt:
	go fmt ./...

# Run go vet against code
vet: generate
	go vet ./...

# Check markdown files syntax
md-lint: markdownlint
	$(MARKDOWNLINT) \
		-c "$(CURDIR)/.markdownlint.json" \
		--ignore "$(CURDIR)/vendor" \
		"$(CURDIR)"

#####################
#    Dev helpers    #
#####################

# Install CRDs into a cluster
install: manifests
	$(KUSTOMIZE) build config/crd \
		| kubectl apply --validate=false -f -

# Uninstall CRDs from a cluster
uninstall: manifests
	$(KUSTOMIZE) build config/crd \
		| kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
deploy: manifests
	cd config/manager && $(KUSTOMIZE) edit set image controller="$(IMG)"
	$(KUSTOMIZE) build config/default \
		| kubectl apply --validate=false -f -

NAMESPACE_NAME=harboperator-system

sample: gomplate
	export \
		LBAAS_DOMAIN=$$(kubectl get svc nginx-nginx-ingress-controller -o jsonpath='{.status.loadBalancer.ingress[0].hostname}') \
		CORE_DATABASE_SECRET=$$(kubectl get secret core-database-postgresql -o jsonpath='{.data.postgresql-password}' | base64 --decode) \
		CLAIR_DATABASE_SECRET=$$(kubectl get secret clair-database-postgresql -o jsonpath='{.data.postgresql-password}' | base64 --decode) ; \
	kubectl kustomize config/samples \
		| gomplate \
		| kubectl apply -f -

install-dependencies: helm
	$(HELM) get notes core-database \
		|| $(HELM) install core-database stable/postgresql
	$(HELM) get notes clair-database \
		|| $(HELM) install clair-database stable/postgresql
	$(HELM) get notes jobservice-broker \
		|| $(HELM) install jobservice-broker stable/redis-ha
	$(HELM) get notes clair-adapter-broker \
		|| $(HELM) install clair-adapter-broker stable/redis-ha
	$(HELM) get notes registry-cache \
		|| $(HELM) install registry-cache stable/redis-ha
	$(HELM) get notes nginx \
		|| $(HELM) install nginx stable/nginx-ingress

# Install local certificate
# Required for webhook server to start
dev-certificate:
	rm -rf "$(TMPDIR)k8s-webhook-server/serving-certs"
	$(MAKE) "$(TMPDIR)k8s-webhook-server/serving-certs"

$(TMPDIR)k8s-webhook-server/serving-certs:
	mkdir -p "$(TMPDIR)k8s-webhook-server/serving-certs"
	openssl req \
		-new \
		-newkey rsa:4096 \
		-days 365 \
		-nodes \
		-x509 \
		-subj "/C=FR/O=Dev/OU=$(shell whoami)/CN=example.com" \
		-keyout "$(TMPDIR)k8s-webhook-server/serving-certs/tls.key" \
		-out "$(TMPDIR)k8s-webhook-server/serving-certs/tls.crt"

#####################
#     Dev Tools     #
#####################

# Get the currently used golang install path (in GOPATH/bin, unless GOBIN is set)
ifeq (,$(shell go env GOBIN))
GOBIN=$(shell go env GOPATH)/bin
else
GOBIN=$(shell go env GOBIN)
endif

t:
	echo $(GOBIN)

# Get the npm install path
NPMBIN=$$(npm --global bin)

SHELL := /bin/bash

# find or download controller-gen
# download controller-gen if necessary
controller-gen:
ifeq (, $(shell which controller-gen))
	@{ \
	set -e ;\
	CONTROLLER_GEN_TMP_DIR=$$(mktemp -d) ;\
	cd $$CONTROLLER_GEN_TMP_DIR ;\
	go mod init tmp ;\
	go get sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4 ;\
	rm -rf $$CONTROLLER_GEN_TMP_DIR ;\
	}
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# find or download markdownlint
# download markdownlint if necessary
markdownlint:
ifeq (, $(shell which markdownlint))
	# https://github.com/igorshubovych/markdownlint-cli#installation
	npm install --global markdownlint-cli@0.16.0 --no-save
MARKDOWNLINT=$(NPMBIN)/markdownlint
else
MARKDOWNLINT=$(shell which markdownlint)
endif

# find or download golangci-lint
# download golangci-lint if necessary
golangci-lint:
ifeq (, $(shell which golangci-lint))
	# https://github.com/golangci/golangci-lint#install
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.22.2
GOLANGCI_LINT=$(GOBIN)/golangci-lint
else
GOLANGCI_LINT=$(shell which golangci-lint)
endif

# find or download pkger
# download pkger if necessary
pkger:
ifeq (, $(shell which pkger))
	# https://github.com/markbates/pkger#installation
	go get github.com/markbates/pkger/cmd/pkger@v0.12.8
PKGER=$(GOBIN)/pkger
else
PKGER=$(shell which pkger)
endif

# find or download kubebuilder
# download kubebuilder if necessary
kubebuilder:
ifeq (, $(shell which kubebuilder))
	# https://kubebuilder.io/quick-start.html#installation
	curl -sSL "https://go.kubebuilder.io/dl/2.0.1/$(shell go env GOOS)/$(shell go env GOARCH)" \
		| tar -xz -C /tmp/
	mv /tmp/kubebuilder_2.0.1_$(shell go env GOOS)_$(shell go env GOARCH)/bin/* $(GOBIN)
KUBEBUILDER=$(GOBIN)/kubebuilder
else
KUBEBUILDER=$(shell which kubebuilder)
endif

# find or download kustomize
# download kustomize if necessary
kustomize:
ifeq (, $(shell which kustomize))
	# https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md
	curl -sSL https://api.github.com/repos/kubernetes-sigs/kustomize/releases/latest \
		| grep browser_download \
		| grep linux \
		| cut -d '"' -f 4 \
		| xargs curl -sSL \
		| tar -xz -C /tmp/
	mv /tmp/kustomize $(KUSTOMIZE)
	chmod u+x $(KUSTOMIZE)
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# find or download gomplate
# download gomplate if necessary
gomplate:
ifeq (, $(shell which gomplate))
	# https://docs.gomplate.ca/installing/#install-with-npm
	npm install --global gomplate@3.6.0 --no-save
GOMPLATE=$(NPMBIN)/gomplate
else
GOMPLATE=$(shell which gomplate)
endif

# find helm or raise an error
helm:
ifeq (, $(shell which helm))
	echo "Helm not found. Please install it: https://helm.sh/docs/intro/install/#from-script" >&2 \
		&& false
HELM=helm-not-found
else
HELM=$(shell which helm)
endif
