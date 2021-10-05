package main

import (
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for _, test := range []string{
		"(a | b*) a (b | e)*",
	}{
		_, parsed := Parse(test)
		if parsed.String() != strings.Replace(test, " ", "", -1) {
			t.Errorf("Parsed regular expression: '%s', does not match '%s'", test, parsed)
		}
	}
}
