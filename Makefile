
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

GO_SOURCES                  := $(sort $(subst $(CURDIR)/,,$(shell go list -f '$(gosourcetemplate)' ./...)))
GONOGENERATED_SOURCES       := $(sort $(shell grep -L 'DO NOT EDIT.' -- $(GO_SOURCES)))
GOWITHTESTS_SOURCES         := $(sort $(subst $(CURDIR)/,,$(shell go list -test -f '$(gosourcetemplate)' ./...)))
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
run: go-generate vendor certmanager $(TMPDIR)k8s-webhook-server/serving-certs/tls.crt
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
#      Tests        #
#####################

.PHONY: go-dependencies-test
go-dependencies-test: fmt
	go mod tidy
	$(MAKE) vendor
	$(MAKE) diff

.PHONY: generated-diff-test
generated-diff-test: fmt generate
	$(MAKE) diff

.PHONY: diff
diff:
	git status
	git diff
	test -z "$$(git diff-index --diff-filter=d --name-only HEAD)"

.PHONY: go-test
go-test: go-generate
ifeq (, $(USE_EXISTING_CLUSTER))
	echo "Warning: USE_EXISTING_CLUSTER variable is not defined" >&2
endif
	go test -vet=off ./... \
		-coverprofile cover.out

.PHONY: release
release-test: goreleaser
	$(GORELEASER) release --rm-dist --snapshot

CHART_RELEASE_NAME ?= harbor-operator
CHART_HARBOR_CLASS ?=

helm-install: helm helm-generate
	$(HELM) upgrade --install $(CHART_RELEASE_NAME) $(CHARTS_DIRECTORY)/harbor-operator-$(RELEASE_VERSION).tgz \
		--set-string image.repository="$(shell echo $(IMG) | sed 's/:.*//')" \
		--set-string image.tag="$(shell echo $(IMG) | sed 's/.*://')" \
		--set-string harborClass='${CHART_HARBOR_CLASS}'

#####################
#     Packaging     #
#####################

# Build manager binary
.PHONY: manager
manager: go-generate vendor
	go build \
		-mod vendor \
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

vendor: go.mod go.sum
	go mod vendor

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
		-mod vendor \
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
		--ignore "$(CURDIR)/vendor" \
		--ignore "$(CURDIR)/node_modules" \
		"$(CURDIR)"

docker-lint: hadolint
	$(HADOLINT) Dockerfile

make-lint: checkmake
	$(CHECKMAKE) Makefile

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
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/deployment.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/deployment | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Deployment' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/deployment.yaml
	cat config/helm/deployment/foot.yaml >> $(CHART_TEMPLATE_PATH)/deployment.yaml

$(CHART_TEMPLATE_PATH)/role.yaml: kustomize $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/role.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/role.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Role' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=ClusterRole' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=RoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/role.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/role.yaml

$(CHART_TEMPLATE_PATH)/clusterrole.yaml: kustomize $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/clusterrole.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/clusterrole.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=ClusterRole' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=ClusterRoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/clusterrole.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/clusterrole.yaml

$(CHART_TEMPLATE_PATH)/rolebinding.yaml: kustomize $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/rolebinding.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/rolebinding.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=RoleBinding' | \
	$(KUSTOMIZE) cfg grep --annotate=false --invert-match 'kind=ClusterRoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/rolebinding.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/rolebinding.yaml

$(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml: kustomize $(wildcard config/helm/rbac/*) $(wildcard config/rbac/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml
	echo '{{- if .Values.rbac.create }}' >> $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/rbac | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=ClusterRoleBinding' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml
	echo '{{- end -}}' >> $(CHART_TEMPLATE_PATH)/clusterrolebinding.yaml

$(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml: kustomize $(wildcard config/helm/webhook/*) $(wildcard config/webhook/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/webhook | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=ValidatingWebhookConfiguration' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/validatingwebhookconfiguration.yaml

$(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml: kustomize $(wildcard config/helm/webhook/*) $(wildcard config/webhook/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/webhook | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=MutatingWebhookConfiguration' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/mutatingwebhookconfiguration.yaml

$(CHART_TEMPLATE_PATH)/certificate.yaml: kustomize $(wildcard config/helm/certmanager/*) $(wildcard config/certmanager/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/certificate.yaml
	$(KUSTOMIZE) build --reorder legacy config/helm/certificate | \
	$(KUSTOMIZE) cfg grep --annotate=false 'kind=Certificate' | \
	sed "s/'\({{[^}}]*}}\)'/\1/g" \
		>> $(CHART_TEMPLATE_PATH)/certificate.yaml

$(CHART_TEMPLATE_PATH)/issuer.yaml: kustomize $(wildcard config/helm/certmanager/*) $(wildcard config/certmanager/*)
	echo '# $(DO_NOT_EDIT)' > $(CHART_TEMPLATE_PATH)/issuer.yaml
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
install: go-generate kustomize
	kubectl apply -f config/crd/bases

# Uninstall CRDs from a cluster
.PHONY: uninstall
uninstall: go-generate kustomize
	kubectl delete -f config/crd/bases

go-generate: controller-gen stringer
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
	kubectl get namespace $(NAMESPACE) 2>&1 > /dev/null || kubectl create namespace $(NAMESPACE)

INGRESS_NAMESPACE := nginx-ingress

.PHONY: ingress
ingress: helm
	$(MAKE) kube-namespace NAMESPACE=$(INGRESS_NAMESPACE)
	$(HELM) upgrade --install nginx stable/nginx-ingress \
		--namespace $(INGRESS_NAMESPACE) \
		--set-string controller.config.proxy-body-size=0

CERTMANAGER_NAMESPACE := cert-manager

.PHONY: certmanager
certmanager: helm jetstack
	$(MAKE) kube-namespace NAMESPACE=$(CERTMANAGER_NAMESPACE)
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
	$(TMPDIR)k8s-webhook-server/serving-certs/tls.crt

$(TMPDIR)k8s-webhook-server/serving-certs/tls.crt:
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
	GO_DEPENDENCY='github.com/golangci/golangci-lint/cmd/golangci-lint@e2d717b873ff02afab1903f34889cb8b621d7723' $(MAKE) go-binary
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

# find or download helm-docs
# download helm-docs if necessary
.PHONY: helm-docs
helm-docs:
ifeq (, $(shell which helm-docs))
	# https://github.com/norwoodj/helm-docs/tree/master#installation
	GO_DEPENDENCY='github.com/norwoodj/helm-docs/cmd/helm-docs@v0.15.0' $(MAKE) go-binary
HELM_DOCS=$(GOBIN)/helm-docs
else
HELM_DOCS=$(shell which helm-docs)
endif
