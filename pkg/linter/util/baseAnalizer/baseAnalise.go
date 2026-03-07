package baseAnalizer

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"linter/pkg/config"
	"linter/pkg/linter/model"
	"linter/pkg/linter/util"
	"strings"
	"unicode"
)

var cfg *config.Config

// BaseAnalyzer теперь собирает все правки воедино
func BaseAnalyzer(pass model.Reporter, arg ast.Expr, config *config.Config) {
	cfg = config
	var msg string

	if tv, ok := pass.TypesInfo().Types[arg]; ok && tv.Value != nil {
		msg = strings.Trim(tv.Value.ExactString(), `"`+"`")
	} else if subCall, ok := arg.(*ast.CallExpr); ok && util.IsFmtSprintf(subCall) && len(subCall.Args) > 0 {
		BaseAnalyzer(pass, subCall.Args[0], config)
		return
	}

	if msg == "" {
		return
	}

	res := &analysisResult{
		originalMsg: msg,
		fixedMsg:    msg,
	}

	checkers := []errorsChecker{
		applyFirstLetterFix,
		applySyntaxFix,
		applySensitiveFix,
	}

	for _, checker := range checkers {
		checker(res)
	}

	if len(res.errors) > 0 {
		reportCombined(pass, arg, res)
	}
}

func applyFirstLetterFix(res *analysisResult) {
	runes := []rune(res.fixedMsg)
	if len(runes) > 0 && unicode.IsUpper(runes[0]) {
		runes[0] = unicode.ToLower(runes[0])
		res.fixedMsg = string(runes)
		res.errors = append(res.errors, "should start with a lowercase letter")
	}
}

func applySyntaxFix(res *analysisResult) {
	runes := []rune(res.fixedMsg)
	var cleanedRunes []rune

	reportedErrors := make(textErrors)

	syntaxCheckers := []textSyntaxChecker{
		checkOnlyEnglishSyntax,
		checkSpecialCharsSyntax,
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

func applySensitiveFix(res *analysisResult) {
	tempMsg := res.fixedMsg
	lowerMsg := strings.ToLower(tempMsg)
	found := false

	for _, kw := range cfg.SensitiveWords {
		if strings.Contains(lowerMsg, kw) {
			mask := strings.Repeat("*", len(kw))
			tempMsg = strings.ReplaceAll(tempMsg, kw, mask)
			found = true
		}
	}

	if found {
		res.fixedMsg = tempMsg
		res.errors = append(res.errors, "contains potentially sensitive data")
	}
}

func reportCombined(pass model.Reporter, arg ast.Expr, res *analysisResult) {
	var fullMessage string

	errorList := "log message issues:\n  - " + strings.Join(res.errors, "\n  - ")

	if cfg.Output.ShowInConsole {
		fullMessage = fmt.Sprintf("%s\n\tsuggested:\t\"%s\"", errorList, res.fixedMsg)
	} else {
		fullMessage = errorList
	}

	diagnostic := analysis.Diagnostic{
		Pos:     arg.Pos(),
		Message: fullMessage,
	}

	if cfg.Output.ShowSuggestions {
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

	pass.Report(diagnostic)
}

func checkOnlyEnglishSyntax(res *analysisResult, r rune, reported textErrors) bool {
	isNonLatinLetter := unicode.IsLetter(r) && (r < 'A' || (r > 'Z' && r < 'a') || r > 'z')

	if isNonLatinLetter {
		if !reported[model.MsgEnglish] {
			res.errors = append(res.errors, model.MsgEnglish)
			reported[model.MsgEnglish] = true
		}
		return true
	}
	return false
}

func checkSpecialCharsSyntax(res *analysisResult, r rune, reported textErrors) bool {
	isSpecial := r > 127 && !unicode.IsLetter(r)

	if isSpecial {
		if !reported[model.MsgSpecialChars] {
			res.errors = append(res.errors, model.MsgSpecialChars)
			reported[model.MsgSpecialChars] = true
		}
		return true
	}
	return false
}
