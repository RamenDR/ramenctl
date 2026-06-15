// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

// htmlfmt formats HTML files for readability.
//
// Usage:
//
//	go run ./tools/htmlfmt file.html
package main

import (
	"fmt"
	"os"

	"github.com/yosssi/gohtml"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <file.html>\n", os.Args[0])
		os.Exit(1)
	}

	filename := os.Args[1]

	data, err := os.ReadFile(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filename, err)
		os.Exit(1)
	}

	gohtml.Condense = true
	formatted := gohtml.Format(string(data)) + "\n"

	if err := os.WriteFile(filename, []byte(formatted), 0o640); err != nil {
		fmt.Fprintf(os.Stderr, "Error writing %s: %v\n", filename, err)
		os.Exit(1)
	}
}
