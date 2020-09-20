PWD=$(shell pwd)

GOFILES= $$(go list -f '{{join .GoFiles " "}}')

.PHONY: mocks test

PATH:=$(PWD)/bin:${PATH}
export PATH

SHELL:=env PATH=$(PATH) /bin/bash

build_app:
	go build -o $(GOPATH)/bin/release-dash $(GOFILES)

deps:
	go mod tidy
	go mod download
