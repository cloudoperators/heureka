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

all: build-binary test-all

build-binary: gqlgen mockery
	GO_ENABLED=0 go build -ldflags="$(LDFLAGS)" -o build/heureka cmd/heureka/main.go

# Build the binary and execute it
run-%: gqlgen mockery
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
	go install github.com/vektra/mockery/v2@v2.52.2
	go install github.com/onsi/ginkgo/v2/ginkgo
	go install github.com/99designs/gqlgen

test-all: mockery gqlgen install-build-dependencies
	ginkgo -r

test-e2e: gqlgen install-build-dependencies
	ginkgo -r internal/e2e

test-app: gqlgen install-build-dependencies
	ginkgo -r internal/app

test-access: gqlgen install-build-dependencies
	ginkgo -r internal/api/graphql/access

test-db: gqlgen install-build-dependencies
	ginkgo -r internal/database/mariadb

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

install-migrate:
	go install -tags 'heureka-migration' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

create-migration:
	@(test -v MIGRATION_NAME && migrate create -ext sql -dir internal/database/mariadb/migrations ${MIGRATION_NAME}) || echo MIGRATION_NAME not specified >&2
