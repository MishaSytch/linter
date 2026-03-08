package util

import (
	"fmt"
	"github.com/MishaSytch/linter/pkg/linter/model"
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

func (m *MockReporter) GetTypeAndValue(expr ast.Expr) (types.TypeAndValue, bool) {
	if m.typesInfo == nil {
		return types.TypeAndValue{}, false
	}
	tv, ok := m.typesInfo.Types[expr]
	return tv, ok
}

func (m *MockReporter) TypeOf(expr ast.Expr) types.Type {
	if m.typesInfo == nil {
		return nil
	}
	// Пытаемся достать тип из стандартной мапы TypesInfo
	if tv, ok := m.typesInfo.Types[expr]; ok {
		return tv.Type
	}
	return nil
}

func (m *MockReporter) GetSelection(sel *ast.SelectorExpr) (*types.Selection, bool) {
	if m.typesInfo == nil {
		return nil, false
	}
	s, ok := m.typesInfo.Selections[sel]
	return s, ok
}

func (m *MockReporter) GetObject(id *ast.Ident) (types.Object, bool) {
	if m.typesInfo == nil {
		return nil, false
	}
	obj, ok := m.typesInfo.Uses[id]
	if !ok {
		obj, ok = m.typesInfo.Defs[id]
	}
	return obj, ok
}

func (m *MockReporter) ObjectOf(id *ast.Ident) types.Object { return nil }

var _ model.Reporter = (*MockReporter)(nil)

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
	model.LoggerList = []string{"log", "zap", "slog", "logger", "z", "sugar"}

	tests := []struct {
		name     string
		code     string
		want     bool
		mockType string
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
			reports := &[]string{}
			pass := &MockReporter{
				reports: reports,
				typesInfo: &types.Info{
					Types: make(map[ast.Expr]types.TypeAndValue),
				},
			}

			expr, _ := parser.ParseExpr(tt.code)
			call, ok := expr.(*ast.CallExpr)
			if !ok {
				t.Fatalf("test code is not a call expression: %s", tt.code)
			}

			if tt.mockType != "" {
				if sel, ok := call.Fun.(*ast.SelectorExpr); ok {
					pass.typesInfo.Types[sel.X] = types.TypeAndValue{
						Type: types.NewNamed(
							types.NewTypeName(token.NoPos, nil, tt.mockType, nil),
							nil, nil,
						),
					}
				}
			}

			if got := IsLogCall(pass, call); got != tt.want {
				t.Errorf("IsLogCall(%s) = %v, want %v", tt.code, got, tt.want)
			}
		})
	}
}
