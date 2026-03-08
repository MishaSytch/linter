package util

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"testing"
)

// MockReporter мок для Reporter
type MockReporter struct {
	reports   *[]string
	typesInfo *types.Info
}

func (m *MockReporter) Reportf(pos token.Pos, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	*m.reports = append(*m.reports, msg)
}

func (m *MockReporter) Report(d analysis.Diagnostic) {
	*m.reports = append(*m.reports, d.Message)
}

func (m *MockReporter) TypesInfo() *types.Info {
	return m.typesInfo
}

func TestIsFmtSprintf(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "valid fmt.Sprintf",
			code: `fmt.Sprintf("hello %s", "world")`,
			want: true,
		},
		{
			name: "different function in fmt",
			code: `fmt.Printf("hello")`,
			want: false,
		},
		{
			name: "same function name, different package",
			code: `myslog.Sprintf("hello")`,
			want: false,
		},
		{
			name: "not a selector expression",
			code: `Sprintf("hello")`,
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _ := parser.ParseExpr(tt.code)
			call, ok := expr.(*ast.CallExpr)
			if !ok {
				t.Fatalf("test code is not a call expression: %s", tt.code)
			}

			if got := IsFmtSprintf(call); got != tt.want {
				t.Errorf("IsFmtSprintf() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsLogCall(t *testing.T) {
	tests := []struct {
		name string
		code string
		want bool
	}{
		{
			name: "standard log Printf",
			code: `log.Printf("info")`,
			want: true,
		},
		{
			name: "standard log Print",
			code: `log.Print("hello")`,
			want: true,
		},

		{
			name: "slog info via variable",
			code: `logger.Info("fail")`,
			want: true,
		},
		{
			name: "zap error via variable",
			code: `z.Error("critical")`,
			want: true,
		},
		{
			name: "zap sugar warn",
			code: `sugar.Warn("careful")`,
			want: true,
		},

		{
			name: "zap global L info",
			code: `zap.L().Info("global")`,
			want: true,
		},
		{
			name: "zap global S error",
			code: `zap.S().Error("global sugar")`,
			want: true,
		},
		{
			name: "slog default info",
			code: `slog.Default().Info("default slog")`,
			want: true,
		},

		{
			name: "zap from factory function",
			code: `getLog().Debug("factory")`,
			want: true,
		},

		{
			name: "unknown package",
			code: `fmt.Printf("hello")`,
			want: false,
		},
		{
			name: "custom method on string",
			code: `"my string".Upper()`,
			want: false,
		},
		{
			name: "local function call without selector",
			code: `Info("local")`,
			want: false,
		},
		{
			name: "variable named logger but no method",
			code: `logger("direct call")`,
			want: false,
		},
		{
			name: "unrelated type with same method name",
			code: `myStorage.Info("getting data")`,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, _ := parser.ParseExpr(tt.code)
			call, ok := expr.(*ast.CallExpr)
			if !ok {
				t.Fatalf("test code is not a call expression: %s", tt.code)
			}

			if got := IsLogCall(pass, call); got != tt.want {
				t.Errorf("IsLogCall(%s) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
