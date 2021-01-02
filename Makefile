export GO111MODULE=on

GINKGO_VERSION?=v1.14.2
GOMOCK_VERSION?=v1.4.4
KILLGRAVE_VERSION?=v0.4.0
PKGER_VERSION?=v0.17.1

PWD=$(shell pwd)

PATH:=$(PWD)/bin:$(PATH)
export PATH

export BUILDKIT_PROGRESS=plain
export DOCKER_BUILDKIT=1

SHELL:=env PATH=$(PATH) /bin/bash

.PHONY: build
build:
	pkger
	go install -v .

.PHONY: clean
clean:
	@rm -rf vendor

.PHONY: deps
deps: deps_tools mocks
	$(MAKE) mocks
	go mod download

.PHONY: deps_test
deps_test:
	go get github.com/friendsofgo/killgrave/cmd/killgrave@$(KILLGRAVE_VERSION)
	go get github.com/onsi/ginkgo/ginkgo@$(GINKGO_VERSION)

.PHONY: deps_tools
deps_tools:
	go get github.com/golang/mock/mockgen@$(GOMOCK_VERSION)
	go get github.com/markbates/pkger/cmd/pkger@$(PKGER_VERSION)

.PHONY: docker_build
docker_build:
	docker build \
		-t release-dash \
		--target run \
		.

.PHONY: docker_run
docker_run: docker_build
	docker run \
		-dit \
		-e GITHUB_PAT \
		-p 8080:8080 \
		release-dash

.PHONY: mocks
mocks:
	rm -rf mocks
	go generate -v "-mod=mod" ./...

.PHONY: run
run: build
	@release-dash

.PHONY: run_src
run_src: deps
	go run main.go

.PHONY: test_all
test_all: deps deps_test mocks test_unit test_integration

.PHONY: test_integration
test_integration:
	go test -count 1 -timeout=120s -cover -race -v -tags=integration ./testintegration

.PHONY: test_unit
test_unit: mocks
	go test -count 1 -timeout=30s -cover -race -v ./...
