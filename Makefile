PWD=$(shell pwd)

GOFILES= $$(go list -f '{{join .GoFiles " "}}')

.PHONY: mocks test

PATH:=$(PWD)/bin:${PATH}
export PATH

SHELL:=env PATH=$(PATH) /bin/bash

build_app:
	pkger
	go build -o $(GOPATH)/bin/release-dash $(GOFILES)

clean:
	@rm -rf vendor

deps:
	go mod vendor
	go get github.com/markbates/pkger/cmd/pkger
