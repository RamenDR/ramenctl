// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

// findNextSuffix finds previous files for the same command and selects the next available suffix
// to avoid overwriting previous command output.
//
// Notes:
//   - We ignore unexpected files (e.g. validate-application-error.log) since they cannot clash with
//     real output file (validate-application-2.log).
//   - We check all files matching the command name, so partial output will be considered and we
//     never overwrite an existing file.
func findNextSuffix(outDir, commandName string) (string, error) {
	// Glob for all files starting with commandName to handle multi-file output.
	pattern := filepath.Join(outDir, commandName+"*.*")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return "", err
	}

	// 0: No files, 1: Base file exists, N: Highest suffix found.
	last := 0

	for _, match := range matches {
		fileName := filepath.Base(match)
		if n := parseSequenceNumber(fileName, commandName); n > last {
			last = n
		}
	}

	// Result logic: 0 -> "", 1 -> "-2", N -> "-(N+1)".
	if last > 0 {
		return fmt.Sprintf("-%d", last+1), nil
	}

	return "", nil
}

// parseSequenceNumber extracts the sequence number from a filename based on the commandName.
// It returns:
//   - 1 for the base file (command-name.ext)
//   - N for numbered files (command-name-N.ext)
//   - 0 if the file does not follow the expected format and should be ignored.
func parseSequenceNumber(fileName, commandName string) int {
	base := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	suffix := strings.TrimPrefix(base, commandName)

	// The first run does not have a suffix (command-name.log).
	if suffix == "" {
		return 1
	}

	// Must start with '-' and have at least one character after it.
	if len(suffix) < 2 || suffix[0] != '-' {
		return 0
	}

	// Try to parse the numeric part, ignoring invalid number.
	n, err := strconv.Atoi(suffix[1:])
	if err != nil {
		return 0
	}

	return n
}
