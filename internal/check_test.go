// internal/tools.go
//go:build tools
// +build tools

package internal

import (
	_ "go.uber.org/zap"
	_ "go.uber.org/zap/zapcore"
)
