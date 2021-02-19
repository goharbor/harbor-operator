
# Image URL to use all building/pushing image targets
IMG ?= goharbor/harbor-operator:dev
RELEASE_VERSION ?= 0.0.0-dev

CONFIGURATION_FROM ?= env,file:$(CURDIR)/config-dev.yml
export CONFIGURATION_FROM

REGISTRY_TEMPLATE_PATH     ?= $(CURDIR)/config/config/assets/registry-config.yaml.tmpl
PORTAL_TEMPLATE_PATH       ?= $(CURDIR)/config/config/assets/portal-config.conf.tmpl
REGISTRYCTL_TEMPLATE_PATH  ?= $(CURDIR)/config/config/assets/registryctl-config.yaml.tmpl
JOBSERVICE_TEMPLATE_PATH   ?= $(CURDIR)/config/config/assets/jobservice-config.yaml.tmpl
CORE_TEMPLATE_PATH         ?= $(CURDIR)/config/config/assets/core-config.conf.tmpl
CHARTMUSEUM_TEMPLATE_PATH  ?= $(CURDIR)/config/config/assets/chartmuseum-config.yaml.tmpl
NOTARYSERVER_TEMPLATE_PATH ?= $(CURDIR)/config/config/assets/notaryserver-config.json.tmpl
NOTARYSIGNER_TEMPLATE_PATH ?= $(CURDIR)/config/config/assets/notarysigner-config.json.tmpl

export REGISTRY_TEMPLATE_PATH
export PORTAL_TEMPLATE_PATH
export REGISTRYCTL_TEMPLATE_PATH
export JOBSERVICE_TEMPLATE_PATH
export CORE_TEMPLATE_PATH
export CHARTMUSEUM_TEMPLATE_PATH
export NOTARYSERVER_TEMPLATE_PATH
export NOTARYSIGNER_TEMPLATE_PATH

ifeq (,$(shell which kubens 2> /dev/null))
NAMESPACE ?= $$(kubectl config get-contexts "$$(kubectl config current-context)" --no-headers | awk -F " " '{ if ($$5=="") print "default" ; else print $$5; }')
else
NAMESPACE ?= $$(kubens -c)
endif

CHARTS_DIRECTORY      := charts
CHART_HARBOR_OPERATOR := $(CHARTS_DIRECTORY)/harbor-operator

########

define gosourcetemplate
{{- $$dir := .Dir }}
{{- range $$_, $$file := .GoFiles }}
	{{- if ne ( index $$file 0 | printf "%c" ) "/" }}
		{{- printf "%s/%s " $$dir $$file }}
	{{- end }}
{{- end -}}
endef

GO_SOURCES                  := $(sort $(subst $(CURDIR)/,,$(shell go list -mod=readonly -f '$(gosourcetemplate)' ./... 2> /dev/null)))
GONOGENERATED_SOURCES       := $(sort $(shell grep -L 'DO NOT EDIT.' -- $(GO_SOURCES)))
GOWITHTESTS_SOURCES         := $(sort $(subst $(CURDIR)/,,$(shell go list -mod=readonly -test -f '$(gosourcetemplate)' ./... 2> /dev/null)))
GO4CONTROLLERGEN_SOURCES    := $(sort $(shell grep -l '// +' -- $(GONOGENERATED_SOURCES)))

.SUFFIXES:       # Delete the default suffixes
.SUFFIXES: .go   # Define our suffix list

########

TMPDIR ?= /tmp/
export TMPDIR

.PHONY: all clean
all: manager

# Run tests
.PHONY:test
test: go-test go-dependencies-test

# Run against the configured Kubernetes cluster in ~/.kube/config
.PHONY: run
run: go-generate certmanager $(TMPDIR)k8s-webhook-server/serving-certs/tls.crt
	go run *.go

# Run linters against all files
.PHONY: lint
lint: \
	go-lint \
	helm-lint \
	docker-lint \
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
#      Tests        #
#####################

.PHONY: go-dependencies-test
go-dependencies-test: fmt
	go mod tidy
	$(MAKE) diff

.PHONY: generated-diff-test
generated-diff-test: fmt generate
	$(MAKE) diff

.PHONY: diff
diff:
	git status
	git diff --stat --diff-filter=d --exit-code HEAD

GO_TEST_OPTS ?= -p 1 -vet=off

.PHONY: go-test
go-test: install
ifeq (, $(USE_EXISTING_CLUSTER))
	$(warning USE_EXISTING_CLUSTER variable is not defined)
endif
	go test \
		$(GO_TEST_OPTS) \
		./... \
		-coverprofile cover.out

.PHONY: release
release-test: goreleaser
	$(GORELEASER) release --rm-dist --snapshot

CHART_RELEASE_NAME ?= harbor-operator
CHART_HARBOR_CLASS ?=

helm-install: helm helm-generate
	$(MAKE) kube-namespace
	$(HELM) upgrade --namespace "$(NAMESPACE)" --install $(CHART_RELEASE_NAME) $(CHARTS_DIRECTORY)/harbor-operator-$(RELEASE_VERSION).tgz \
		--set-string image.repository="$$(echo $(IMG) | sed 's/:.*//')" \
		--set-string image.tag="$$(echo $(IMG) | sed 's/.*://')" \
		--set-string harborClass='$(CHART_HARBOR_CLASS)'

#####################
#     Packaging     #
#####################

# Build manager binary
.PHONY: manager
manager: go-generate
	go build \
		-o bin/manager \
		-ldflags "-X $$(go list -m).OperatorVersion=$(RELEASE_VERSION)" \
		*.go

.PHONY:helm-generate
helm-generate: $(CHARTS_DIRECTORY)/index.yaml

.PHONY: release
release: goreleaser
	# export GITHUB_TOKEN=...
	$(GORELEASER) release --rm-dist

.PHONY: manifests
manifests: config/rbac config/crd/bases config/webhook

config/webhook: controller-gen $(GO4CONTROLLERGEN_SOURCES)
	$(CONTROLLER_GEN) webhook output:artifacts:config="$@" paths="./..."
	touch "$@"

config/rbac: controller-gen $(GO4CONTROLLERGEN_SOURCES)
	$(CONTROLLER_GEN) rbac:roleName="manager-role" output:artifacts:config="$@" paths="./..."
	touch "$@"

config/crd/bases: controller-gen $(GO4CONTROLLERGEN_SOURCES)
	$(CONTROLLER_GEN) crd:crdVersions="v1" output:artifacts:config="$@" paths="./..."
	touch "$@"

.PHONY: generate
generate: go-generate helm-generate

go.mod: $(GONOGENERATED_SOURCES)
	go mod tidy

go.sum: go.mod $(GONOGENERATED_SOURCES)
	go get ./...

# Build the docker image
.PHONY: docker-build
docker-build: dist/harbor-operator_linux_amd64/manager
	docker build dist/harbor-operator_linux_amd64 \
		-f Dockerfile \
		-t "$(IMG)"

# Push the docker image
.PHONY: docker-push
docker-push:
	docker push "$(IMG)"

dist/harbor-operator_linux_amd64/manager:
	mkdir -p dist/harbor-operator_linux_amd64
	CGO_ENABLED=0 \
    GOOS="linux" \
    GOARCH="amd64" \
	go build \
		-o dist/harbor-operator_linux_amd64/manager \
		-ldflags "-X $$(go list -m).OperatorVersion=$(RELEASE_VERSION)" \
		*.go

#####################
#      Linters      #
#####################

# Run go linters
.PHONY: go-lint
go-lint: golangci-lint vet go-generate
	$(GOLANGCI_LINT) run --verbose

# Run go fmt against code
.PHONY: fmt
fmt: go-generate
	go fmt ./...

# Run go vet against code
.PHONY: vet
vet: go-generate
	go vet ./...

# Check markdown files syntax
.PHONY: md-lint
md-lint: markdownlint $(CHART_HARBOR_OPERATOR)/README.md
	$(MARKDOWNLINT) \
		-c "$(CURDIR)/.markdownlint.json" \
		--ignore "$(CURDIR)/node_modules" \
		"$(CURDIR)"

docker-lint: hadolint
	$(HADOLINT) Dockerfile

helm-lint: helm helm-generate
	$(HELM) lint $(CHART_HARBOR_OPERATOR)

####################
#    Helm chart    #
####################

CHART_REPO_URL := /harbor-operator/charts

DO_NOT_EDIT := Code generated by make. DO NOT EDIT.

$(CHARTS_DIRECTORY)/index.yaml: $(CHARTS_DIRECTORY)/harbor-operator-$(RELEASE_VERSION).tgz
	$(HELM) repo index \
		--url $(CHART_REPO_URL) \
		$(CHARTS_DIRECTORY)

CHART_TEMPLATE_PATH := $(CHART_HARBOR_OPERATOR)/templates

CRD_GROUP := goharbor.io

$(CHARTS_DIRECTORY)/harbor-operator-$(RELEASE_VERSION).tgz: $(CHART_HARBOR_OPERATOR)/README.md $(CHART_HARBOR_OPERATOR)/crds \
	$(CHART_HARBOR_OPERATOR)/assets $(wildcard $(CHART_HARBOR_OPERATOR)/assets/*) \
	$(CHART_HARBOR_OPERATOR)/charts $(CHART_HARBOR_OPERATOR)/Chart.lock \
	$(CHART_TEMPLATE_PATH)/role.yaml $(CHART_TEMPLATE_PATH)/clusterrole.yaml \
	$(CHART_TEMPLATE_PATH)/rolebinding.yaml $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml \
	$(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml $(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml \
	$(CHART_TEMPLATE_PATH)/certificate.yaml $(CHART_TEMPLATE_PATH)/issuer.yaml \
	$(CHART_TEMPLATE_PATH)/deployment.yaml
	$(HELM) package $(CHART_HARBOR_OPERATOR) \
		--version $(RELEASE_VERSION) \
		--app-version $(RELEASE_VERSION) \
		--destination $(CHARTS_DIRECTORY)

$(CHART_HARBOR_OPERATOR)/crds: config/crd/bases
	rm -f '$@'
	ln -vs ../../config/crd/bases '$@'

$(CHART_HARBOR_OPERATOR)/assets:
	rm -f '$@'
	ln -vs ../../config/config/assets '$@'

$(CHART_TEMPLATE_PATH)/deployment.yaml: kustomize $(wildcard config/helm/deployment/*) $(wildcard config/manager/*) $(wildcard config/config/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/deployment.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/deployment | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Deployment' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/deployment.yaml
	cat config/helm/deployment/foot.yaml >> $(CHART_TEMPLATE_PATH)/deployment.yaml

$(CHART_TEMPLATE_PATH)/role.yaml: kustomize $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/role.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/role.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Role' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=ClusterRole' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=RoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/role.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/role.yaml

$(CHART_TEMPLATE_PATH)/clusterrole.yaml: kustomize $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/clusterrole.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/clusterrole.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=ClusterRole' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=ClusterRoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/clusterrole.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/clusterrole.yaml

$(CHART_TEMPLATE_PATH)/rolebinding.yaml: kustomize $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/rolebinding.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/rolebinding.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=RoleBinding' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=ClusterRoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/rolebinding.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/rolebinding.yaml

$(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml: kustomize $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=ClusterRoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml

$(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml: kustomize $(wildcard config/helm/webhook/*) $(wildcard config/webhook/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/webhook | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=ValidatingWebhookConfiguration' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml

$(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml: kustomize $(wildcard config/helm/webhook/*) $(wildcard config/webhook/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/webhook | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=MutatingWebhookConfiguration' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml

$(CHART_TEMPLATE_PATH)/certificate.yaml: kustomize $(wildcard config/helm/certmanager/*) $(wildcard config/certmanager/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/certificate.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/certificate | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Certificate' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/certificate.yaml

$(CHART_TEMPLATE_PATH)/issuer.yaml: kustomize $(wildcard config/helm/certmanager/*) $(wildcard config/certmanager/*)
	echo '{{- /* $(DO_NOT_EDIT) */ -}}' > $(CHART_TEMPLATE_PATH)/issuer.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/certificate | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Issuer' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/issuer.yaml

$(CHART_HARBOR_OPERATOR)/charts: $(CHART_HARBOR_OPERATOR)/Chart.lock
	$(HELM) dependency build $(CHART_HARBOR_OPERATOR)

$(CHART_HARBOR_OPERATOR)/Chart.lock: $(CHART_HARBOR_OPERATOR)/Chart.yaml
	$(HELM) dependency update $(CHART_HARBOR_OPERATOR)

$(CHART_HARBOR_OPERATOR)/README.md: helm-docs $(CHART_HARBOR_OPERATOR)/README.md.gotmpl $(CHART_HARBOR_OPERATOR)/values.yaml $(CHART_HARBOR_OPERATOR)/Chart.yaml
	cd $(CHART_HARBOR_OPERATOR) ; $(HELM_DOCS)

#####################
#    Dev helpers    #
#####################

# Install CRDs into a cluster
.PHONY: install
install: go-generate
	kubectl apply -f config/crd/bases

# Uninstall CRDs from a cluster
.PHONY: uninstall
uninstall: go-generate
	kubectl delete -f config/crd/bases

go-generate: controller-gen stringer
	export PATH="$(BIN):$${PATH}" ; \
	go generate ./...

# Deploy RBAC in the configured Kubernetes cluster in ~/.kube/config
.PHONY: deploy-rbac
deploy-rbac: go-generate kustomize
	$(KUSTOMIZE) build --reorder legacy config/rbac \
		| kubectl apply --validate=false -f -

.PHONY: sample
sample: sample-harbor

.PHONY: sample-database
sample-database: kustomize
	$(KUSTOMIZE) build --reorder legacy 'config/samples/database' \
		| kubectl apply -f -

.PHONY: sample-redis
sample-redis: kustomize
	$(KUSTOMIZE) build 'config/samples/redis' \
		| kubectl apply -f -

.PHONY: sample-github-secret
sample-github-secret:
	! test -z $(GITHUB_TOKEN)
	kubectl create secret generic \
		github-credentials \
			--type=goharbor.io/github \
			--from-literal=github-token=$(GITHUB_TOKEN) \
			--dry-run=client -o yaml \
		| kubectl apply -f -

.PHONY: sample-%
sample-%: kustomize postgresql redis sample-github-secret
	$(KUSTOMIZE) build --reorder legacy 'config/samples/$*' \
		| kubectl apply -f -
	kubectl get goharbor

.PHONY: install-dependencies
install-dependencies: certmanager postgresql redis ingress

.PHONY: redis
redis: helm sample-redis
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami
	$(HELM) upgrade --install harbor-redis bitnami/redis \
		--set-string existingSecret=harbor-redis \
		--set-string existingSecretPasswordKey=redis-password \
		--set usePassword=true

.PHONY: postgresql
postgresql: helm sample-database
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami
	$(HELM) upgrade --install harbor-database bitnami/postgresql \
		--set-string initdbScriptsConfigMap=harbor-init-db \
		--set-string existingSecret=harbor-database-password

.PHONY: kube-namespace
kube-namespace:
	kubectl get namespace "$(NAMESPACE)" 2>&1 > /dev/null \
	|| kubectl create namespace "$(NAMESPACE)"

INGRESS_NAMESPACE := nginx-ingress

.PHONY: ingress
ingress: helm
	$(MAKE) kube-namespace NAMESPACE=$(INGRESS_NAMESPACE)
	$(HELM) repo add ingress-nginx https://kubernetes.github.io/ingress-nginx # https://github.com/kubernetes/ingress-nginx/tree/master/charts/ingress-nginx#get-repo-info
	$(HELM) upgrade --install nginx ingress-nginx/ingress-nginx \
		--namespace $(INGRESS_NAMESPACE) \
		--set-string controller.config.proxy-body-size=0

CERTMANAGER_NAMESPACE := cert-manager

.PHONY: certmanager
certmanager: helm jetstack
	$(MAKE) kube-namespace NAMESPACE=$(CERTMANAGER_NAMESPACE)
	$(HELM) repo add jetstack https://charts.jetstack.io # https://cert-manager.io/docs/installation/kubernetes/
	$(HELM) upgrade --install certmanager jetstack/cert-manager \
		--namespace $(CERTMANAGER_NAMESPACE) \
		--version v1.0.3 \
		--set installCRDs=true

.PHONY: jetstack
jetstack:
	$(HELM) repo add jetstack https://charts.jetstack.io

# Install local certificate
# Required for webhook server to start
.PHONY: dev-certificate
dev-certificate:
	$(RM) -r "$(TMPDIR)k8s-webhook-server/serving-certs"
	$(MAKE) $(TMPDIR)k8s-webhook-server/serving-certs/tls.crt

$(TMPDIR)k8s-webhook-server/serving-certs/tls.crt:
	mkdir -p "$(TMPDIR)k8s-webhook-server/serving-certs"
	openssl req \
		-new \
		-newkey rsa:4096 \
		-days 365 \
		-nodes \
		-x509 \
		-subj "/C=FR/O=Dev/OU=$$(whoami)/CN=example.com" \
		-keyout "$(TMPDIR)k8s-webhook-server/serving-certs/tls.key" \
		-out "$(TMPDIR)k8s-webhook-server/serving-certs/tls.crt"

#####################
#     Dev Tools     #
#####################

BIN ?= $(CURDIR)/bin

$(BIN):
	mkdir -p "$(BIN)"

.PHONY:clean
clean:
	rm -rf $(BIN) node_modules dist

# find or download controller-gen
# download controller-gen if necessary
CONTROLLER_GEN_VERSION := 0.4.1
CONTROLLER_GEN := $(BIN)/controller-gen

.PHONY: controller-gen
controller-gen:
	@$(CONTROLLER_GEN) --version 2>&1 \
		| grep 'v$(CONTROLLER_GEN_VERSION)' > /dev/null \
	|| rm -f $(CONTROLLER_GEN)
	@$(MAKE) $(CONTROLLER_GEN)

$(CONTROLLER_GEN):
	$(MAKE) $(BIN)
	# https://github.com/kubernetes-sigs/controller-tools/tree/master/cmd/controller-gen
	go get 'sigs.k8s.io/controller-tools/cmd/controller-gen@v$(CONTROLLER_GEN_VERSION)'
	go build -mod=readonly -o $(CONTROLLER_GEN) sigs.k8s.io/controller-tools/cmd/controller-gen
	go mod tidy

# find or download markdownlint
# download markdownlint if necessary
MARKDOWNLINT_VERSION := 0.16.0
MARKDOWNLINT := $(BIN)/markdownlint

.PHONY: markdownlint
markdownlint:
	@$(MARKDOWNLINT) version 2>&1 \
		| grep '$(MARKDOWNLINT_VERSION)' > /dev/null \
	|| rm -f $(MARKDOWNLINT)
	@$(MAKE) $(MARKDOWNLINT)

$(MARKDOWNLINT):
	$(MAKE) $(BIN)
	# https://github.com/igorshubovych/markdownlint-cli#installation
	npm install markdownlint-cli@$(MARKDOWNLINT_VERSION) --no-save
	ln -s "$$(npm bin)/markdownlint" $(MARKDOWNLINT)

# find or download golangci-lint
# download golangci-lint if necessary
GOLANGCI_LINT := $(BIN)/golangci-lint
GOLANGCI_LINT_VERSION := 1.33.0

.PHONY: golangci-lint
golangci-lint:
	@$(GOLANGCI_LINT) version --format short 2>&1 \
		| grep '$(GOLANGCI_LINT_VERSION)' > /dev/null \
	|| rm -f $(GOLANGCI_LINT)
	@$(MAKE) $(GOLANGCI_LINT)

$(GOLANGCI_LINT):
	$(MAKE) $(BIN)
	# https://golangci-lint.run/usage/install/#linux-and-windows
	curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh \
		| sh -s -- -b $(BIN) 'v$(GOLANGCI_LINT_VERSION)'

# find or download kubebuilder
# download kubebuilder if necessary
KUBEBUIDER_VERSION := 2.0.1
KUBEBUILDER=$(BIN)/kubebuilder

.PHONY: kubebuilder
kubebuilder:
	@$(KUBEBUILDER) version 2>&1 \
		| grep 'KubeBuilderVersion:"$(KUBEBUIDER_VERSION)"' > /dev/null \
	|| rm -f $(KUBEBUILDER)
	@$(MAKE) $(KUBEBUILDER)

$(KUBEBUILDER):
	$(MAKE) $(BIN)
	# https://kubebuilder.io/quick-start.html#installation
	curl -sSL "https://go.kubebuilder.io/dl/$(KUBEBUIDER_VERSION)/$$(go env GOOS)/$$(go env GOARCH)" \
		| tar -xz -C /tmp
	mv /tmp/kubebuilder_$(KUBEBUIDER_VERSION)_$$(go env GOOS)_$$(go env GOARCH)/bin/kubebuilder $(KUBEBUILDER)

# find or download kustomize
# download kustomize if necessary
KUSTOMIZE_VERSION := 3.8.8
KUSTOMIZE := $(BIN)/kustomize

.PHONY: kustomize
kustomize:
	@$(KUSTOMIZE) version --short 2>&1 \
		| grep 'kustomize/v$(KUSTOMIZE_VERSION)' > /dev/null \
	|| rm -f $(KUSTOMIZE)
	@$(MAKE) $(KUSTOMIZE)

$(KUSTOMIZE):
	$(MAKE) $(BIN)
	# https://kubectl.docs.kubernetes.io/installation/kustomize/binaries/
	curl -sSL "https://github.com/kubernetes-sigs/kustomize/releases/download/kustomize/v$(KUSTOMIZE_VERSION)/kustomize_v$(KUSTOMIZE_VERSION)_$$(go env GOOS)_$$(go env GOARCH).tar.gz" \
		| tar -xzC '$(BIN)' kustomize

# find helm or raise an error
.PHONY: helm
helm:
ifeq (, $(shell which helm 2> /dev/null))
	$(error Helm not found. Please install it: https://helm.sh/docs/intro/install/#from-script)
HELM=helm-not-found
else
HELM=$(shell which helm 2> /dev/null)
endif

# find or download goreleaser
GORELEASER_VERSION := 0.129.0
GORELEASER := $(BIN)/goreleaser

.PHONY: goreleaser
goreleaser:
	@$(GORELEASER) --version 2>&1 \
		| grep 'version: $(GORELEASER_VERSION)' > /dev/null \
	|| rm -f $(GORELEASER)
	@$(MAKE) $(GORELEASER)

$(GORELEASER):
	$(MAKE) $(BIN)
	# https://goreleaser.com/install/
	export BINDIR='$(BIN)' ; \
	curl -sfL https://install.goreleaser.com/github.com/goreleaser/goreleaser.sh \
		| bash -s 'v$(GORELEASER_VERSION)'

# find or download stringer
# download stringer if necessary
STRINGER_VERSION := v0.0.0-20201223010750-3fa0e8f87c1a
STRINGER := $(BIN)/stringer

.PHONY: stringer
stringer:
	$(warning stringer command has no `version` command)
	#@$(STRINGER) version 2>&1 \
	#	| grep '$(STRINGER_VERSION)' > /dev/null \
	#|| rm -f $(STRINGER)
	@$(MAKE) $(STRINGER)

$(STRINGER):
	$(MAKE) $(BIN)
	# https://pkg.go.dev/golang.org/x/tools/cmd/stringer
	go get 'golang.org/x/tools/cmd/stringer@$(STRINGER_VERSION)'
	go build -mod=readonly -o $(STRINGER) golang.org/x/tools/cmd/stringer
	go mod tidy

# find or download hadolint
# download hadolint if necessary
HADOLINT_VERSION := 1.18.0
HADOLINT := $(BIN)/hadolint

.PHONY: hadolint
hadolint:
	@$(HADOLINT) --version 2>&1 \
		| grep '$(HADOLINT_VERSION)' > /dev/null \
	|| rm -f $(HADOLINT)
	@$(MAKE) $(HADOLINT)

$(HADOLINT):
	$(MAKE) $(BIN)
	# https://github.com/hadolint/hadolint/releases/
	curl -sL "https://github.com/hadolint/hadolint/releases/download/v$(HADOLINT_VERSION)/hadolint-$$(uname -s)-x86_64" \
		> $(HADOLINT)
	chmod u+x $(HADOLINT)

# find or download helm-docs
# download helm-docs if necessary
HELM_DOCS_VERSION := 1.4.0
HELM_DOCS := $(BIN)/helm-docs

.PHONY: helm-docs
helm-docs:
	@$(HELM_DOCS) --version 2>&1 \
		| grep '$(HELM_DOCS_VERSION)' > /dev/null \
	|| rm -f $(HELM_DOCS)
	@$(MAKE) $(HELM_DOCS)

$(HELM_DOCS):
	$(MAKE) $(BIN)
	# https://github.com/norwoodj/helm-docs/tree/master#installation
	curl -sL "https://github.com/norwoodj/helm-docs/releases/download/v$(HELM_DOCS_VERSION)/helm-docs_$(HELM_DOCS_VERSION)_$$(uname -s)_x86_64.tar.gz" \
		| tar -xzC '$(BIN)' helm-docs
