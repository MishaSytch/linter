package baseAnalizer

import (
	"fmt"
	"github.com/MishaSytch/linter/pkg/config"
	"github.com/MishaSytch/linter/pkg/linter/model"
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
	if arg == nil {
		return
	}

	if tv, ok := a.Pass.GetTypeAndValue(arg); ok && tv.Value != nil {
		msg := strings.Trim(tv.Value.ExactString(), `"`+"`")
		if msg != "" {
			a.runChecks(arg, msg)
		}
		return
	}

	switch expr := arg.(type) {
	case *ast.BinaryExpr:
		if expr.Op == token.ADD {
			a.Analyze(expr.X)
			a.Analyze(expr.Y)
		}

	case *ast.Ident:
		if expr.Obj != nil && expr.Obj.Decl != nil {
			switch decl := expr.Obj.Decl.(type) {
			case *ast.AssignStmt:
				for _, rhs := range decl.Rhs {
					a.Analyze(rhs)
				}
			case *ast.ValueSpec:
				for _, val := range decl.Values {
					a.Analyze(val)
				}
			case *ast.Field:
				if tv, ok := a.Pass.GetTypeAndValue(expr); ok && tv.Value != nil {
					a.runChecks(expr, strings.Trim(tv.Value.ExactString(), `"`))
				}
			}
		}

	case *ast.CompositeLit:
		for _, elt := range expr.Elts {
			a.Analyze(elt)
		}

	case *ast.KeyValueExpr:
		if id, ok := expr.Key.(*ast.Ident); ok {
			a.checkSensitiveData(id, id.Name)
		}
		a.Analyze(expr.Value)

	case *ast.CallExpr:
		for _, callArg := range expr.Args {
			a.Analyze(callArg)
		}

	case *ast.UnaryExpr:
		if expr.Op == token.AND {
			a.Analyze(expr.X)
		}

	case *ast.SelectorExpr:
		a.checkSensitiveData(expr.Sel, expr.Sel.Name)
		a.Analyze(expr.X)
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
