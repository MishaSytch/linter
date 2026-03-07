package baseAnalizer

import (
	"go/ast"
	"linter/pkg/config"
	"linter/pkg/linter/model"
	"linter/pkg/linter/util"
	"strings"
	"unicode"
)

var cfg *config.Config

// BaseAnalyzer базовый анализатор кода
func BaseAnalyzer(pass Reporter, arg ast.Expr, config *config.Config) {
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
func checkFirstLetterCase(pass Reporter, arg ast.Expr, msg string) {
	runes := []rune(msg)
	if len(runes) > 0 && unicode.IsUpper(runes[0]) {
		pass.Reportf(arg.Pos(), model.MsgLowerCaseRules)
	}
}

// checkTextSyntax проверяет отсутствие кириллицы и спецсимволов
func checkTextSyntax(pass Reporter, arg ast.Expr, msg string) {
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
func checkSensitiveData(pass Reporter, arg ast.Expr, msg string) {
	lowerMsg := strings.ToLower(msg)
	for _, kw := range cfg.SensitiveWords {
		if strings.Contains(lowerMsg, kw) {
			pass.Reportf(arg.Pos(), model.MsgSensitiveData, kw)
			break
		}
	}
}

// checkOnlyEnglishSyntax проверяет, является ли символ буквой и входит ли он в латинский алфавит
func checkOnlyEnglishSyntax(pass Reporter, arg ast.Expr, errors textErrors, r rune) {
	if !errors[model.MsgEnglish] && unicode.IsLetter(r) && (r < 'A' || (r > 'Z' && r < 'a') || r > 'z') {
		pass.Reportf(arg.Pos(), model.MsgEnglish)
		errors[model.MsgEnglish] = true
	}
}

// checkSpecialCharsSyntax проверяет наличие спецсимволов и не-ASCII символов
func checkSpecialCharsSyntax(pass Reporter, arg ast.Expr, errors textErrors, r rune) {
	if errors[model.MsgSpecialChars] {
		return
	}
	if r <= 127 {
		return
	}

	if !unicode.IsLetter(r) {
		pass.Reportf(arg.Pos(), model.MsgSpecialChars)
		errors[model.MsgSpecialChars] = true
	}
}
