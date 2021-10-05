package main

import (
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"strings"
)

var Lexer = lexer.MustSimple([]lexer.Rule{
	{"Number", `[+-]?([0-9]*[.])?[0-9]+`, nil},
	{"Ident", `[a-zA-Z_]\w*`, nil},
	{"EOL", `[\n\r]+`, nil},
	// We add punctuation matching to our Lexer as these characters need to be consumed
	{"Punct", `[-,()*/+%{};&!=:<>]|\[|\]`, nil},
	{"whitespace", `\s*`, nil},
})

// Operator describes the type which all operators must be declared as.
type Operator int

// Memory describes the current memory map for the calculator. Mapping of variable identifiers to their values.
type Memory map[string]float64

// Supported operators.
const (
	Mul Operator = iota
	Div
	Add
	Sub
)

var operatorMap = map[string]Operator{
	"+": Add,
	"-": Sub,
	"*": Mul,
	"/": Div,
}

func (o *Operator) Capture(s []string) error {
	var ok bool
	*o, ok = operatorMap[s[0]]
	if !ok {
		panic(fmt.Sprintf("Unsupported operator: %s", s[0]))
	}
	return nil
}

// A Factor can be either a number, variable identifier, or a set of parenthesis containing another Expression. Note
// that there is no defined way of taking the exponent in this grammar.
type Factor struct {
	Number      *float64    `  @Number`
	Variable    *string     `| @Ident`
	Parenthesis *Expression `| "(" @@ ")"`
}

// A Product consists of a multiplication or subtraction operator followed by a Factor. Product is defined lower down
// in the grammar in order to impose evaluation precedence on these tokens.
type Product struct {
	Operator Operator `@("*" | "/")`
	Factor   *Factor  `@@`
}

// A Term is a Factor followed by none or many Products.
type Term struct {
	Left  *Factor    `@@`
	Right []*Product `@@*`
}

// An OpTerm consists of an addition or subtraction operator followed by a Term.
type OpTerm struct {
	Operator Operator `@("+" | "-")`
	Term     *Term    `@@`
}

// An Expression is a Term followed by none or many OpTerms.
type Expression struct {
	Tokens     []lexer.Token
	Left *Term      `@@`
	Right []*OpTerm `@@*`
}

// A Clear statement is the keyword "clear" followed by a variable identifier.
type Clear struct {
	Variable *string `"clear" @Ident`
}

// An Assignment consists of the string "let", followed by a variable identifier followed by the declaration+assignment
// operator followed by an Expression.
type Assignment struct {
	Tokens     []lexer.Token
	Variable   *string     `"let" @Ident`
	Expression *Expression `"=" @@`
}


// A Statement can either be an expression or a variable assignment.
type Statement struct {
	Tokens     []lexer.Token
	Clear      *Clear      `(   @@`
	Assignment *Assignment `  | @@`
	Expression *Expression `  | @@ ) EOL`
}

func (o Operator) String() string {
	switch o {
	case Mul:
		return "*"
	case Div:
		return "/"
	case Sub:
		return "-"
	case Add:
		return "+"
	}
	panic("unsupported operator")
}

func (v *Factor) String() string {
	if v.Number != nil {
		return fmt.Sprintf("%g", *v.Number)
	}
	if v.Variable != nil {
		return *v.Variable
	}
	return "(" + v.Parenthesis.String() + ")"
}

func (p *Product) String() string {
	return fmt.Sprintf("%s %s", p.Operator.String(), p.Factor.String())
}

func (t *Term) String() string {
	out := []string{t.Left.String()}
	for _, r := range t.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (o *OpTerm) String() string {
	return fmt.Sprintf("%s %s", o.Operator, o.Term)
}

func (e *Expression) String() string {
	fmt.Printf("%p\n", e)
	out := []string{e.Left.String()}
	for _, r := range e.Right {
		out = append(out, r.String())
	}
	return strings.Join(out, " ")
}

func (c *Clear) String() string {
	return fmt.Sprintf("clear %s", *(c.Variable))
}

func (a *Assignment) String() string {
	return fmt.Sprintf("%s = %s", *(a.Variable), a.Expression.String())
}

func (s *Statement) String() string {
	switch {
	case s.Clear != nil:
		return s.Clear.String()
	case s.Assignment != nil:
		return s.Assignment.String()
	}
	return s.Expression.String()
}

func (o Operator) Eval(l, r float64) float64 {
	switch o {
	case Mul:
		return l * r
	case Div:
		return l / r
	case Add:
		return l + r
	case Sub:
		return l - r
	default:
		panic("Unsupported operator")
	}
}

func (v *Factor) Eval(ctx Memory) float64 {
	switch {
	case v.Number != nil:
		return *v.Number
	case v.Variable != nil:
		value, ok := ctx[*v.Variable]
		if !ok {
			panic("no such variable " + *v.Variable)
		}
		return value
	default:
		return v.Parenthesis.Eval(ctx)
	}
}

func (t *Term) Eval(ctx Memory) float64 {
	n := t.Left.Eval(ctx)
	for _, r := range t.Right {
		n = r.Operator.Eval(n, r.Factor.Eval(ctx))
	}
	return n
}

func (e *Expression) Eval(ctx Memory) float64 {
	l := e.Left.Eval(ctx)
	for _, r := range e.Right {
		l = r.Operator.Eval(l, r.Term.Eval(ctx))
	}
	return l
}

func (c *Clear) Eval(ctx Memory) {
	delete(ctx, *(c.Variable))
}

func (a *Assignment) Eval(ctx Memory) {
	ctx[*(a.Variable)] = a.Expression.Eval(ctx)
}

func (s *Statement) Eval(ctx Memory) (float64, *Memory) {
	switch {
	case s.Clear != nil:
		s.Clear.Eval(ctx)
		return 0, &ctx
	case s.Assignment != nil:
		s.Assignment.Eval(ctx)
		return 0, &ctx
	}
	return s.Expression.Eval(ctx), &ctx
}

func BuildParser(grammar interface{}) *participle.Parser {
	return participle.MustBuild(grammar,
		participle.Lexer(Lexer),
		participle.CaseInsensitive("Ident"),
		participle.UseLookahead(2),
	)
}
