package baseAnalizer

import (
	"github.com/MishaSytch/linter/pkg/config"
	"regexp"
	"strings"
	"unicode"
)

func isSpecialChars(r rune) bool {
	return r > 127 && !unicode.IsLetter(r)
}

func isNonEnglish(r rune) bool {
	return unicode.IsLetter(r) && (r < 'A' || (r > 'Z' && r < 'a') || r > 'z')
}

func containSensitiveData(msg string, cfg *config.Config) (fixedMsg string, allFoundIssues []string) {
	for _, kw := range cfg.SensitiveRules.SensitiveWords {
		lowerKw := strings.ToLower(kw)
		if strings.Contains(msg, lowerKw) {
			allFoundIssues = append(allFoundIssues, kw)
			mask := strings.Repeat("*", len(kw))
			msg = strings.ReplaceAll(msg, kw, mask)
		}
	}

	for _, p := range cfg.SensitiveRules.Patterns {
		re, err := regexp.Compile(p.Regex)
		if err != nil {
			continue
		}

		if re.MatchString(msg) {
			allFoundIssues = append(allFoundIssues, p.Name)
			msg = re.ReplaceAllString(msg, "********")
		}
	}

	return msg, allFoundIssues
}
