SERVER_IMAGE    := ghcr.io/cloudoperators/heureka
VERSION  ?= $(shell git log -1 --pretty=format:"%H")
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

.PHONY: all test doc gqlgen mockery test-all test-e2e test-app test-db fmt compose-prepare compose-up compose-down compose-restart compose-build

# Source the .env file to use the env vars with make
-include .env
ifeq ($(wildcard .env),.env)
    $(info .env file found, exporting variables)
    export $(shell sed 's/=.*//' .env)
endif

build-binary:
	GO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o build/heureka cmd/heureka/main.go

# Build the binary and execute it
run-%:
	GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags="$(LDFLAGS)" -o build/$* cmd/$*/main.go
	DB_SCHEMA=./internal/database/mariadb/init/schema.sql ./build/$*

# Start ONLY the database container and replace the pg_hba_conf.sh created by `create-pg-hba-conf`
start: stop
	docker-compose --profile db up

# Start all container. This expects the heureka bin to be amd64 because the image in the docker-compose is amd64
start-all-%: stop
	docker-compose --profile db --profile heureka up  --build --force-recreate

stop:
	@rm -rf ./.db/data
	docker-compose --profile db --profile heureka down

clean:
	docker-compose down --rmi all --volumes --remove-orphans

echo:
	echo "version:" $(VERSION)

build-image:
	docker buildx build -t $(SERVER_IMAGE):$(VERSION) -t $(SERVER_IMAGE):latest . --load

build-scanner-images: build-scanner-k8s-assets-image build-scanner-keppel build-scanner-nvd

build-scanner-k8s-assets-image:
	docker buildx build -t $(SERVER_IMAGE)-scanner-k8s-assets:$(VERSION) -t $(SERVER_IMAGE)-scanner-k8s-assets:latest -f Dockerfile.scanner-k8s-assets . --load

build-scanner-keppel:
	docker buildx build -t $(SERVER_IMAGE)-scanner-keppel:$(VERSION) -t $(SERVER_IMAGE)-scanner-keppel:latest -f Dockerfile.scanner-keppel . --load

build-scanner-nvd:
	docker buildx build -t $(SERVER_IMAGE)-scanner-nvd:$(VERSION) -t $(SERVER_IMAGE)-scanner-nvd:latest -f Dockerfile.scanner-nvd . --load

push:
	docker push $(SERVER_IMAGE):$(VERSION)
	docker push $(SERVER_IMAGE):latest

run:
	go run cmd/heureka/main.go

gqlgen:
	cd internal/api/graphql && go run github.com/99designs/gqlgen generate

mockery:
	mockery

GINKGO := go run github.com/onsi/ginkgo/v2/ginkgo
test-all:
	$(GINKGO) -r

test-e2e:
	$(GINKGO) -r internal/e2e

test-app:
	$(GINKGO) -r internal/app

test-db:
	$(GINKGO) -r internal/database/mariadb

fmt:
	go fmt ./...

DOCKER_COMPOSE := docker-compose -f docker-compose.yaml
DOCKER_COMPOSE_SERVICES := heureka-app heureka-db
compose-prepare:
	sed 's/^SEED_MODE=false/SEED_MODE=true/g' .test.env > .env

compose-up:
	$(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_SERVICES)

compose-down:
	$(DOCKER_COMPOSE) down $(DOCKER_COMPOSE_SERVICES)

compose-restart: compose-down compose-up

compose-build:
	$(DOCKER_COMPOSE) build $(DOCKER_COMPOSE_SERVICES)

##@ Deployment

ifndef ignore-not-found
  ignore-not-found = false
endif

.PHONY: kustomize-build-crds
kustomize-build-crds: generate-manifests kustomize
	$(KUSTOMIZE) build $(CRD_MANIFESTS_PATH)
	
##@ Build Dependencies

## Tool Binaries
KUSTOMIZE ?= $(LOCALBIN)/kustomize
CONTROLLER_GEN ?= $(LOCALBIN)/controller-gen
GOIMPORTS ?= $(LOCALBIN)/goimports
GOLINT ?= $(LOCALBIN)/golangci-lint
ENVTEST ?= $(LOCALBIN)/setup-envtest
HELMIFY ?= $(LOCALBIN)/helmify

## Tool Versions
KUSTOMIZE_VERSION ?= v5.4.2
CONTROLLER_TOOLS_VERSION ?= v0.15.0
GOLINT_VERSION ?= v1.60.2
GINKGOLINTER_VERSION ?= v0.16.2

KUSTOMIZE_INSTALL_SCRIPT ?= "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"
.PHONY: kustomize
kustomize: $(KUSTOMIZE) ## Download kustomize locally if necessary.
$(KUSTOMIZE): $(LOCALBIN)
	test -s $(LOCALBIN)/kustomize || curl -s $(KUSTOMIZE_INSTALL_SCRIPT) | bash -s -- $(subst v,,$(KUSTOMIZE_VERSION)) $(LOCALBIN)

.PHONY: controller-gen
controller-gen: $(CONTROLLER_GEN) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: controller-gen-docker
controller-gen-docker: $(CONTROLLER_GEN_DOCKER) ## Download controller-gen locally if necessary.
$(CONTROLLER_GEN_DOCKER): $(LOCALBIN)
	GOPATH=$(shell pwd) go install sigs.k8s.io/controller-tools/cmd/controller-gen@$(CONTROLLER_TOOLS_VERSION)

.PHONY: envtest
envtest: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: envtest-docker
envtest-docker: $(ENVTEST) ## Download envtest-setup locally if necessary.
$(ENVTEST_DOCKER): $(LOCALBIN)
	GOPATH=$(shell pwd) go install sigs.k8s.io/controller-runtime/tools/setup-envtest@latest

.PHONY: goimports
goimports: $(GOIMPORTS)
$(GOIMPORTS): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install golang.org/x/tools/cmd/goimports@latest

.PHONY: golint
golint: $(GOLINT)
$(GOLINT): $(LOCALBIN)
	GOBIN=$(LOCALBIN) go install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLINT_VERSION)
	GOBIN=$(LOCALBIN) go install github.com/nunnatsa/ginkgolinter/cmd/ginkgolinter@$(GINKGOLINTER_VERSION)

.PHONY: serve-docs
serve-docs: generate-manifests
ifeq (, $(shell which hugo))
	@echo "Hugo is not installed in your machine. Please install it to serve the documentation locally. Please refer to https://gohugo.io/installation/ for installation instructions."
else
	cd website && hugo server
endif