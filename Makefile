
# Image URL to use all building/pushing image targets
IMG ?= goharbor/harbor-operator:dev

SHELL = /bin/sh

.PHONY: all
all: manager

# Run tests
.PHONY: test
test: generate
	go test ./... \
		-coverprofile cover.out

# Build manager binary
.PHONY: manager
manager: generate fmt vet
	go build \
		-o bin/manager \
		main.go

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: generate fmt vet $(TMPDIR)k8s-webhook-server/serving-certs
	# See https://github.com/settings/tokens for GITHUB_TOKEN. No permissions required.
	set -u ; \
		CONFIGURATION_FROM='file:./config-dev.yml' \
		REGISTRY_TEMPLATE_PATH=./config/manager/assets/registry-config.yaml.tmpl \
		REGISTRYCTL_TEMPLATE_PATH=./config/manager/assets/registryctl-config.yaml.tmpl \
		JOBSERVICE_TEMPLATE_PATH=./config/manager/assets/jobservice-config.yaml.tmpl \
		CORE_TEMPLATE_PATH=./config/manager/assets/core-config.conf.tmpl \
		CHARTMUSEUM_TEMPLATE_PATH=./config/manager/assets/chartmuseum-config.yaml.tmpl \
		NOTARY_SERVER_TEMPLATE_PATH=./config/manager/assets/notary-server-config.json.tmpl \
		NOTARY_SERVER_MIGRATION_SOURCE="github://holyhope:$${GITHUB_TOKEN}@theupdateframework/notary/migrations/server/postgresql#v0.6.1" \
		NOTARY_SIGNER_TEMPLATE_PATH=./config/manager/assets/notary-signer-config.json.tmpl \
		NOTARY_SIGNER_MIGRATION_SOURCE="github://holyhope:$${GITHUB_TOKEN}@theupdateframework/notary/migrations/signer/postgresql#v0.6.1" \
	go run *.go

# Run linters against all files
.PHONY: lint
lint: \
	go-lint \
	md-lint

# Install all dev dependencies
.PHONY: dev-tools
dev-tools: \
	controller-gen \
	golangci-lint \
	gomplate \
	helm \
	kubebuilder \
	kustomize \
	markdownlint \
	stringer

.PHONY: release
release: goreleaser
	# export GITHUB_TOKEN=...
	$(GORELEASER) release --rm-dist

#####################
#     Packaging     #
#####################

.PHONY: generate
generate: controller-gen stringer
	go generate ./...
	$(MAKE) vendor

vendor: go.mod go.sum
	go mod vendor

ASSETS := $(wildcard assets/*)

# Build the docker image
.PHONY: docker-build
docker-build:
	docker build . -t "$(IMG)"

# Push the docker image
.PHONY: docker-push
docker-push:
	docker push "$(IMG)"

#####################
#      Linters      #
#####################

# Run go linters
.PHONY: go-lint
go-lint: golangci-lint fmt vet generate
	$(GOLANGCI_LINT) run --verbose

# Run go fmt against code
.PHONY: fmt
fmt:
	go fmt ./...

# Run go vet against code
.PHONY: vet
vet: generate
	go vet ./...

# Check markdown files syntax
.PHONY: md-lint
md-lint: markdownlint
	$(MARKDOWNLINT) \
		-c "$(CURDIR)/.markdownlint.json" \
		--ignore "$(CURDIR)/vendor" \
		"$(CURDIR)"

#####################
#    Dev helpers    #
#####################

# Install CRDs into a cluster
.PHONY: install
install: generate kustomize
	$(KUSTOMIZE) build config/crd \
		| kubectl apply --validate=false -f -

# Uninstall CRDs from a cluster
.PHONY: uninstall
uninstall: generate kustomize
	$(KUSTOMIZE) build config/crd \
		| kubectl delete -f -

# Deploy controller in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy: generate kustomize
	cd config/manager && $(KUSTOMIZE) edit set image controller="$(IMG)"
	$(KUSTOMIZE) build config/default \
		| kubectl apply --validate=false -f -

# Deploy RBAC in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy
deploy-rbac: generate kustomize
	$(KUSTOMIZE) build config/rbac \
		| kubectl apply --validate=false -f -

.PHONY: sample
sample: kustomize
	$(KUSTOMIZE) build config/samples \
		| kubectl apply -f -
	kubectl get goharbor

.PHONY: sample
sample-%: kustomize
	$(KUSTOMIZE) build 'config/samples/$*' \
		| kubectl apply -f -
	kubectl get goharbor

.PHONY: install-dependencies
install-dependencies: certmanager redis postgresql ingress

redis: helm
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami
	$(HELM) upgrade --install harbor-redis bitnami/redis \
		--set usePassword=true

postgresql: helm
	$(MAKE) sample-database
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami
	$(HELM) upgrade --install harbor-database bitnami/postgresql \
		--set-string initdbScriptsConfigMap=harbor-init-db

INGRESS_NAMESPACE := nginx-ingress

ingress: helm
	kubectl get namespace $(INGRESS_NAMESPACE) || kubectl create namespace $(INGRESS_NAMESPACE)
	$(HELM) upgrade --install nginx stable/nginx-ingress \
		--namespace $(INGRESS_NAMESPACE) \
		--set-string controller.config.proxy-body-size=0

CERTMANAGER_NAMESPACE := cert-manager

certmanager: helm
	$(HELM) repo add jetstack https://charts.jetstack.io
	kubectl get namespace $(CERTMANAGER_NAMESPACE) || kubectl create namespace $(CERTMANAGER_NAMESPACE)
	$(HELM) upgrade --install certmanager jetstack/cert-manager \
		--namespace $(CERTMANAGER_NAMESPACE) \
		--version v0.15.1 \
		--set installCRDs=true


# Install local certificate
# Required for webhook server to start
.PHONY: dev-certificate
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
.PHONY: controller-gen
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
.PHONY: markdownlint
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
.PHONY: golangci-lint
golangci-lint:
ifeq (, $(shell which golangci-lint))
	# https://github.com/golangci/golangci-lint#install
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.22.2
GOLANGCI_LINT=$(GOBIN)/golangci-lint
else
GOLANGCI_LINT=$(shell which golangci-lint)
endif

# find or download kubebuilder
# download kubebuilder if necessary
.PHONY: kubebuilder
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
.PHONY: kustomize
kustomize:
ifeq (, $(shell which kustomize))
	# https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md
	curl -s https://raw.githubusercontent.com/kubernetes-sigs/kustomize/7eca29daeee6b583f5394a45d8edfd41c15dbe6d/hack/install_kustomize.sh | bash
	mv ./kustomize $(GOBIN)
	chmod u+x $(GOBIN)/kustomize
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
endif

# find or download gomplate
# download gomplate if necessary
.PHONY: gomplate
gomplate:
ifeq (, $(shell which gomplate))
	# https://docs.gomplate.ca/installing/#install-with-npm
	npm install --global gomplate@3.6.0 --no-save
GOMPLATE=$(NPMBIN)/gomplate
else
GOMPLATE=$(shell which gomplate)
endif

# find helm or raise an error
.PHONY: helm
helm:
ifeq (, $(shell which helm))
	echo "Helm not found. Please install it: https://helm.sh/docs/intro/install/#from-script" >&2 \
		&& false
HELM=helm-not-found
else
HELM=$(shell which helm)
endif

# find or download goreleaser
.PHONY: goreleaser
goreleaser:
ifeq (, $(shell which goreleaser))
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh \
		| sh -s v0.129.0
	mv ./bin/goreleaser $(GOBIN)
GORELEASER=$(GOBIN)/goreleaser
else
GORELEASER=$(shell which goreleaser)
endif

# find or download stringer
# download stringer if necessary
.PHONY: stringer
stringer:
ifeq (, $(shell which stringer))
	# https://pkg.go.dev/golang.org/x/tools/cmd/stringer
	go get golang.org/x/tools/cmd/stringer@v0.0.0-20200626171337-aa94e735be7f
STRINGER=$(GOBIN)/stringer
else
STRINGER=$(shell which stringer)
endif
:
