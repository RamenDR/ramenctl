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

// Error logs single operation error.
func Error(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "   ❌ "+format+"\n", args...)
}

// Completed logs command completion.
func Completed(format string, args ...any) {
	fmt.Printf("\n✅ "+format+"\n", args...)
}

// Hint logs a suggestion to the user, indented under the previous message.
func Hint(format string, args ...any) {
	fmt.Printf("   "+format+"\n", args...)
}

// Fatal logs command failure and exit.
func Fatal(err error) {
	fmt.Fprintf(os.Stderr, "\n❌ %s\n", err)
	os.Exit(1)
}
