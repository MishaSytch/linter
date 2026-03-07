package baseAnalizer

import (
	"fmt"
	"github.com/MishaSytch/linter/pkg/config"
	"github.com/MishaSytch/linter/pkg/linter/model"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"testing"
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

func setupBaseTestAggregate() (model.Reporter, *[]string) {
	reports := &[]string{}
	mock := &MockReporterAggregate{
		reports: reports,
		typesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
		},
	}
	return mock, reports
}

func TestBaseAnalyzerAggregate(t *testing.T) {
	mock, reports := setupBaseTestAggregate()
	m := mock.(*MockReporterAggregate)
	testCfg := &config.Config{
		SensitiveRules: config.SensitiveRules{
			SensitiveWords: []string{"secret"},
			Patterns:       []config.SensitivePattern{},
		},
		Output: config.OutputConfig{
			ErrorsAggregate: true,
			TestRun:         true,
		},
	}
	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		{
			"Valid",
			"good message",
			false,
		},
		{
			"Russian",
			"сообщение",
			true,
		},
		{
			"Secret",
			"my secret",
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*reports = nil
			expr := &ast.BasicLit{Value: `"` + tt.msg + `"`}

			m.typesInfo.Types[expr] = types.TypeAndValue{
				Value: constant.MakeString(tt.msg),
			}

			BaseAnalyzer(m, expr, testCfg)

			if (len(*reports) > 0) != tt.wantErr {
				t.Errorf("BaseAnalyzer() '%s' errors: %v", tt.msg, *reports)
			}
		})
	}
}

func Test_applySyntaxFix(t *testing.T) {
	tests := []struct {
		name            string
		msg             string
		wantErrorsCount int
		wantFixedMsg    string
	}{
		{
			name:            "Clean English text",
			msg:             "valid message",
			wantErrorsCount: 0,
			wantFixedMsg:    "valid message",
		},
		{
			name:            "Only Cyrillic",
			msg:             "ошибка",
			wantErrorsCount: 1,
			wantFixedMsg:    "",
		},
		{
			name:            "Cyrillic and Emoji",
			msg:             "ошибка 🫡",
			wantErrorsCount: 2,
			wantFixedMsg:    " ",
		},
		{
			name:            "Multiple emojis",
			msg:             "🫡🚀",
			wantErrorsCount: 1,
			wantFixedMsg:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &analysisResult{fixedMsg: tt.msg}
			applySyntaxFix(res)

			if len(res.errors) != tt.wantErrorsCount {
				t.Errorf("applySyntaxFix() '%s' got %d errors, want %d", tt.msg, len(res.errors), tt.wantErrorsCount)
			}
			if res.fixedMsg != tt.wantFixedMsg {
				t.Errorf("applySyntaxFix() '%s' fixedMsg = '%s', want '%s'", tt.msg, res.fixedMsg, tt.wantFixedMsg)
			}
		})
	}
}

func Test_checkOnlyEnglishSyntaxWithBool(t *testing.T) {
	tests := []struct {
		name         string
		r            rune
		shouldDelete bool
	}{
		{
			"Latin",
			'a',
			false,
		},
		{
			"Cyrillic",
			'а',
			true,
		},
		{
			"Number",
			'1',
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &analysisResult{}
			reported := make(textErrors)

			gotDelete := checkOnlyEnglishSyntaxWithBool(res, tt.r, reported)

			if gotDelete != tt.shouldDelete {
				t.Errorf("checkOnlyEnglishSyntax() rune %c, gotDelete = %v, want %v", tt.r, gotDelete, tt.shouldDelete)
			}
		})
	}
}
func Test_applySensitiveFix(t *testing.T) {
	cfg = &config.Config{
		SensitiveRules: config.SensitiveRules{
			SensitiveWords: []string{"secret"},
			Patterns: []config.SensitivePattern{
				{
					Name:  "IPv4 Address",
					Regex: `(?:\d{1,3}\.){3}\d{1,3}`,
				},
			},
		},
	}

	tests := []struct {
		name         string
		msg          string
		wantError    bool
		wantFixedMsg string
	}{
		{
			"Leak secret (Keyword)",
			"my secret is 123",
			true,
			"my ****** is 123",
		},
		{
			"Leak IP (Pattern)",
			"connecting to 192.168.1.1",
			true,
			"connecting to ********",
		},
		{
			"Mixed leak (Keyword and Pattern)",
			"secret ip is 127.0.0.1",
			true,
			"****** ip is ********",
		},
		{
			"Clear text",
			"hello world",
			false,
			"hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res := &analysisResult{
				fixedMsg: tt.msg,
				errors:   []string{},
			}

			applySensitiveFix(res)

			hasError := len(res.errors) > 0
			if hasError != tt.wantError {
				t.Errorf("%s: applySensitiveFix() error = %v, want %v. Errors: %v",
					tt.name, hasError, tt.wantError, res.errors)
			}

			if res.fixedMsg != tt.wantFixedMsg {
				t.Errorf("%s: fixedMsg got %s, want %s",
					tt.name, res.fixedMsg, tt.wantFixedMsg)
			}
		})
	}
}
