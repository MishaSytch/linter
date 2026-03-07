package baseAnalizer

import (
	"fmt"
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"linter/pkg/config"
	"linter/pkg/linter/model"
	"linter/pkg/linter/util"
	"regexp"
	"strings"
	"unicode"
)

// BaseAnalyzer базовый анализатор кода
func BaseAnalyzer(pass model.Reporter, arg ast.Expr, config *config.Config) {
	cfg = config

	var msg string

	if tv, ok := pass.TypesInfo().Types[arg]; ok && tv.Value != nil {
		msg = strings.Trim(tv.Value.ExactString(), `"`+"`")
	} else {
		if subCall, ok := arg.(*ast.CallExpr); ok && util.IsFmtSprintf(subCall) && len(subCall.Args) > 0 {
			BaseAnalyzer(pass, subCall.Args[0], config)
		}
		return
	}

	if msg == "" {
		return
	}

	checkers := []errorsChecker{
		checkFirstLetterCase,
		checkTextSyntax,
		checkSensitiveData,
	}

	for _, checker := range checkers {
		checker(pass, arg, msg)
	}
}

// checkFirstLetterCase проверяет, что сообщение начинается со строчной буквы
func checkFirstLetterCase(pass model.Reporter, arg ast.Expr, msg string) {
	runes := []rune(msg)
	if len(runes) > 0 && unicode.IsUpper(runes[0]) {
		runes[0] = unicode.ToLower(runes[0])
		fixedMsg := string(runes)

		var reportMsg string
		if cfg.Output.TestRun {
			reportMsg = model.MsgLowerCaseRules
		} else {
			reportMsg = fmt.Sprintf("log message should start with a lowercase letter \n\tsuggested: \t%s", fixedMsg)
		}

		pass.Report(analysis.Diagnostic{
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
func checkTextSyntax(pass model.Reporter, arg ast.Expr, msg string) {
	errors := make(textErrors)
	checkers := []textSyntaxChecker{
		checkOnlyEnglishSyntax,
		checkSpecialCharsSyntax,
	}

	for _, r := range msg {
		for _, check := range checkers {
			check(pass, arg, errors, r)
		}

		if len(errors) == len(checkers) {
			break
		}
	}
}

// checkSensitiveData ищет утечки секретов в тексте сообщения
func checkSensitiveData(pass model.Reporter, arg ast.Expr, msg string) {
	fixedMsg := msg
	lowerMsg := strings.ToLower(msg)
	var allFoundIssues []string

	for _, kw := range cfg.SensitiveRules.SensitiveWords {
		lowerKw := strings.ToLower(kw)
		if strings.Contains(lowerMsg, lowerKw) {
			allFoundIssues = append(allFoundIssues, kw)
			mask := strings.Repeat("*", len(kw))
			fixedMsg = strings.ReplaceAll(fixedMsg, kw, mask)
		}
	}

	for _, p := range cfg.SensitiveRules.Patterns {
		re, err := regexp.Compile(p.Regex)
		if err != nil {
			continue
		}

		if re.MatchString(fixedMsg) {
			allFoundIssues = append(allFoundIssues, p.Name)
			fixedMsg = re.ReplaceAllString(fixedMsg, "********")
		}
	}

	if len(allFoundIssues) > 0 {
		var reportMsg string
		if cfg.Output.TestRun {
			reportMsg = fmt.Sprintf(model.MsgSensitiveData, strings.Join(allFoundIssues, ", "))
		} else {
			reportMsg = fmt.Sprintf("log message contains potentially sensitive data \n\tsuggested: \t%s", fixedMsg)
		}

		pass.Report(analysis.Diagnostic{
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
func checkOnlyEnglishSyntax(pass model.Reporter, arg ast.Expr, errors textErrors, r rune) {
	if !errors[model.MsgEnglish] && unicode.IsLetter(r) && (r < 'A' || (r > 'Z' && r < 'a') || r > 'z') {
		pass.Reportf(arg.Pos(), model.MsgEnglish)
		errors[model.MsgEnglish] = true
	}
}

// checkSpecialCharsSyntax проверяет наличие спецсимволов и не-ASCII символов
func checkSpecialCharsSyntax(pass model.Reporter, arg ast.Expr, errors textErrors, r rune) {
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
			if runeVal <= 127 || unicode.IsLetter(runeVal) {
				fixedMsg += string(runeVal)
			}
		}

		var reportMsg string
		if cfg.Output.TestRun {
			reportMsg = model.MsgSpecialChars
		} else {
			reportMsg = fmt.Sprintf("%s  \n\tsuggested: \t%s", model.MsgSpecialChars, fixedMsg)
		}

		pass.Report(analysis.Diagnostic{
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
