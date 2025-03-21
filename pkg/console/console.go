// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package console

import (
	"fmt"
	"os"
	"strings"
)

// TODO: Measure actual console and update when window size changes
var width = 80

func Info(format string, args ...any) {
	fmt.Printf("⭐ "+format+"\n", args...)
}

func Progress(format string, args ...any) {
	line := fmt.Sprintf("👀 "+format+" ...", args...)
	fmt.Print(pad(line), "\r")
}

func Completed(format string, args ...any) {
	line := fmt.Sprintf("✅ "+format, args...)
	fmt.Print(pad(line), "\n")
}

func Error(err error) {
	fmt.Fprintf(os.Stderr, "❌ %s\n", err)
}

func Fatal(err error) {
	Error(err)
	os.Exit(1)
}

// pad string to console width, leaving room for the line terminator.
func pad(s string) string {
	if len(s) < width-1 {
		return s + strings.Repeat(" ", width-len(s)-1)
	}
	return s
}
