package patterns

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
)

var caseInsensitiveFlag = "(?i)"
var specialChars = []string{
	"\\", ".", "^", "$", "(", ")", "[", "]", "{", "}", "?", "+", "|", "/"}

// TODO: Maybe add a map so that we can first check the map, and if it's not
// there, do a linear search.
type Matcher struct {
	patterns        []*regexp.Regexp
	originalStrings []string
	// TODO: Should this have a different kind of mutex?.
	mu sync.Mutex
}

func (m *Matcher) MatchesAny(str string) *string {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i, re := range m.patterns {
		if re.MatchString(str) {
			return &m.originalStrings[i]
		}
	}

	return nil
}

func (m *Matcher) IsEmpty() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.patterns) == 0
}

func (m *Matcher) Set(strs []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.patterns = []*regexp.Regexp{}
	m.originalStrings = strs

	for _, s := range strs {
		if len(s) == 0 {
			continue
		}

		re, err := wildcardToRegex(s)
		if err != nil {
			fmt.Println("error")
		} else {
			m.patterns = append(m.patterns, re)
		}
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
