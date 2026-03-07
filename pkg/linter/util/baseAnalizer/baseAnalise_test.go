package baseAnalizer

import (
	"fmt"
	"go/ast"
	"go/constant"
	"go/token"
	"go/types"
	"golang.org/x/tools/go/analysis"
	"linter/pkg/config"
	"linter/pkg/linter/model"
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

func setupBaseTest() (model.Reporter, *[]string) {
	reports := &[]string{}
	mock := &MockReporter{
		reports: reports,
		typesInfo: &types.Info{
			Types: make(map[ast.Expr]types.TypeAndValue),
		},
	}
	return mock, reports
}

func TestBaseAnalyzer(t *testing.T) {
	mock, reports := setupBaseTest()
	m := mock.(*MockReporter)
	testCfg := &config.Config{SensitiveWords: []string{"secret"}}
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

func Test_checkTextSyntax(t *testing.T) {
	pass, reports := setupBaseTest()
	tests := []struct {
		name            string
		msg             string
		wantErrorsCount int
	}{
		{
			name:            "Clean English text",
			msg:             "valid message",
			wantErrorsCount: 0,
		},
		{
			name:            "Only Cyrillic",
			msg:             "ошибка",
			wantErrorsCount: 1,
		},
		{
			name:            "Cyrillic and Emoji",
			msg:             "ошибка 🫡",
			wantErrorsCount: 2,
		},
		{
			name:            "Multiple emojis",
			msg:             "🫡🚀",
			wantErrorsCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*reports = nil
			checkTextSyntax(pass, &ast.BasicLit{}, tt.msg)

			if len(*reports) != tt.wantErrorsCount {
				t.Errorf("checkTextSyntax() '%s' got %d errors, want %d. Errors: %v",
					tt.msg, len(*reports), tt.wantErrorsCount, *reports)
			}
		})
	}
}

func Test_checkFirstLetterCase(t *testing.T) {
	pass, reports := setupBaseTest()
	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		{
			"Valid lowercase",
			"correct message",
			false,
		},
		{
			"Invalid uppercase",
			"Bad message",
			true,
		},
		{
			"Non-letter start",
			"123 message",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*reports = nil
			checkFirstLetterCase(pass, &ast.BasicLit{}, tt.msg)

			hasError := len(*reports) > 0
			if hasError != tt.wantErr {
				t.Errorf("checkFirstLetterCase() error = %v, wantErr %v", hasError, tt.wantErr)
			}
		})
	}
}

func Test_checkOnlyEnglishSyntax(t *testing.T) {
	pass, reports := setupBaseTest()
	tests := []struct {
		name    string
		r       rune
		wantErr bool
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
			*reports = nil
			errors := make(textErrors)
			checkOnlyEnglishSyntax(pass, &ast.BasicLit{}, errors, tt.r)

			hasError := len(*reports) > 0
			if hasError != tt.wantErr {
				t.Errorf("checkOnlyEnglishSyntax() rune %c, error = %v", tt.r, hasError)
			}
		})
	}
}

func Test_checkSpecialCharsSyntax(t *testing.T) {
	pass, reports := setupBaseTest()

	tests := []struct {
		name    string
		r       rune
		wantErr bool
	}{
		{
			"Standard ASCII",
			'!',
			false,
		},
		{
			"Emoji",
			'🫡',
			true,
		},
		{
			"Math symbol",
			'∑',
			true,
		},
		{
			"Spanish symbol",
			'ñ',
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*reports = nil
			errors := make(textErrors)
			checkSpecialCharsSyntax(pass, &ast.BasicLit{}, errors, tt.r)

			hasError := len(*reports) > 0
			if hasError != tt.wantErr {
				t.Errorf("checkSpecialCharsSyntax() rune %c, error = %v", tt.r, hasError)
			}
		})
	}
}

func Test_checkSensitiveData(t *testing.T) {
	pass, reports := setupBaseTest()
	cfg = &config.Config{SensitiveWords: []string{"password", "token"}}
	tests := []struct {
		name    string
		msg     string
		wantErr bool
	}{
		{
			"Clear text",
			"hello world",
			false,
		},
		{
			"Leak password",
			"my password is 123",
			true,
		},
		{
			"Leak token",
			"SECRET_TOKEN_HERE",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			*reports = nil
			checkSensitiveData(pass, &ast.BasicLit{}, tt.msg)

			if (len(*reports) > 0) != tt.wantErr {
				t.Errorf("checkSensitiveData() '%s' error = %v", tt.msg, len(*reports) > 0)
			}
		})
	}
}
