export GO111MODULE=on

GINKGO_VERSION?=v1.14.2
KILLGRAVE_VERSION?=v0.4.0
PKGER_VERSION?=v0.17.1

PWD=$(shell pwd)

PATH:=$(PWD)/bin:${PATH}
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
deps:
	go mod tidy
	go mod download
	go get github.com/markbates/pkger/cmd/pkger@$(PKGER_VERSION)

deps_test:
	go get github.com/friendsofgo/killgrave/cmd/killgrave@$(KILLGRAVE_VERSION)
	go get github.com/onsi/ginkgo/ginkgo@$(GINKGO_VERSION)

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


.PHONY: docker_test
docker_test: docker_build
	docker build \
		--cache-from release-dash \
		-t release-dash-test \
		--target test .
	docker run \
		release-dash-test

.PHONY: mocks
mocks:
	rm -rf mocks
	go generate -v "-mod=mod" ./...

.PHONY: run
run: build
	release-dash

.PHONY: run_src
run_src: deps
	go run main.go

test: mocks
	go test -count 1 -v $(shell go list ./... | grep -v /e2e)
