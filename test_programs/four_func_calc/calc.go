package main

import (
	"fmt"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

//var Lexer = lexer.MustSimple([]lexer.Rule{
//	{"EOL", `[\n\r]+`, nil},
//	{"EOF", EOF, nil},
//	{"Char", scanner.Char, nil},
//	{"Ident", scanner.Ident, nil},
//	{"Int", scanner.Int, nil},
//	{"Float", scanner.Float, nil},
//	{"String", scanner.String, nil},
//	{"RawString", scanner.RawString, nil},
//	{"Comment", scanner.Comment, nil},
//})

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
	Number      *float64    `  @(Float|Int)`
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
	Left *Term       `@@`
	Right []*OpTerm  `@@*`
}

// An Assignment consists of a variable identifier followed by the declaration+assignment operator followed by an
// Expression.
type Assignment struct {
	Variable *string        `@Ident ":="`
	Expression *Expression  `@@`
}

// A Statement can either be an expression or a variable assignment.
type Statement struct {
	Expression *Expression `@@`
	Assignment *Assignment `| @@`
	Eol        *string     `(EOL | ";")`
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

func (a *Assignment) Eval(ctx Memory) {
	ctx[*(a.Variable)] = a.Expression.Eval(ctx)
}

func (s *Statement) Eval(ctx Memory) float64 {
	if s.Assignment != nil {
		s.Assignment.Eval(ctx)
		return 0
	}
	return s.Expression.Eval(ctx)
}

func BuildParser(grammar interface{}) *participle.Parser {
	return participle.MustBuild(grammar,
		participle.Lexer(Lexer),
	)
}
