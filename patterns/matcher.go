package patterns

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"sync"
)

var caseInsensitiveFlag = "(?i)"
var specialChars = []string{
	"\\", ".", "^", "$", "(", ")", "[", "]", "{", "}", "?", "*", "+", "|", "/"}

type Matcher struct {
	patterns        []*regexp.Regexp
	originalStrings []string
	mu              sync.RWMutex
}

func (m *Matcher) MatchesAny(str string) *string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for i, re := range m.patterns {
		if re.MatchString(str) {
			return &m.originalStrings[i]
		}
	}

	return nil
}

func (m *Matcher) IsEmpty() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
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
			log.Println(err)
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
