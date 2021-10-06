package main

import (
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"strings"
)

var Lexer = lexer.MustSimple([]lexer.Rule{
	{"Punct", `[()*|]`, nil},
	{"Char", `[a-z]`, nil},
	{"whitespace", `\s*`, nil},
})

type Base struct {
	Tokens []lexer.Token

	Char  *string `  @Char`
	Regex *Regex  `| "(" @@ ")"`
}

type Factor struct {
	Tokens []lexer.Token

	Base    *Base `@@`
	Closure bool  `@"*"?`
}

type Term struct {
	Tokens []lexer.Token

	Factors []*Factor `@@+`
}

type Regex struct {
	Tokens []lexer.Token

	Term  *Term  `@@ ("|"`
	Regex *Regex `@@)?`
}

func (b *Base) String() string {
	if b.Char != nil {
		return *(b.Char)
	}
	return fmt.Sprintf("(%s)", b.Regex.String())
}

func (f *Factor) String() string {
	out := f.Base.String()
	if f.Closure {
		out += "*"
	}
	return out
}

func (t *Term) String() string {
	out := make([]string, 0)
	for _, factor := range t.Factors {
		out = append(out, factor.String())
	}
	return strings.Join(out, "")
}

func (r *Regex) String() string {
	out := r.Term.String()
	if r.Regex != nil {
		out += fmt.Sprintf("|%s", r.Regex.String())
	}
	return out
}

func Parse(regexString string) (error, *Regex) {
	parser := participle.MustBuild(&Regex{},
		participle.Lexer(Lexer),
	)

	regex := &Regex{}
	if err := parser.ParseString("", regexString, regex); err != nil {
		return err, regex
	}
	return nil, regex
}
