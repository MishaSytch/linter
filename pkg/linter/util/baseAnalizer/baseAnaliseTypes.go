package baseAnalizer

type analysisResult struct {
	originalMsg string
	fixedMsg    string
	errors      []string
}

type errorsChecker func(res *analysisResult)
type textErrors map[string]bool
type textSyntaxChecker func(res *analysisResult, r rune, errors textErrors) bool
