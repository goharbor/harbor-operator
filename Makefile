
# Image URL to use all building/pushing image targets
IMG ?= goharbor/harbor-operator:dev

CONFIGURATION_FROM ?= file:$(CURDIR)/config-dev.yml
export CONFIGURATION_FROM

CHARTS_DIRECTORY      := $(CURDIR)/charts
CHART_HARBOR_OPERATOR := $(CHARTS_DIRECTORY)/harbor-operator

REGISTRY_TEMPLATE_PATH     ?= $(CHART_HARBOR_OPERATOR)/assets/registry-config.yaml.tmpl
PORTAL_TEMPLATE_PATH       ?= $(CHART_HARBOR_OPERATOR)/assets/portal-config.conf.tmpl
REGISTRYCTL_TEMPLATE_PATH  ?= $(CHART_HARBOR_OPERATOR)/assets/registryctl-config.yaml.tmpl
JOBSERVICE_TEMPLATE_PATH   ?= $(CHART_HARBOR_OPERATOR)/assets/jobservice-config.yaml.tmpl
CORE_TEMPLATE_PATH         ?= $(CHART_HARBOR_OPERATOR)/assets/core-config.conf.tmpl
CHARTMUSEUM_TEMPLATE_PATH  ?= $(CHART_HARBOR_OPERATOR)/assets/chartmuseum-config.yaml.tmpl
NOTARYSERVER_TEMPLATE_PATH ?= $(CHART_HARBOR_OPERATOR)/assets/notaryserver-config.json.tmpl
NOTARYSIGNER_TEMPLATE_PATH ?= $(CHART_HARBOR_OPERATOR)/assets/notarysigner-config.json.tmpl

export REGISTRY_TEMPLATE_PATH
export PORTAL_TEMPLATE_PATH
export REGISTRYCTL_TEMPLATE_PATH
export JOBSERVICE_TEMPLATE_PATH
export CORE_TEMPLATE_PATH
export CHARTMUSEUM_TEMPLATE_PATH
export NOTARYSERVER_TEMPLATE_PATH
export NOTARYSIGNER_TEMPLATE_PATH

# See https://github.com/settings/tokens for GITHUB_TOKEN. No permissions required.
NOTARYSERVER_MIGRATION_SOURCE ?= github://holyhope:$${GITHUB_TOKEN}@theupdateframework/notary/migrations/server/postgresql\#v0.6.1
NOTARYSIGNER_MIGRATION_SOURCE ?= github://holyhope:$${GITHUB_TOKEN}@theupdateframework/notary/migrations/signer/postgresql\#v0.6.1

export NOTARYSERVER_MIGRATION_SOURCE
export NOTARYSIGNER_MIGRATION_SOURCE

########

define gosourcetemplate
{{- $$dir := .Dir }}
{{- range $$_, $$file := .GoFiles }}
	{{- if ne ( index $$file 0 | printf "%c" ) "/" }}
		{{- printf "%s/%s " $$dir $$file }}
	{{- end }}
{{- end -}}
endef

GO_SOURCES                  := $(sort $(subst $(CURDIR)/,,$(shell go list -f '$(gosourcetemplate)' ./...)))
GONOGENERATED_SOURCES       := $(sort $(shell grep -L 'DO NOT EDIT.' -- $(GO_SOURCES)))
GOWITHTESTS_SOURCES         := $(sort $(subst $(CURDIR)/,,$(shell go list -test -f '$(gosourcetemplate)' ./...)))
GO4CONTROLLERGEN_SOURCES    := $(sort $(shell grep -l '// +' -- $(GONOGENERATED_SOURCES)))

.SUFFIXES:            # Delete the default suffixes
.SUFFIXES: .go   # Define our suffix list

########

TMPDIR ?= /tmp/
export TMPDIR

.PHONY: all clean
all: manager

# Run tests
.PHONY: test
test: generate
	go test -vet=off ./... \
		-coverprofile cover.out

.PHONY: dependencies-test
dependencies-test: fmt
	go mod tidy
	$(MAKE) vendor
	git status
	test -z "$$(git diff-index --diff-filter=d --name-only HEAD)"


# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: generate vendor $(TMPDIR)k8s-webhook-server/serving-certs
	go run *.go

# Run linters against all files
.PHONY: lint
lint: \
	go-lint \
	helm-lint \
	docker-lint \
	make-lint \
	md-lint

# Install all dev dependencies
.PHONY: dev-tools
dev-tools: \
	controller-gen \
	golangci-lint \
	helm \
	kubebuilder \
	kustomize \
	markdownlint \
	stringer

#####################
#     Packaging     #
#####################

RELEASE_VERSION ?= 0.0.0-dev

# Build manager binary
.PHONY: manager
manager: generate vendor
	go build \
		-mod vendor \
		-o bin/manager \
		-ldflags "-X $$(go list -m).OperatorVersion=$(RELEASE_VERSION)" \
		*.go

.PHONY: package-chart
package-chart: chart
	$(HELM) package $(CHARTS_DIRECTORY)/harbor-operator \
		--app-version $(RELEASE_VERSION) \
		--destination $(CHARTS_DIRECTORY)

.PHONY: release
release: goreleaser
	# export GITHUB_TOKEN=...
	$(GORELEASER) release --rm-dist

.PHONY: manifests
manifests: controller-gen
	$(MAKE) config/rbac config/crd/bases config/webhook

config/webhook: $(GO4CONTROLLERGEN_SOURCES)
	$(CONTROLLER_GEN) webhook output:artifacts:config="$@" paths="./..."
	touch "$@"

config/rbac: $(GO4CONTROLLERGEN_SOURCES)
	$(CONTROLLER_GEN) rbac:roleName="manager-role" output:artifacts:config="$@" paths="./..."
	touch "$@"

config/crd/bases: $(GO4CONTROLLERGEN_SOURCES)
	$(CONTROLLER_GEN) crd:crdVersions="v1" output:artifacts:config="$@" paths="./..."
	touch "$@"

.PHONY: generate
generate: controller-gen stringer
	go generate ./...

vendor: go.mod go.sum
	go mod vendor

go.mod: $(GONOGENERATED_SOURCES)
	go mod tidy

go.sum: go.mod $(GONOGENERATED_SOURCES)
	go get ./...

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
go-lint: golangci-lint vet generate
	$(GOLANGCI_LINT) run --verbose

# Run go fmt against code
.PHONY: fmt
fmt: generate
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
		--ignore "$(CURDIR)/node_modules" \
		"$(CURDIR)"

docker-lint: hadolint
	$(HADOLINT) Dockerfile

make-lint: checkmake
	$(CHECKMAKE) Makefile

helm-lint: helm
	$(HELM) lint $(CHART_HARBOR_OPERATOR)

####################
#    Helm chart    #
####################

CHART_CRDS_PATH       := $(CHART_HARBOR_OPERATOR)/crds
CHART_TEMPLATE_PATH   := $(CHART_HARBOR_OPERATOR)/templates
CHART_TESTS_PATH      := $(CHART_TEMPLATE_PATH)/tests

CRD_GROUP := goharbor.io

DO_NOT_EDIT := Code generated by make. DO NOT EDIT.

$(CHART_CRDS_PATH)/$(CRD_GROUP)_%.yaml: kustomize config/crd/bases $(wildcard config/crd/*) $(wildcard config/helm/crds/*) $(wildcard config/crd/patches/*)
	mkdir -p $(CHART_CRDS_PATH)
	echo '# $(DO_NOT_EDIT)' > $(CHART_CRDS_PATH)/$*.yaml
	$(KUSTOMIZE) build config/helm/crds | \
	$(KUSTOMIZE) cfg grep --annotate=false 'metadata.name=$*\.$(subst .,\.,$(CRD_GROUP))' \
		>> $(CHART_CRDS_PATH)/$*.yaml

.PHONY: $(CHART_TESTS_PATH)/test.yaml
$(CHART_TESTS_PATH)/test.yaml: kustomize config/tests/postgresql/helm.yaml $(wildcard config/helm/tests/*) $(wildcard config/samples/harbor-full/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TESTS_PATH)/test.yaml
	$(KUSTOMIZE) build 'config/helm/tests' \
		>> $(CHART_TESTS_PATH)/test.yaml

.PHONY: config/tests/postgresql/helm.yaml
config/tests/postgresql/helm.yaml: helm
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami
	echo '# $(DO_NOT_EDIT)' > config/helm/tests/postgresql/helm.yaml
	$(HELM) template test bitnami/postgresql \
		--set-string initdbScriptsConfigMap=harbor-init-db \
		--set-string existingSecret='test-postgresql' \
	>> config/helm/tests/postgresql/helm.yaml

$(CHART_TEMPLATE_PATH)/role.yaml: kustomize config/rbac $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/role.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/role.yaml
	$(KUSTOMIZE) build config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Role' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=ClusterRole' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=RoleBinding' \
		>> $(CHART_TEMPLATE_PATH)/role.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/role.yaml

$(CHART_TEMPLATE_PATH)/clusterrole.yaml: kustomize config/rbac $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/clusterrole.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/clusterrole.yaml
	$(KUSTOMIZE) build config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=ClusterRole' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=ClusterRoleBinding' \
		>> $(CHART_TEMPLATE_PATH)/clusterrole.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/clusterrole.yaml

$(CHART_TEMPLATE_PATH)/rolebinding.yaml: kustomize config/rbac $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/rolebinding.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/rolebinding.yaml
	$(KUSTOMIZE) build config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=RoleBinding' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=ClusterRoleBinding' \
		>> $(CHART_TEMPLATE_PATH)/rolebinding.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/rolebinding.yaml

$(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml: kustomize config/rbac $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml
	$(KUSTOMIZE) build config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=ClusterRoleBinding' \
		>> $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml


$(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml: kustomize config/webhook $(wildcard config/helm/webhook/*) $(wildcard config/webhook/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml
	$(KUSTOMIZE) build config/helm/webhook | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=ValidatingWebhookConfiguration' | \
	sed "s/'\({{.*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml

$(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml: kustomize config/webhook $(wildcard config/helm/webhook/*) $(wildcard config/webhook/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml
	$(KUSTOMIZE) build config/helm/webhook | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=MutatingWebhookConfiguration' | \
	sed "s/'\({{.*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml

$(CHART_TEMPLATE_PATH)/certificate.yaml: kustomize config/certmanager $(wildcard config/helm/certmanager/*) $(wildcard config/certmanager/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/certificate.yaml
	$(KUSTOMIZE) build config/helm/certificate | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Certificate' \
		>> $(CHART_TEMPLATE_PATH)/certificate.yaml

$(CHART_TEMPLATE_PATH)/issuer.yaml: kustomize config/certmanager $(wildcard config/helm/certmanager/*) $(wildcard config/certmanager/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/issuer.yaml
	$(KUSTOMIZE) build config/helm/certificate | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Issuer' \
		>> $(CHART_TEMPLATE_PATH)/issuer.yaml

.PHONY:chart
chart: helm-generate

.PHONY:helm-generate
helm-generate: $(subst config/crd/bases/,$(CHART_CRDS_PATH)/,$(wildcard config/crd/bases/*.yaml)) \
	$(CHART_TEMPLATE_PATH)/role.yaml $(CHART_TEMPLATE_PATH)/clusterrole.yaml \
	$(CHART_TEMPLATE_PATH)/rolebinding.yaml $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml \
	$(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml $(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml \
	$(CHART_TEMPLATE_PATH)/certificate.yaml $(CHART_TEMPLATE_PATH)/issuer.yaml \
	$(CHART_TESTS_PATH)/test.yaml

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
.PHONY: deploy-rbac
deploy-rbac: generate kustomize
	$(KUSTOMIZE) build config/rbac \
		| kubectl apply --validate=false -f -

.PHONY: sample
sample: sample-harbor

.PHONY: sample
sample-%: kustomize
	$(KUSTOMIZE) build 'config/samples/$*' \
		| kubectl apply -f -
	kubectl get goharbor

.PHONY: install-dependencies
install-dependencies: certmanager redis postgresql ingress

.PHONY: redis
redis: helm
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami
	$(HELM) upgrade --install harbor-redis bitnami/redis \
		--set usePassword=true

.PHONY: postgresql
postgresql: helm
	$(MAKE) sample-database
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami
	$(HELM) upgrade --install harbor-database bitnami/postgresql \
		--set-string initdbScriptsConfigMap=harbor-init-db

INGRESS_NAMESPACE := nginx-ingress

.PHONY: ingress
ingress: helm
	kubectl get namespace $(INGRESS_NAMESPACE) || kubectl create namespace $(INGRESS_NAMESPACE)
	$(HELM) upgrade --install nginx stable/nginx-ingress \
		--namespace $(INGRESS_NAMESPACE) \
		--set-string controller.config.proxy-body-size=0

CERTMANAGER_NAMESPACE := cert-manager

.PHONY: certmanager
certmanager: helm jetstack
	kubectl get namespace $(CERTMANAGER_NAMESPACE) || kubectl create namespace $(CERTMANAGER_NAMESPACE)
	$(HELM) upgrade --install certmanager jetstack/cert-manager \
		--namespace $(CERTMANAGER_NAMESPACE) \
		--version v0.15.1 \
		--set installCRDs=true

.PHONY: jetstack
jetstack:
	$(HELM) repo add jetstack https://charts.jetstack.io

# Install local certificate
# Required for webhook server to start
.PHONY: dev-certificate
dev-certificate:
	$(RM) -r "$(TMPDIR)k8s-webhook-server/serving-certs"
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

$(GOBIN):
	mkdir -p "$(GOBIN)"

# Get the npm install path
NPMOPTS=#--global

NPMBIN=$(shell npm $(NPMOPTS) bin)

$(NPMBIN):
	mkdir -p "$(NPMBIN)"

.PHONY: go-binary
go-binary: $(GOBIN)
	@{ \
		set -uex ; \
		export CONTROLLER_GEN_TMP_DIR="$$(mktemp -d)" ; \
		cd "$$CONTROLLER_GEN_TMP_DIR" ; \
		go mod init tmp ; \
		go get "$${GO_DEPENDENCY}" ; \
		rm -rf "$${CONTROLLER_GEN_TMP_DIR}" ; \
	}

# find or download controller-gen
# download controller-gen if necessary
.PHONY: controller-gen
controller-gen:
ifeq (, $(shell which controller-gen))
	GO_DEPENDENCY='sigs.k8s.io/controller-tools/cmd/controller-gen@v0.2.4' $(MAKE) go-binary
CONTROLLER_GEN=$(GOBIN)/controller-gen
else
CONTROLLER_GEN=$(shell which controller-gen)
endif

# find or download markdownlint
# download markdownlint if necessary
.PHONY: markdownlint
markdownlint:
ifeq (, $(shell which markdownlint))
	$(MAKE) $(NPMBIN)
	# https://github.com/igorshubovych/markdownlint-cli#installation
	npm install $(NPMOPTS) markdownlint-cli@0.16.0 --no-save
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
	GO_DEPENDENCY='github.com/golangci/golangci-lint/cmd/golangci-lint@v1.27.0' $(MAKE) go-binary
GOLANGCI_LINT=$(GOBIN)/golangci-lint
else
GOLANGCI_LINT=$(shell which golangci-lint)
endif

# find or download kubebuilder
# download kubebuilder if necessary
.PHONY: kubebuilder
kubebuilder:
ifeq (, $(shell which kubebuilder))
	$(MAKE) $(GOBIN)
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
	$(MAKE) $(GOBIN)
	# https://github.com/kubernetes-sigs/kustomize/blob/master/docs/INSTALL.md
	curl -s https://raw.githubusercontent.com/kubernetes-sigs/kustomize/7eca29daeee6b583f5394a45d8edfd41c15dbe6d/hack/install_kustomize.sh | bash
	mv ./kustomize $(GOBIN)
	chmod u+x $(GOBIN)/kustomize
KUSTOMIZE=$(GOBIN)/kustomize
else
KUSTOMIZE=$(shell which kustomize)
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
	$(MAKE) $(GOBIN)
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
	GO_DEPENDENCY='golang.org/x/tools/cmd/stringer@v0.0.0-20200626171337-aa94e735be7f' $(MAKE) go-binary
STRINGER=$(GOBIN)/stringer
else
STRINGER=$(shell which stringer)
endif

# find or download hadolint
# download hadolint if necessary
.PHONY: hadolint
hadolint:
ifeq (, $(shell which hadolint))
	$(MAKE) $(GOBIN)
	# https://github.com/hadolint/hadolint/releases/
	curl -sL https://github.com/hadolint/hadolint/releases/download/v1.18.0/hadolint-$(shell uname -s)-x86_64 \
		> $(GOBIN)/hadolint
	chmod u+x $(GOBIN)/hadolint
HADOLINT=$(GOBIN)/hadolint
else
HADOLINT=$(shell which hadolint)
endif


# find or download checkmake
# download checkmake if necessary
.PHONY: checkmake
checkmake:
ifeq (, $(shell which checkmake))
	# https://github.com/mrtazz/checkmake#installation
	GO_DEPENDENCY='github.com/mrtazz/checkmake/cmd/checkmake@0.1.0' $(MAKE) go-binary
CHECKMAKE=$(GOBIN)/checkmake
else
CHECKMAKE=$(shell which checkmake)
endif
