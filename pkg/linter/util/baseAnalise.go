package util

import (
	"go/ast"
	"golang.org/x/tools/go/analysis"
	"linter/pkg/linter/model"
	"strings"
	"unicode"
)

type textErrors map[string]bool

type textSyntaxChecker func(pass *analysis.Pass, arg ast.Expr, errors textErrors, r rune)

// BaseAnalyzer базовый анализатор кода
func BaseAnalyzer(pass *analysis.Pass, arg ast.Expr) {
	var msg string

	if tv, ok := pass.TypesInfo.Types[arg]; ok && tv.Value != nil {
		msg = strings.Trim(tv.Value.ExactString(), `"`+"`")
	} else {
		if subCall, ok := arg.(*ast.CallExpr); ok && IsFmtSprintf(subCall) && len(subCall.Args) > 0 {
			BaseAnalyzer(pass, subCall.Args[0])
		}
		return
	}

	if msg == "" {
		return
	}

	checkFirstLetterCase(pass, arg, msg)
	checkTextSyntax(pass, arg, msg)
	checkSensitiveData(pass, arg, msg)
}

// checkFirstLetterCase проверяет, что сообщение начинается со строчной буквы
func checkFirstLetterCase(pass *analysis.Pass, arg ast.Expr, msg string) {
	runes := []rune(msg)
	if len(runes) > 0 && unicode.IsUpper(runes[0]) {
		pass.Reportf(arg.Pos(), model.MsgLowerCaseRules)
	}
}

// checkTextSyntax проверяет отсутствие кириллицы и спецсимволов
func checkTextSyntax(pass *analysis.Pass, arg ast.Expr, msg string) {
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

// checkOnlyEnglishSyntax проверяет, является ли символ буквой и входит ли он в латинский алфавит.
func checkOnlyEnglishSyntax(pass *analysis.Pass, arg ast.Expr, errors textErrors, r rune) {
	if !errors[model.MsgEnglish] && unicode.IsLetter(r) && (r < 'A' || (r > 'Z' && r < 'a') || r > 'z') {
		pass.Reportf(arg.Pos(), model.MsgEnglish)
		errors[model.MsgEnglish] = true
	}
}

// checkSpecialCharsSyntax проверяет наличие спецсимволов и не-ASCII графики.
func checkSpecialCharsSyntax(pass *analysis.Pass, arg ast.Expr, errors textErrors, r rune) {
	if !errors[model.MsgSpecialChars] && (unicode.IsSymbol(r) || unicode.IsGraphic(r)) && r > 127 {
		pass.Reportf(arg.Pos(), model.MsgSpecialChars)
		errors[model.MsgSpecialChars] = true
	}
}

// checkSensitiveData ищет утечки секретов в тексте сообщения
func checkSensitiveData(pass *analysis.Pass, arg ast.Expr, msg string) {
	lowerMsg := strings.ToLower(msg)
	for _, kw := range model.SensitiveKeywords {
		if strings.Contains(lowerMsg, kw) {
			pass.Reportf(arg.Pos(), model.MsgSensitiveData, kw)
			break
		}
	}
}
