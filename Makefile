# SPDX-FileCopyrightText: The RamenDR authors
# SPDX-License-Identifier: Apache-2.0

GO ?= go

# v0.5.1 when building from tag (release)
# v0.5.1-1-gcf79160 when building without tag (development)
version := $(shell git describe --tags)
commit := $(shell git rev-parse HEAD)

build := github.com/ramendr/ramenctl/pkg/build

# % go build -ldflags="-help"
#  -X 	definition
#    	add string value definition of the form importpath.name=value
ldflags := -X '$(build).Version=$(version)' \
		   -X '$(build).Commit=$(commit)'

.PHONY: ramenctl examples test clean coverage lint fmt

all: ramenctl examples

ramenctl:
	CGO_ENABLED=0 $(GO) build -ldflags="$(ldflags)" cmd/ramenctl.go

examples:
	$(GO) build -o examples/odf examples/odf.go

fmt:
	golangci-lint fmt

spell:
	codespell --skip="go.sum"

spell-fix:
	codespell --skip="go.sum" --write-changes

lint:
	golangci-lint run ./...

test:
	$(GO) test --coverprofile cover.out -ldflags="$(ldflags)" -v ./...

coverage:
	$(GO) tool cover -html=cover.out

clean:
	rm -f ramenctl examples/odf
