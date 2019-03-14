
.PHONY: run
run: ## Runs the service in development and watches for changes
ifeq ($(shell which modd),)
	brew install modd
endif
	modd

.PHONY: test
test: ## Runs the tests and higlights race conditions
	GO111MODULE=on go test --count=1 -race $$(go list ./...)

SERVICE ?= demo_messenger
NOW=$(shell date -u '+%Y-%m-%d_%I:%M:%S%p')
VERSION=$(shell cat VERSION.txt)

LDFLAGS="-X main.version=$(VERSION) -X main.buildstamp=$(NOW) -X main.creator=$(CREATOR)"
DIST_PATH=dist

.PHONY: build
build: ## Builds the executable for linux amd64
	rm -rf $(DIST_PATH)
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
		go build -a -installsuffix cgo \
		-ldflags $(LDFLAGS) \
		-o $(DIST_PATH)/$(SERVICE) ./cmd/$(SERVICE)

.PHONY: lint
lint: ## Runs more than 20 different linters using golangci-lint to ensure consistency in code.
ifeq ($(shell which golangci-lint),)
	brew install golangci/tap/golangci-lint
endif
	golangci-lint run

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
