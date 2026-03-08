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

// BaseAggregateAnalyzer базовый анализатор кода c агрегацией ошибок
type BaseAggregateAnalyzer struct {
	Pass   model.Reporter
	Config *config.Config
}

var _ model.FunctionalAnalyzer = (*BaseAggregateAnalyzer)(nil)

func (a *BaseAggregateAnalyzer) Analyze(arg ast.Expr) {
	if tv, ok := a.Pass.GetTypeAndValue(arg); ok && tv.Value != nil {
		msg := strings.Trim(tv.Value.ExactString(), `"`+"`")
		if msg != "" {
			a.runChecks(arg, msg)
		}
		return
	}

	if call, ok := arg.(*ast.CallExpr); ok {
		if util.IsFmtSprintf(call) && len(call.Args) > 0 {
			for _, arg := range call.Args {
				a.Analyze(arg)
			}
			return
		}
	}

	if bin, ok := arg.(*ast.BinaryExpr); ok && bin.Op == token.ADD {
		a.Analyze(bin.X)
		a.Analyze(bin.Y)
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
			a.applySensitiveFix(&analysisResult{
				originalMsg: id.Name,
				fixedMsg:    id.Name,
			})
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
		a.applySensitiveFix(&analysisResult{
			originalMsg: expr.Sel.Name,
			fixedMsg:    expr.Sel.Name,
		})
		a.Analyze(expr.X)
	}
}

func (a *BaseAggregateAnalyzer) runChecks(arg ast.Expr, msg string) {
	res := &analysisResult{
		originalMsg: msg,
		fixedMsg:    msg,
	}

	checkers := []errorsCheckerAggregate{
		a.applyFirstLetterFix,
		a.applySyntaxFix,
		a.applySensitiveFix,
	}

	for _, checker := range checkers {
		checker(res)
	}

	if len(res.errors) > 0 {
		a.reportCombined(arg, res)
	}
}

func (a *BaseAggregateAnalyzer) applyFirstLetterFix(res *analysisResult) {
	runes := []rune(res.fixedMsg)
	if len(runes) > 0 && unicode.IsUpper(runes[0]) {
		runes[0] = unicode.ToLower(runes[0])
		res.fixedMsg = string(runes)
		res.errors = append(res.errors, "should start with a lowercase letter")
	}
}

func (a *BaseAggregateAnalyzer) applySyntaxFix(res *analysisResult) {
	runes := []rune(res.fixedMsg)
	var cleanedRunes []rune

	reportedErrors := make(textErrors)

	syntaxCheckers := []textSyntaxCheckerWithBool{
		a.checkOnlyEnglishSyntaxWithBool,
		a.checkSpecialCharsSyntaxWithBool,
	}

	for _, r := range runes {
		shouldDelete := false

		for _, check := range syntaxCheckers {
			if check(res, r, reportedErrors) {
				shouldDelete = true
			}
		}

		if shouldDelete {
			continue
		}

		cleanedRunes = append(cleanedRunes, r)
	}

	res.fixedMsg = string(cleanedRunes)
}

func (a *BaseAggregateAnalyzer) applySensitiveFix(res *analysisResult) {
	fixedMsg := res.fixedMsg
	lowerMsg := strings.ToLower(fixedMsg)
	var foundIssues []string

	var newFoundIssues []string
	fixedMsg, newFoundIssues = containSensitiveData(lowerMsg, a.Config)
	foundIssues = append(foundIssues, newFoundIssues...)

	if len(foundIssues) > 0 {
		res.fixedMsg = fixedMsg

		msg := fmt.Sprintf("log message contains potentially sensitive data: %s",
			strings.Join(foundIssues, ", "))

		res.errors = append(res.errors, msg)
	}
}

func (a *BaseAggregateAnalyzer) reportCombined(arg ast.Expr, res *analysisResult) {
	if a.Config.Output.TestRun {
		for _, errStr := range res.errors {
			a.Pass.Report(analysis.Diagnostic{
				Pos:     arg.Pos(),
				Message: errStr,
				SuggestedFixes: []analysis.SuggestedFix{
					{
						Message: "Fix",
						TextEdits: []analysis.TextEdit{
							{
								Pos:     arg.Pos(),
								End:     arg.End(),
								NewText: []byte("\"" + res.fixedMsg + "\""),
							},
						},
					},
				},
			})
		}
		return
	}

	var fullMessage string
	errorList := "log message issues:\n  - " + strings.Join(res.errors, "\n  - ")

	if a.Config.Output.ShowInConsole {
		fullMessage = fmt.Sprintf("%s\n\tsuggested:\t\"%s\"", errorList, res.fixedMsg)
	} else {
		fullMessage = errorList
	}

	diagnostic := analysis.Diagnostic{
		Pos:     arg.Pos(),
		Message: fullMessage,
	}

	if a.Config.Output.ShowSuggestions {
		diagnostic.SuggestedFixes = []analysis.SuggestedFix{
			{
				Message: "Apply all fixes",
				TextEdits: []analysis.TextEdit{
					{
						Pos:     arg.Pos(),
						End:     arg.End(),
						NewText: []byte("\"" + res.fixedMsg + "\""),
					},
				},
			},
		}
	}

	a.Pass.Report(diagnostic)
}

// checkOnlyEnglishSyntaxWithBool проверяет, является ли символ буквой и входит ли он в латинский алфавит
// и возвращает требование к исправлению
func (a *BaseAggregateAnalyzer) checkOnlyEnglishSyntaxWithBool(res *analysisResult, r rune, reported textErrors) bool {
	isNonLatinLetter := isNonEnglish(r)

	if isNonLatinLetter {
		if !reported[model.MsgEnglish] {
			res.errors = append(res.errors, model.MsgEnglish)
			reported[model.MsgEnglish] = true
		}
		return true
	}
	return false
}

// checkSpecialCharsSyntaxWithBool проверяет наличие спецсимволов и не-ASCII символов
// и возвращает требование к исправлению
func (a *BaseAggregateAnalyzer) checkSpecialCharsSyntaxWithBool(res *analysisResult, r rune, reported textErrors) bool {
	isSpecial := isSpecialChars(r)

	if isSpecial {
		if !reported[model.MsgSpecialChars] {
			res.errors = append(res.errors, model.MsgSpecialChars)
			reported[model.MsgSpecialChars] = true
		}
		return true
	}
	return false
}
