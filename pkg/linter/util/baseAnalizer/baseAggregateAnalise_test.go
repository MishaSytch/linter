package baseAnalizer

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"testing"

	"github.com/MishaSytch/linter/pkg/config"
	"github.com/MishaSytch/linter/pkg/linter/model"
	"golang.org/x/tools/go/analysis"
)

// MockReporterAggregate мок для Reporter
type MockReporterAggregate struct {
	reports   *[]string
	typesInfo *types.Info
}

func (m *MockReporterAggregate) Reportf(pos token.Pos, format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	*m.reports = append(*m.reports, msg)
}

func (m *MockReporterAggregate) Report(d analysis.Diagnostic) {
	*m.reports = append(*m.reports, d.Message)
}

func (m *MockReporterAggregate) TypesInfo() *types.Info {
	return m.typesInfo
}

func (m *MockReporterAggregate) GetTypeAndValue(expr ast.Expr) (types.TypeAndValue, bool) {
	if m.typesInfo == nil {
		return types.TypeAndValue{}, false
	}
	tv, ok := m.typesInfo.Types[expr]
	return tv, ok
}

func (m *MockReporterAggregate) ObjectOf(id *ast.Ident) types.Object { return nil }
func (m *MockReporterAggregate) TypeOf(expr ast.Expr) types.Type     { return nil }
func (m *MockReporterAggregate) GetSelection(sel *ast.SelectorExpr) (*types.Selection, bool) {
	return nil, false
}
func (m *MockReporterAggregate) GetObject(id *ast.Ident) (types.Object, bool) { return nil, false }

var _ model.Reporter = (*MockReporterAggregate)(nil)

func setupBaseTestAggregate(cfg *config.Config) (*BaseAggregateAnalyzer, *[]string) {
	reports := &[]string{}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	mock := &MockReporterAggregate{
		reports:   reports,
		typesInfo: info,
	}
	return &BaseAggregateAnalyzer{
		Pass:   mock,
		Config: cfg,
	}, reports
}

func TestBaseAnalyzerAggregate(t *testing.T) {
	testCfg := &config.Config{
		SensitiveRules: config.SensitiveRules{
			SensitiveWords: []string{"secret"},
		},
		Output: config.OutputConfig{
			TestRun: true,
		},
	}
	analyzer, reports := setupBaseTestAggregate(testCfg)

	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		{"Valid", "good message", false},
		{"Russian", "сообщение", true},
		{"Secret", "my secret", true},
		{"Uppercase", "Bad message", true},
		{"All", "Bad message good message сообщение my secret", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*reports = nil
			expr := &ast.BasicLit{Value: `"` + tt.msg + `"`}

			analyzer.Pass.TypesInfo().Types[expr] = types.TypeAndValue{
				Type:  types.Typ[types.String],
				Value: constant.MakeString(tt.msg),
			}

			analyzer.Analyze(expr)

			if (len(*reports) > 0) != tt.wantErr {
				t.Errorf("Analyze() '%s' errors: %v", tt.msg, *reports)
			}
		})
	}
}
