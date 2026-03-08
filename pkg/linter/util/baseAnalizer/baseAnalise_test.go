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

// MockReporter теперь умеет возвращать типы
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

func (m *MockReporter) TypesInfo() *types.Info { return m.typesInfo }

func (m *MockReporter) GetTypeAndValue(expr ast.Expr) (types.TypeAndValue, bool) {
	if m.typesInfo == nil {
		return types.TypeAndValue{}, false
	}
	tv, ok := m.typesInfo.Types[expr]
	return tv, ok
}

func (m *MockReporter) ObjectOf(id *ast.Ident) types.Object { return nil }
func (m *MockReporter) TypeOf(expr ast.Expr) types.Type     { return nil }
func (m *MockReporter) GetSelection(sel *ast.SelectorExpr) (*types.Selection, bool) {
	return nil, false
}
func (m *MockReporter) GetObject(id *ast.Ident) (types.Object, bool) { return nil, false }

var _ model.Reporter = (*MockReporter)(nil)

func setupBaseTest(cfg *config.Config) (*BaseAnalyzer, *[]string) {
	reports := &[]string{}
	info := &types.Info{
		Types: make(map[ast.Expr]types.TypeAndValue),
	}
	mock := &MockReporter{
		reports:   reports,
		typesInfo: info,
	}
	return &BaseAnalyzer{
		Pass:   mock,
		Config: cfg,
	}, reports
}

func TestBaseAnalyzer_Analyze(t *testing.T) {
	testCfg := &config.Config{
		SensitiveRules: config.SensitiveRules{
			SensitiveWords: []string{"secret"},
		},
		Output: config.OutputConfig{TestRun: true},
	}
	analyzer, reports := setupBaseTest(testCfg)

	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		{"Valid", "good message", false},
		{"Russian", "сообщение", true},
		{"Secret", "my secret", true},
		{"Uppercase", "Bad message", true},
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

func Test_checkTextSyntax(t *testing.T) {
	analyzer, reports := setupBaseTest(&config.Config{})
	tests := []struct {
		name            string
		msg             string
		wantErrorsCount int
	}{
		{"Clean English", "valid message", 0},
		{"Cyrillic and Emoji", "ошибка 🫡", 2},
		{"Multiple emojis", "🫡🚀", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*reports = nil
			analyzer.checkTextSyntax(&ast.BasicLit{}, tt.msg)

			if len(*reports) != tt.wantErrorsCount {
				t.Errorf("got %d errors, want %d. Errors: %v", len(*reports), tt.wantErrorsCount, *reports)
			}
		})
	}
}

func Test_checkSensitiveData(t *testing.T) {
	testCfg := &config.Config{
		SensitiveRules: config.SensitiveRules{
			SensitiveWords: []string{"secret"},
			Patterns: []config.SensitivePattern{
				{Name: "API_KEY", Regex: `(?i)api_key_[a-z0-9]{10}`},
			},
		},
		Output: config.OutputConfig{TestRun: true},
	}
	analyzer, reports := setupBaseTest(testCfg)

	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		{"Leak keyword", "my SECRET_TOKEN", true},
		{"Leak pattern", "api_key_1234567890", true},
		{"Clear", "hello world", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*reports = nil
			analyzer.checkSensitiveData(&ast.BasicLit{}, tt.msg)

			if (len(*reports) > 0) != tt.wantErr {
				t.Errorf("got error = %v, want %v", len(*reports) > 0, tt.wantErr)
			}
		})
	}
}
