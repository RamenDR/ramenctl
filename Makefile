# SPDX-FileCopyrightText: The RamenDR authors
# SPDX-License-Identifier: Apache-2.0

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

.PHONY: ramenctl examples test clean coverage

all: ramenctl examples

ramenctl:
	CGO_ENABLED=0 go build -ldflags="$(ldflags)" cmd/ramenctl.go

examples:
	go build -o examples/odf examples/odf.go

test:
	go test --coverprofile cover.out -ldflags="$(ldflags)" -v ./...

coverage:
	go tool cover -html=cover.out

clean:
	rm -f ramenctl examples/odf
