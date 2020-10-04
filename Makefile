export GO111MODULE=on

PWD=$(shell pwd)

GOFILES= $$(go list -f '{{join .GoFiles " "}}')

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
	go get github.com/markbates/pkger/cmd/pkger

.PHONY: docker_build
docker_build:
	docker build -t release-dash .

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
	go generate -v ./...

.PHONY: run
run: build
	release-dash

.PHONY: run_src
run_src: deps
	go run main.go

test: mocks
	go test -v ./...