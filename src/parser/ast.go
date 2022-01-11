package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
	"strings"
)

type Null bool

func (n *Null) Capture(values []string) error {
	*n = values[0] == "null"
	return nil
}

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

// JSON describes a JSON literal.
type JSON struct {
	Pos lexer.Position

	Object *Object `@@ |`
	Array  *Array  `@@`
}

// Object describes a JSON object within a JSON literal.
type Object struct {
	Pos lexer.Position

	Pairs []*Pair `"{" (@@ ( "," @@)*)? "}"`
}

// Pair describes a key-value pair inside a JSON object within a JSON literal.
type Pair struct {
	Pos lexer.Position

	Key   *Expression `@@`
	Value *Expression `":" @@`
}

// Array describes an array within a JSON literal.
type Array struct {
	Pos lexer.Position

	Elements []*Expression `"[" (@@ ( "," @@)*)? "]"`
}

// Factor represents a factor within an expression.
type Factor struct {
	Pos lexer.Position

	Null          *Null         `  @Null`
	Boolean       *Boolean      `| @(True | False)`
	Number        *float64      `| @Number`
	StringLit     *string       `| @StringLit`
	JSONPath      *JSONPath     `| @@`
	JSON          *JSON         `| @@`
	FunctionCall  *FunctionCall `| @@`
	MethodCall    *MethodCall   `| @@`
	SubExpression *Expression   `| "(" @@ ")"`
}

type Prec1Term struct {
	Pos lexer.Position

	Left  *Factor  `@@`
	Right []*Prec0 `@@*`
}

// Prec0 captures expressions which use multiply and divide.
type Prec0 struct {
	Pos lexer.Position

	Operator eval.Operator `@("*" | "/" | "%")`
	Factor   *Factor       `@@`
}

type Prec2Term struct {
	Pos lexer.Position

	Left  *Prec1Term `@@`
	Right []*Prec1   `@@*`
}

// Prec1 captures expressions which use plus and minus.
type Prec1 struct {
	Pos lexer.Position

	Operator eval.Operator `@("+" | "-")`
	Factor   *Prec1Term    `@@`
}

type Prec3Term struct {
	Pos lexer.Position

	Left  *Prec2Term `@@`
	Right []*Prec2   `@@*`
}

// Prec2 captures expressions which use less than, greater than, less than or equal to and greater than or equal to.
type Prec2 struct {
	Pos lexer.Position

	Operator eval.Operator `@("<" | ">" | "<=" | ">=")`
	Factor   *Prec2Term    `@@`
}

type Prec4Term struct {
	Pos lexer.Position

	Left  *Prec3Term `@@`
	Right []*Prec3   `@@*`
}

// Prec3 captures expressions which use not equal and equal.
type Prec3 struct {
	Pos lexer.Position

	Operator eval.Operator `@("!=" | "==")`
	Factor   *Prec3Term    `@@`
}

type Prec5Term struct {
	Pos lexer.Position

	Left  *Prec4Term `@@`
	Right []*Prec4   `@@*`
}

// Prec4 captures expressions which use logical and.
type Prec4 struct {
	Pos lexer.Position

	Operator eval.Operator `@"&&"`
	Factor   *Prec4Term    `@@`
}

type Expression struct {
	Pos lexer.Position

	Left  *Prec5Term `@@`
	Right []*Prec5   `@@*`
}

// Prec5 captures expressions which use logical or.
type Prec5 struct {
	Pos lexer.Position

	Operator eval.Operator `@"||"`
	Factor   *Prec5Term    `@@`
}

// FunctionBody describes what follows a function identifier.
type FunctionBody struct {
	Pos lexer.Position

	Parameters []*JSONPath `"(" ( @@ ( "," @@ )* )? ")"`
	Block      *Block      `@@ End`
}

// MethodCall describes a call to a HTTP method.
type MethodCall struct {
	Pos lexer.Position

	Method    eval.Method   `"$" @Method`
	Arguments []*Expression `"(" (@@ ( "," @@ )*)? ")"`
}

// FunctionCall describes a call to a function.
type FunctionCall struct {
	Pos lexer.Position

	JSONPath  *JSONPath     `"$" @@`
	Arguments []*Expression `"(" (@@ ( "," @@ )*)? ")"`
}

// ReturnStatement describes a return statement which can be at the end of any block.
type ReturnStatement struct {
	Pos lexer.Position

	Value *Expression `Return @@? ";"`
}

// ThrowStatement describes an expression that is thrown which can be at the end of any block.
type ThrowStatement struct {
	Pos lexer.Position

	Value *Expression `Throw @@? ";"`
}

// TestStatement describes an expression that will be evaluated and tested by the interpreter and returned to user.
type TestStatement struct {
	Pos lexer.Position

	Expression *Expression `Test @@`
}

// JSONPath describes a path to a property within a variable.
type JSONPath struct {
	Pos lexer.Position

	Parts []*Part `@@ ( "." @@ )*`
}

// Part describes an index or property within a JSONPath.
type Part struct {
	Pos lexer.Position

	Property *string  `@Ident`
	Indices  []*Index `@@*`
}

type Index struct {
	ExpressionIndex *Expression `  "[" @@ "]"`
	FilterIndex     *Block      `| Filter @@ Filter`
}

// Statement describes one of the statements that can be used in each "line" of a Block.
type Statement struct {
	Pos lexer.Position

	Assignment         *Assignment         `  @@`
	FunctionCall       *FunctionCall       `| @@`
	MethodCall         *MethodCall         `| @@`
	Break              *string             `| @Break`
	Test               *TestStatement      `| @@`
	While              *While              `| @@`
	For                *For                `| @@`
	ForEach            *ForEach            `| @@`
	Batch              *Batch              `| @@`
	TryCatch           *TryCatch           `| @@`
	FunctionDefinition *FunctionDefinition `| @@`
	IfElifElse         *IfElifElse         `| @@`
}

// Assignment describes an assignment to a property pointed to by a JSONPath.
type Assignment struct {
	Pos lexer.Position

	JSONPath *JSONPath   `@@ "="`
	Value    *Expression `@@`
}

// While loop.
type While struct {
	Pos lexer.Position

	Condition *Expression `While @@ Do`
	Block     *Block      `@@ End`
}

// For loop with assignment, condition, and step.
type For struct {
	Pos lexer.Position

	Var        *Assignment `For @@ ";"`
	Condition  *Expression `@@`
	Step       *Assignment `(";" @@)?`
	Block      *Block      `Do @@ End`
}

// ForEach loop with iterator(s).
type ForEach struct {
	Pos lexer.Position

	Key   *string     `For @Ident`
	Value *string     `("," @Ident)?`
	In    *Expression `In @@ Do`
	Block *Block      `@@ End`
}

// Batch describes a block of code where all HTTP method calls are executed in parallel.
type Batch struct {
	Pos lexer.Position

	Block *Block `Batch This @@ End`
}

// TryCatch describes a try-catch structure. The "as" segment must always be defined so a variable can be allocated with
// the caught exception.
type TryCatch struct {
	Pos lexer.Position

	Try     *Block  `Try This @@`
	CatchAs *string `Catch As @Ident Then`
	Caught  *Block  `@@ End`
}

// FunctionDefinition describes the definition of function.
type FunctionDefinition struct {
	Pos lexer.Position

	JSONPath *JSONPath     `Function @@`
	Body     *FunctionBody `@@`
}

// IfElifElse is the main construct which defines a if-elif-else statement.
type IfElifElse struct {
	Pos lexer.Position

	IfCondition *Expression `If @@ Then`
	IfBlock     *Block      `@@`
	Elifs       []*Elif     `@@*`
	Else        *Block      `(Else @@)? End`
}

// Elif is used within IfElifElse to match Elif branches.
type Elif struct {
	Pos lexer.Position

	Condition *Expression `Elif @@ Then`
	Block     *Block      `@@`
}

// Block describes a "block" of statements which might end with a return statement.
type Block struct {
	Pos lexer.Position

	Statements []*Statement     `( @@? ";" )*`
	Return     *ReturnStatement `( @@ |`
	Throw      *ThrowStatement  `  @@ )?`
}

// Program describes an entire program which is just a Block of statements.
type Program struct {
	Pos lexer.Position

	Block *Block `@@`
}

var lex = lexer.MustSimple([]lexer.Rule{
	{"comment", `//.*`, nil},

	{"StringLit", `(")([^"\\]*(?:\\.[^"\\]*)*)(")`, nil},
	{"Method", fmt.Sprintf("(%s)", strings.Join(eval.MethodStrings(), "|")), nil},
	{"While", `while\s`, nil},
	{"For", `for\s`, nil},
	{"Do", `\sdo\s`, nil},
	{"This", `this\s`, nil},
	{"Break", `break`, nil},
	{"Then", `\sthen\s`, nil},
	{"End", `end`, nil},
	{"Function", `function\s`, nil},
	{"Return", `return`, nil},
	{"Throw", `throw`, nil},
	{"If", `if\s`, nil},
	{"Elif", `elif\s`, nil},
	{"Else", `else\s`, nil},
	{"Catch", `catch\s`, nil},
	{"Test", `test\s`, nil},
	{"In", `\sin\s`, nil},
	{"As", `as\s`, nil},
	{"True", `true`, nil},
	{"False", `false`, nil},
	{"Null", `null`, nil},
	{"Batch", `batch\s`, nil},
	{"Try", `try\s`, nil},
	{"Number", `[-+]?(\d*\.)?\d+`, nil},
	{"Operators", `\|\||&&|<=|>=|!=|==|[-+*/%=!<>]`, nil},
	{"Filter", "```", nil},
	{"Punct", `[$;,.(){}:]|\[|\]`, nil},
	{"Ident", `[a-zA-Z_]\w*`, nil},
	{"whitespace", `\s+`, nil},
})

func Parse(filename, s string) (error, *Program) {
	parser := participle.MustBuild(&Program{},
		participle.Lexer(lex),
		participle.CaseInsensitive("Ident"),
		participle.Unquote("StringLit"),
		participle.UseLookahead(2),
	)
	program := &Program{}
	return parser.ParseString(filename, s, program), program
}
