package patterns

import (
	"testing"
)

func TestMatchesAny(t *testing.T) {
	var matcher Matcher
	matcher.Set([]string{
		"host.com", "hello", "world",
		"foo%", "bar", "bonjour%%%friend.%%", "%\\%",
		"漢字",
	})

	tests := []struct {
		input    string
		expected *string
	}{
		{"HOst.com", &matcher.originalStrings[0]},
		{"HOst..com", nil},
		{"HOstcom", nil},
		{"hello", &matcher.originalStrings[1]},
		{"heLlO", &matcher.originalStrings[1]},
		{"world", &matcher.originalStrings[2]},
		{"foobar", &matcher.originalStrings[3]},
		{"baz", nil},
		{"Bonjour-My-Friend.Com", &matcher.originalStrings[5]},
		{"BonjourFriend.", &matcher.originalStrings[5]},
		{"BonjourFriend", nil},
		{"hola\\mundo", &matcher.originalStrings[6]},
		{"私は漢字を", &matcher.originalStrings[7]},
	}

	for _, test := range tests {
		result := matcher.MatchesAny(test.input)
		if (result == nil && test.expected != nil) || (result != nil && test.expected == nil) {
			t.Errorf("MatchesAny(%s) = %v; want %v", test.input, result, test.expected)
		} else if result != nil && *result != *test.expected {
			t.Errorf("MatchesAny(%s) = %v; want %v", test.input, *result, *test.expected)
		}
	}
}

func TestSet(t *testing.T) {
	var matcher Matcher
	strs := []string{"89856", "!\"#$%&'()~*=~|{}`*P+<>><'"}
	matcher.Set(strs)
	for _, test := range strs {
		result := matcher.MatchesAny(test)
		if result == nil {
			t.Errorf("MatchesAny(%s) should have matched, but got nil", test)
		}
	}
}
