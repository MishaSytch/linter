package util

import (
	"go/ast"
	"go/parser"
	"testing"
)

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
			name: "standard log",
			code: `log.Printf("info")`,
			want: true,
		},
		{
			name: "slog error",
			code: `slog.Error("fail")`,
			want: true,
		},
		{
			name: "custom logger from config",
			code: `zap.Info("zap log")`,
			want: true,
		},
		{
			name: "unknown logger",
			code: `unknown.Info("hello")`,
			want: false,
		},
		{
			name: "local function call",
			code: `Info("local")`,
			want: false,
		},
		{
			name: "local function call",
			code: `logger("local")`,
			want: false,
		},
		{
			name: "local function call",
			code: `Logger("local")`,
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

			if got := IsLogCall(call); got != tt.want {
				t.Errorf("IsLogCall() = %v, want %v", got, tt.want)
			}
		})
	}
}
