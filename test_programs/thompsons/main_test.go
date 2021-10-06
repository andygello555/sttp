package main

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, test := range []string{
		"(a | b*) a (b | e)*",
		"c*(a|b)((a|c)b)*",
		"a*a(ba|(b|e))(b|e)*",
	}{
		_, parsed := Parse(test)
		if parsed.String() != strings.Replace(test, " ", "", -1) {
			t.Errorf("Parsed regular expression: '%s', does not match '%s'", test, parsed)
		}
	}
}

func TestRegex_Thompson(t *testing.T) {
	//for _, test := range []struct{
	//
	//}
}
