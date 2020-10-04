PWD=$(shell pwd)

GOFILES= $$(go list -f '{{join .GoFiles " "}}')

SHELL:=env /bin/bash

.PHONY: build
build:
	pkger
	go build -o $(GOPATH)/bin/release-dash $(GOFILES)

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

.PHONY: run
run: deps
	go run main.go