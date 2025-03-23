// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package console

import (
	"fmt"
	"os"
)

func Info(format string, args ...any) {
	fmt.Printf("⭐ "+format+"\n", args...)
}

// Step logs a new command step.
func Step(format string, args ...any) {
	fmt.Printf("\n🔎 "+format+" ...\n", args...)
}

// Pass logs single operation completion.
func Pass(format string, args ...any) {
	fmt.Printf("   ✅ "+format+"\n", args...)
}

// Fail log single operation error.
func Error(err error) {
	fmt.Fprintf(os.Stderr, "   ❌ %s\n", err)
}

// Completed logs command completion.
func Completed(format string, args ...any) {
	fmt.Printf("\n✅ "+format+"\n", args...)
}

// Fatal logs command failure and exit.
func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "\n❌ %s\n", err)
	os.Exit(1)
}
