package patterns

import (
	"fmt"
	"regexp"
	"strings"
)

var caseInsensitiveFlag = "(?i)"
var specialChars = []string{
	"\\", ".", "^", "$", "(", ")", "[", "]", "{", "}", "?", "+", "|", "/"}

type Matcher struct {
	patterns        []*regexp.Regexp
	originalStrings []string
}

func (m *Matcher) MatchesAny(str string) *string {
	for i, re := range m.patterns {
		if re.MatchString(str) {
			return &m.originalStrings[i]
		}
	}

	return nil
}

func (m *Matcher) IsEmpty() bool {
	return len(m.patterns) == 0
}

func NewMatcher(strs []string) Matcher {
	patterns := []*regexp.Regexp{}

	for _, s := range strs {
		re, err := wildcardToRegex(s)
		if err != nil {
			fmt.Println("error")
		} else {
			patterns = append(patterns, re)
		}
	}

	return Matcher{
		patterns:        patterns,
		originalStrings: strs,
	}
}

func escapeRegex(s string) string {
	for _, ch := range specialChars {
		s = strings.ReplaceAll(s, ch, "\\"+ch)
	}
	return s
}

func wildcardToRegex(wildcardStr string) (*regexp.Regexp, error) {
	escapedStr := escapeRegex(wildcardStr)
	regexStr := strings.ReplaceAll(escapedStr, "%", ".*")

	re, err := regexp.Compile(caseInsensitiveFlag + regexStr)

	if err != nil {
		return nil, fmt.Errorf("error compiling regex (%v)", err)
	}

	return re, nil
}
