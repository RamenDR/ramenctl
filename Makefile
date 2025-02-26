# SPDX-FileCopyrightText: The RamenDR authors
# SPDX-License-Identifier: Apache-2.0

.PHONY: ramenctl examples

all: ramenctl examples

ramenctl:
	go build -o $@ cmd/main.go

examples:
	go build -o examples/odf examples/odf.go

clean:
	rm -f ramenctl examples/odf
