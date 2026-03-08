package baseAnalizer

import (
	"fmt"
	"github.com/MishaSytch/linter/pkg/config"
	"github.com/MishaSytch/linter/pkg/linter/model"
	"github.com/MishaSytch/linter/pkg/linter/util"
	"go/ast"
	"go/token"
	"golang.org/x/tools/go/analysis"
	"strings"
	"unicode"
)

// BaseAnalyzer базовый анализатор кода
type BaseAnalyzer struct {
	Pass   model.Reporter
	Config *config.Config
}

var _ model.FunctionalAnalyzer = (*BaseAnalyzer)(nil)

func (a *BaseAnalyzer) Analyze(arg ast.Expr) {
	if tv, ok := a.Pass.GetTypeAndValue(arg); ok && tv.Value != nil {
		msg := strings.Trim(tv.Value.ExactString(), `"`+"`")
		if msg != "" {
			a.runChecks(arg, msg)
		}
		return
	}

	if call, ok := arg.(*ast.CallExpr); ok {
		if util.IsFmtSprintf(call) && len(call.Args) > 0 {
			a.Analyze(call.Args[0])
			return
		}
	}

	if bin, ok := arg.(*ast.BinaryExpr); ok && bin.Op == token.ADD {
		a.Analyze(bin.X)
		a.Analyze(bin.Y)
	}
}

func (a *BaseAnalyzer) runChecks(arg ast.Expr, msg string) {
	checkers := []errorsChecker{
		a.checkFirstLetterCase,
		a.checkTextSyntax,
		a.checkSensitiveData,
	}

	for _, checker := range checkers {
		checker(arg, msg)
	}
}

// checkFirstLetterCase проверяет, что сообщение начинается со строчной буквы
func (a *BaseAnalyzer) checkFirstLetterCase(arg ast.Expr, msg string) {
	runes := []rune(msg)
	if len(runes) > 0 && unicode.IsUpper(runes[0]) {
		runes[0] = unicode.ToLower(runes[0])
		fixedMsg := string(runes)

		var reportMsg string
		if a.Config.Output.TestRun {
			reportMsg = model.MsgLowerCaseRules
		} else {
			reportMsg = fmt.Sprintf("%s \n\tsuggested: \t%s", model.MsgLowerCaseRules, fixedMsg)
		}

		a.Pass.Report(analysis.Diagnostic{
			Pos:     arg.Pos(),
			Message: reportMsg,
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Convert first letter to lowercase",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     arg.Pos(),
							End:     arg.End(),
							NewText: []byte("\"" + fixedMsg + "\""),
						},
					},
				},
			},
		})
	}
}

// checkTextSyntax проверяет отсутствие кириллицы и спецсимволов
func (a *BaseAnalyzer) checkTextSyntax(arg ast.Expr, msg string) {
	errors := make(textErrors)
	checkers := []textSyntaxChecker{
		a.checkOnlyEnglishSyntax,
		a.checkSpecialCharsSyntax,
	}

	for _, r := range msg {
		for _, check := range checkers {
			check(arg, errors, r)
		}

		if len(errors) == len(checkers) {
			break
		}
	}
}

// checkSensitiveData ищет утечки секретов в тексте сообщения
func (a *BaseAnalyzer) checkSensitiveData(arg ast.Expr, msg string) {
	fixedMsg := msg
	lowerMsg := strings.ToLower(msg)
	var allFoundIssues []string

	var newFoundIssues []string
	fixedMsg, newFoundIssues = containSensitiveData(lowerMsg, a.Config)
	allFoundIssues = append(allFoundIssues, newFoundIssues...)

	if len(allFoundIssues) > 0 {
		var reportMsg string
		if a.Config.Output.TestRun {
			reportMsg = fmt.Sprintf(model.MsgSensitiveData, strings.Join(allFoundIssues, ", "))
		} else {
			reportMsg = fmt.Sprintf("%s \n\tsuggested: \t%s", model.MsgSensitiveData, fixedMsg)
		}

		a.Pass.Report(analysis.Diagnostic{
			Pos:     arg.Pos(),
			Message: reportMsg,
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Mask sensitive data with asterisks",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     arg.Pos(),
							End:     arg.End(),
							NewText: []byte("\"" + fixedMsg + "\""),
						},
					},
				},
			},
		})
	}
}

// checkOnlyEnglishSyntax проверяет, является ли символ буквой и входит ли он в латинский алфавит
func (a *BaseAnalyzer) checkOnlyEnglishSyntax(arg ast.Expr, errors textErrors, r rune) {
	if !errors[model.MsgEnglish] && isNonEnglish(r) {
		a.Pass.Reportf(arg.Pos(), model.MsgEnglish)
		errors[model.MsgEnglish] = true
	}
}

// checkSpecialCharsSyntax проверяет наличие спецсимволов и не-ASCII символов
func (a *BaseAnalyzer) checkSpecialCharsSyntax(arg ast.Expr, errors textErrors, r rune) {
	if errors[model.MsgSpecialChars] {
		return
	}

	if r <= 127 {
		return
	}

	if !unicode.IsLetter(r) {
		var msg string
		if lit, ok := arg.(*ast.BasicLit); ok {
			msg = strings.Trim(lit.Value, "\"`")
		}
		fixedMsg := ""
		for _, runeVal := range msg {
			if !isSpecialChars(runeVal) {
				fixedMsg += string(runeVal)
			}
		}

		var reportMsg string
		if a.Config.Output.TestRun {
			reportMsg = model.MsgSpecialChars
		} else {
			reportMsg = fmt.Sprintf("%s  \n\tsuggested: \t%s", model.MsgSpecialChars, fixedMsg)
		}

		a.Pass.Report(analysis.Diagnostic{
			Pos:     arg.Pos(),
			Message: reportMsg,
			SuggestedFixes: []analysis.SuggestedFix{
				{
					Message: "Remove special characters",
					TextEdits: []analysis.TextEdit{
						{
							Pos:     arg.Pos(),
							End:     arg.End(),
							NewText: []byte("\"" + fixedMsg + "\""),
						},
					},
				},
			},
		})
		errors[model.MsgSpecialChars] = true
	}
}
