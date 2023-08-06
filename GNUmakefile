TEST?=$$(go list ./... | grep -v 'vendor')
HOSTNAME=hashicorp.com
NAMESPACE=kislerdm
NAME=neon
VERSION=dev
BINARY=terraform-provider-${NAME}_v${VERSION}
OS_ARCH=darwin_arm64

.PHONY: testacc build install test

help: ## Prints help message.
	@ grep -h -E '^[a-zA-Z0-9_-].+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[1m%-30s\033[0m %s\n", $$1, $$2}'

default: help

build:
	@ go build -a -gcflags=all="-l -B -C" -ldflags="-w -s" -o ${BINARY} .

install: build ## Builds and installs the provider.
	@ mkdir -p ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}
	@ mv ${BINARY} ~/.terraform.d/plugins/${HOSTNAME}/${NAMESPACE}/${NAME}/${VERSION}/${OS_ARCH}/

test: ## Runs unit tests.
	@ go test $(TEST) || exit 1
	@ echo $(TEST) | xargs -t -n4 go test $(TESTARGS) -timeout=30s -parallel=4

testacc: ## Runs acceptance tests.
	@ TF_ACC=1 go test ./... -v -tags acceptance -timeout 120m

docu: ## Generates docu.
	@ go generate
