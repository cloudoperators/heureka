SERVER_IMAGE    := ghcr.io/cloudoperators/heureka
VERSION  ?= $(shell git log -1 --pretty=format:"%H")
OS := $(shell go env GOOS)
ARCH := $(shell go env GOARCH)

.PHONY: run generate-client test-all test-scanner test-processor test-all-mods test-keystone test-nova

# Source the .env file to use the env vars with make
-include .env
ifeq ($(wildcard .env),.env)
    $(info .env file found, exporting variables)
    export $(shell sed 's/=.*//' .env)
endif

# Build the binary and execute it
run-%:
	GOOS=$(OS) GOARCH=$(ARCH) go build -ldflags="$(LDFLAGS)" -o build/$* main.go

run: 
	go run main.go

generate-client: 
	cd client && go run github.com/Khan/genqlient

GINKGO := go run github.com/onsi/ginkgo/v2/ginkgo
test-all:
	$(GINKGO) -r

test-scanner:
	$(GINKGO) -r scanner

test-processor:
	$(GINKGO) -r processor

test-all-mods:
	$(GINKGO) -r modules

test-keystone:
	$(GINKGO) -r keystone

test-nova:
	$(GINKGO) -r nova
