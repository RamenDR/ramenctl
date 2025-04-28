// SPDX-FileCopyrightText: The RamenDR authors
// SPDX-License-Identifier: Apache-2.0

package command

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func newLogger(outputDir, commandName string) (*zap.SugaredLogger, func(), error) {
	if err := os.MkdirAll(outputDir, 0o750); err != nil {
		return nil, nil, err
	}

	path := filepath.Join(outputDir, commandName)
	writer, closeFile, err := zap.Open(path)
	if err != nil {
		return nil, nil, err
	}

	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	encoder := zapcore.NewConsoleEncoder(encoderConfig)

	core := zapcore.NewCore(encoder, writer, zapcore.DebugLevel)
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return logger.Sugar(), closeFile, nil
}
