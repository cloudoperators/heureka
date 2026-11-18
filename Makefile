SERVER_IMAGE    := ghcr.io/cloudoperators/heureka
VERSION  ?= $(shell git log -1 --pretty=format:"%H")
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

SHELL=/bin/bash
MIGRATIONS_DIR=internal/database/mariadb/migrations/

.PHONY: all test doc gqlgen mockery test-all test-e2e test-app test-db lint fmt pre-commit-prepare compose-prepare compose-up compose-down compose-restart compose-build check-call-cached check-migration-files

# Source the .env file to use the env vars with make
-include .env
ifeq ($(wildcard .env),.env)
    $(info .env file found, exporting variables)
    export $(shell sed 's/=.*//' .env)
endif

all: build-binary test-all

build-binary: mockery gqlgen
	GO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o build/heureka cmd/heureka/main.go

# Build the binary and execute it
run-%: mockery gqlgen
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

build-scanner-images: build-scanner-k8s-assets-image build-scanner-keppel-image build-scanner-nvd-image

build-scanner-k8s-assets-image:
	docker buildx build -t $(SERVER_IMAGE)-scanner-k8s-assets:$(VERSION) -t $(SERVER_IMAGE)-scanner-k8s-assets:latest -f ./scanner/k8s-assets/Dockerfile . --load

build-scanner-keppel-image:
	docker buildx build -t $(SERVER_IMAGE)-scanner-keppel:$(VERSION) -t $(SERVER_IMAGE)-scanner-keppel:latest -f ./scanner/keppel/Dockerfile . --load

build-scanner-nvd-image:
	docker buildx build -t $(SERVER_IMAGE)-scanner-nvd:$(VERSION) -t $(SERVER_IMAGE)-scanner-nvd:latest -f ./scanner/nvd/Dockerfile . --load

push:
	docker push $(SERVER_IMAGE):$(VERSION)
	docker push $(SERVER_IMAGE):latest

run:
	go run cmd/heureka/main.go

gqlgen: install-build-dependencies gqlclient
	cd internal/api/graphql && gqlgen generate

gqlclient:
	cd scanner/k8s-assets/client && go run github.com/Khan/genqlient

mockery: install-build-dependencies
	mockery

install-build-dependencies:
	go install github.com/vektra/mockery/v3@v3.8.1
	go install github.com/onsi/ginkgo/v2/ginkgo
	go install github.com/99designs/gqlgen@v0.17.95

test-all: mockery gqlgen install-build-dependencies
	ginkgo -r

test-e2e: gqlgen install-build-dependencies
	ginkgo -r internal/e2e

test-app: gqlgen install-build-dependencies
	ginkgo -r internal/app

test-access: gqlgen install-build-dependencies
	ginkgo -r internal/api/graphql/access

test-authorization: gqlgen install-build-dependencies
	ginkgo -r internal/openfga

test-db: gqlgen install-build-dependencies
	ginkgo -r internal/database/mariadb

install-gofumpt:
	go install mvdan.cc/gofumpt@latest

fmt: install-gofumpt
	gofumpt -l -w internal/

install-golangci-lint:
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.13.2

lint: install-golangci-lint
	golangci-lint run

lint-fix: install-golangci-lint
	golangci-lint run --fix

install-pre-commit:
	@echo "Checking pre-commit installation..."

	@if command -v pre-commit >/dev/null 2>&1; then \
		echo "pre-commit already installed"; \
	else \
		echo "Installing pre-commit..."; \
		if [ "$$(uname)" = "Darwin" ]; then \
			if command -v brew >/dev/null 2>&1; then \
				brew install pre-commit; \
			else \
				pip3 install pre-commit; \
			fi; \
		elif [ "$$(uname)" = "Linux" ]; then \
			if command -v pip3 >/dev/null 2>&1; then \
				pip3 install pre-commit; \
			else \
				echo "pip3 not found, please install Python/pip"; \
				exit 1; \
			fi; \
		elif echo "$$(uname)" | grep -qi "mingw\\|msys\\|cygwin"; then \
			pip install pre-commit; \
		else \
			echo "Unknown OS, please install pre-commit manually"; \
			exit 1; \
		fi; \
	fi

pre-commit-prepare: install-pre-commit
	pre-commit install
	pre-commit install --hook-type pre-push

DOCKER_COMPOSE := docker-compose -f docker-compose.yaml
DOCKER_COMPOSE_SERVICES := heureka-app heureka-db valkey
compose-prepare:
	sed 's/^SEED_MODE=false/SEED_MODE=true/g' .test.env > .env

compose-up:
	$(DOCKER_COMPOSE) up -d $(DOCKER_COMPOSE_SERVICES)

compose-down:
	$(DOCKER_COMPOSE) down $(DOCKER_COMPOSE_SERVICES)

compose-restart: compose-down compose-up

compose-build:
	$(DOCKER_COMPOSE) build $(DOCKER_COMPOSE_SERVICES)

install-migrate:
	go install -tags 'heureka-migration' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

create-migration:
	@(test -v MIGRATION_NAME && migrate create -ext sql -dir internal/database/mariadb/migrations ${MIGRATION_NAME}) || echo MIGRATION_NAME not specified >&2

check-call-cached:
	go generate ./internal/cache

# check-migration-files target:
# First check if '!checkexistingmigration' smart commit tag is not used in any commit on your branch
# if it is used, then skip further check
# if smart commit was not found, check for modification of any migration file existing in main branch
# NOTE: new files created in your branch will not be checked for modification/creation
check-migration-files:
	@git log origin/main..HEAD --pretty=%B | grep -Eq "[!]changeexistingmigration" || \
	(comm -12 <(git diff --name-only origin/main...HEAD | grep '^$(MIGRATIONS_DIR)' | sort) <(git ls-tree -r --name-only origin/main | grep '^$(MIGRATIONS_DIR)' | sort) | grep -q . && { echo "Forbidden change detected!"; exit 1; } || echo "No forbidden changes.")

ui-up:
	$(DOCKER_COMPOSE) --profile ui up -d

ui-down:
	$(DOCKER_COMPOSE) --profile ui down

openfga-up:
	$(DOCKER_COMPOSE) --profile openfga up

openfga-down:
	$(DOCKER_COMPOSE) --profile openfga down

# ---------------------------------------------------------------------------
# Local scanner development environment (kind + keppel mock).
# See docs/local-environment.md for the full workflow.
# ---------------------------------------------------------------------------
KIND_CLUSTER    := heureka-local
KIND_CONTEXT    := kind-$(KIND_CLUSTER)
KIND_CONFIG     := local/kind/cluster.yaml
KIND_KUBECONFIG := $(CURDIR)/local/kind/kubeconfig
PLAYGROUND_CHART := ./scanner/playground/k8s-playground

.PHONY: kind-up kind-down playground-up playground-down scan-k8s-assets-local keppel-mock-up keppel-mock-record scan-keppel-local local-up local-down

# Create the kind cluster and export its kubeconfig to a fixed, git-ignored path.
kind-up:
	kind create cluster --config $(KIND_CONFIG)
	kind export kubeconfig --name $(KIND_CLUSTER) --kubeconfig $(KIND_KUBECONFIG)

kind-down:
	kind delete cluster --name $(KIND_CLUSTER)

# Install the sample namespaces/pods the k8s-assets scanner scans.
playground-up:
	helm upgrade --install k8s-playground $(PLAYGROUND_CHART) --kube-context $(KIND_CONTEXT)

playground-down:
	helm uninstall k8s-playground --kube-context $(KIND_CONTEXT)

# Run the k8s-assets scanner on the host against the kind cluster, reporting to
# the docker-compose heureka backend. Requires: kind-up, playground-up, compose-up.
scan-k8s-assets-local:
	cd scanner/k8s-assets && \
		KUBE_CONFIG_TYPE=default \
		KUBE_CONFIG_PATH=$(KIND_KUBECONFIG) \
		HEUREKA_URL=http://localhost/query \
		HEUREKA_CLUSTER_NAME=local-kind \
		HEUREKA_CLUSTER_REGION=local \
		DEFAULT_KEPPEL_REGISTRY=keppel.example.com \
		LOG_LEVEL=debug \
		go run .

# Start the keppel mock (Keystone on :8080 HTTP, Keppel on :8443 HTTPS).
# Runs in the foreground; use a separate terminal.
keppel-mock-up:
	cd scanner/keppel/mock && go run .

# Refresh the mock's fixtures from a real Keppel. Requires real KEPPEL_* creds
# plus RECORD_IDENTITY_UPSTREAM and RECORD_KEPPEL_UPSTREAM in the environment.
keppel-mock-record:
	cd scanner/keppel/mock && KEPPEL_RECORD=1 go run .

# Run the keppel scanner on the host against the local mock, reporting to the
# docker-compose heureka backend. Requires: keppel-mock-up, compose-up, and
# image-shaped components in heureka (run scan-k8s-assets-local first).
scan-keppel-local:
	cd scanner/keppel && \
		KEPPEL_FQDN=localhost:8443 \
		KEPPEL_FQDN_ALIASES=docker.io \
		IDENTITY_ENDPOINT=http://localhost:8080/v3 \
		KEPPEL_USERNAME=dummy KEPPEL_PASSWORD=dummy KEPPEL_USER_DOMAIN=dummy \
		KEPPEL_DOMAIN=dummy KEPPEL_PROJECT=dummy \
		HEUREKA_URL=http://localhost/query \
		LOG_LEVEL=debug \
		go run .

# Bring the whole local environment up (backend + cluster + sample workloads).
local-up: compose-prepare compose-up kind-up playground-up

# Tear it down WITHOUT deleting the DB (deliberately not `make stop`/`clean`).
local-down: playground-down kind-down compose-down

