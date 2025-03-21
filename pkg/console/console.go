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

func Progress(format string, args ...any) {
	fmt.Printf("🔎 "+format+" ...\n", args...)
}

func Completed(format string, args ...any) {
	fmt.Printf("✅ "+format+"\n", args...)
}

func Error(err error) {
	fmt.Fprintf(os.Stderr, "❌ %s\n", err)
}

func Fatal(err error) {
	Error(err)
	os.Exit(1)
}
