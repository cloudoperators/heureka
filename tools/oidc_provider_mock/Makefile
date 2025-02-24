IMAGE_NAME ?= oidc-provider-mock

.DEFAULT_GOAL := help
.PHONY: *

docker-build: ## Builds Docker image
	docker build \
	--pull \
	--tag="$(IMAGE_NAME):latest" \
	.

run: ## Run golang app in console
	go run main.go

build: ## Build golang app
	go build -o oidc-provider-mock.bin .

tidy: ## Update go.mod using go mod tidy
	go mod tidy

help: ## Displays this help screen
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[a-zA-Z_-]+:.*?##/ \
	{ printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 }' $(MAKEFILE_LIST)
