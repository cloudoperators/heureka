SERVER_IMAGE    := ghcr.io/cloudoperators/heureka
VERSION  ?= $(shell git log -1 --pretty=format:"%H")
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

.PHONY: all test doc

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
	docker buildx -t $(SERVER_IMAGE):$(VERSION) -t $(SERVER_IMAGE):latest . --load

push:
	docker push $(SERVER_IMAGE):$(VERSION)
	docker push $(SERVER_IMAGE):latest

run:
	go run cmd/heureka/main.go
