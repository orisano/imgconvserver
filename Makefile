NAME := imgconvserver
VERSION := $(shell git tag -l | tail -1)
REVISION := $(shell git rev-parse --short HEAD)
LDFLAGS := -X 'main.version=$(VERSION)' -X 'main.revision=$(REVISION)'
PACKAGENAME := github.com/akito0107/imgconvserver

.PHONY: setup dep test test/internal main clean install lint lint/internal

all: main

main:
	go build -ldflags "$(LDFLAGS)" -o bin/server cmd/server/*.go

## Install dependencies
setup:
	go get -u github.com/golang/dep/cmd/dep

## install go dependencies
dep:
	dep ensure

## remove build files
clean:
	rm -rf ./bin/*
