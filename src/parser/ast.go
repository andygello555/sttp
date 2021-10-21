package parser

import (
	"fmt"
	"github.com/alecthomas/participle/v2/lexer"
)

type Null bool

type Boolean bool

func (b *Boolean) Capture(values []string) error {
	*b = values[0] == "true"
	return nil
}

type Operator int

const (
	Mul Operator = iota
	Div
	Add
	Sub
	Lt
	Gt
	Lte
	Gte
	Eq
	Ne
	And
	Or
)

var operatorMap = map[string]Operator{
	"*": Mul,
	"/": Div,
	"+": Add,
	"-": Sub,
	"<": Lt,
	">": Gt,
	"<=": Lte,
	">=": Gte,
	"==": Eq,
	"!=": Ne,
	"&&": And,
	"||": Or,
}

func (o *Operator) Capture(s []string) error {
	var ok bool
	*o, ok = operatorMap[s[0]]
	if !ok {
		panic(fmt.Sprintf("Unsupported operator: %s", s[0]))
	}
	return nil
}

type Method int

const (
	GET Method = iota
	HEAD
	POST
	PUT
	DELETE
	CONNECT
	OPTIONS
	TRACE
	PATCH
)

var methodMap = map[string]Method{
	"GET": GET,
	"HEAD": HEAD,
	"POST": POST,
	"PUT": PUT,
	"DELETE": DELETE,
	"CONNECT": CONNECT,
	"OPTIONS": OPTIONS,
	"TRACE": TRACE,
	"PATCH": PATCH,
}

func (m *Method) Capture(s []string) error {
	var ok bool
	*m, ok = methodMap[s[0]]
	if !ok {
		panic(fmt.Sprintf("Unsupported HTTP method: %s", s[0]))
	}
	return nil
}

var lex = lexer.MustSimple([]lexer.Rule{
	{"Number", `([+-]?[0-9]*[.])?[0-9]+`, nil},
	{"StringLit", `"(\\"|[^"])*"`, nil},
	{"Ident", `[a-zA-Z_]\w*`, nil},
	{"Method", `(GET|HEAD|POST|PUT|DELETE|CONNECT|OPTIONS|TRACE|PATCH)`, nil},
	{"While", `while`, nil},
	{"For", `for`, nil},
	{"Do", `do`, nil},
	{"Break", `break`, nil},
	{"Then", `then`, nil},
	{"End", `end`, nil},
	{"Function", `function`, nil},
	{"Return", `return`, nil},
	{"If", `if`, nil},
	{"Elif", `elif`, nil},
	{"Else", `else`, nil},
	{"In", `in`, nil},
	{"True", `true`, nil},
	{"False", `false`, nil},
	{"Null", `null`, nil},
	{"Punct", `[-,()*/+{};&!=:<>]|\[|\]`, nil},
	{"EOL", `[\n\r]+`, nil},
	{"comment", `//.*|/\*.*?\*/`, nil},
	{"whitespace", `\s+`, nil},
})

// JSON describes a JSON literal.
type JSON struct {
	Object *Object `@@ |`
	Array  *Array  `@@`
}

// Object describes a JSON object within a JSON literal.
type Object struct {
	Pairs []*Pair `"{" (@@ ( "," @@)*)? "}"`
}

// Pair describes a key-value pair inside a JSON object within a JSON literal.
type Pair struct {
	Key   *Expression `@@`
	Value *Expression `":" @@`
}

// Array describes an array within a JSON literal.
type Array struct {
	Elements []*Expression `"{" (@@ ( "," @@)*)? "}"`
}

// Factor represents a factor within an expression.
type Factor struct {
	Null          *Null         `  @Null?`
	Boolean       *Boolean      `| @(True | False)`
	Number        *float64      `| @Number`
	String        *string       `| @StringLit`
	JSONPath      *JSONPath     `| @@`
	JSON          *JSON         `| @@`
	FunctionCall  *FunctionCall `| @@`
	MethodCall    *MethodCall   `| @@`
	SubExpression *Expression   `| "(" @@ ")"`
}

type Prec1Term struct {
	Factor *Factor  `@@`
	Next   []*Prec0 `@@*`
}

// Prec0 captures expressions which use multiply and divide.
type Prec0 struct {
	Operator Operator `@("*" | "/")`
	Factor   *Factor  `@@`
	Same     *Prec0   `@@`
}

type Prec2Term struct {
	Factor *Prec1Term `@@`
	Next   []*Prec1   `@@*`
}

// Prec1 captures expressions which use plus and minus.
type Prec1 struct {
	Operator Operator   `@("+" | "-")`
	Factor   *Prec1Term `@@`
	Same     *Prec1     `@@`
}

type Prec3Term struct {
	Factor *Prec2Term `@@`
	Next   []*Prec2   `@@*`
}

// Prec2 captures expressions which use less than, greater than, less than or equal to and greater than or equal to.
type Prec2 struct {
	Operator Operator   `@("<" | ">" | "<=" | ">=")`
	Factor   *Prec2Term `@@`
	Same     *Prec2     `@@`
}

type Prec4Term struct {
	Factor *Prec3Term `@@`
	Next   []*Prec3   `@@*`
}

// Prec3 captures expressions which use not equal and equal.
type Prec3 struct {
	Operator Operator   `@("!=" | "==")`
	Factor   *Prec3Term `@@`
	Same     *Prec3     `@@`
}

type Prec5Term struct {
	Factor *Prec4Term `@@`
	Next   []*Prec4   `@@*`
}

// Prec4 captures expressions which use logical and.
type Prec4 struct {
	Operator Operator   `@"&&"`
	Factor   *Prec4Term `@@`
	Same     *Prec4     `@@`
}

type Expression struct {
	Factor *Prec5Term `@@`
	Next   []*Prec5   `@@*`
}

// Prec5 captures expressions which use logical or.
type Prec5 struct {
	Operator Operator   `@"||"`
	Factor   *Prec5Term `@@`
	Same     *Prec5     `@@`
}

// FunctionBody describes what follows a function identifier.
type FunctionBody struct {
	Parameters []*string `"(" (@Ident ( "," @Ident )*)? ")"`
	Block      *Block    `@@ End`
}

// MethodCall describes a call to a HTTP method.
type MethodCall struct {
	Method    Method        `@Method`
	Arguments []*Expression `"(" (@@ ( "," @@ )*)? ")"`
}

// FunctionCall describes a call to a function.
type FunctionCall struct {
	JSONPath  *JSONPath     `@@`
	Arguments []*Expression `"(" (@@ ( "," @@ )*)? ")"`
}

// ReturnStatement describes a return statement which can at the end of any block.
type ReturnStatement struct {
	Value *Expression `Return @@? ";"`
}

// JSONPath describes a path to a property within a variable.
type JSONPath struct {
	Parts []*Part `@@ ( "." @@ )*`
}

// Part describes an index or property within a JSONPath.
type Part struct {
	Property *string       `@Ident`
	Indices  []*Expression `( "[" @@ "]" )*`
}

// Statement describes one of the statements that can be used in each "line" of a Block.
type Statement struct {
	Assignment         *Assignment         `(   @@`
	FunctionCall       *FunctionCall       `  | @@`
	MethodCall         *MethodCall         `  | @@`
	Break              bool                `  | @Break?`
	While              *While              `  | @@`
	For                *For                `  | @@`
	ForEach            *ForEach            `  | @@`
	FunctionDefinition *FunctionDefinition `  | @@`
	IfElifElse         *IfElifElse         `  | @@ )? ";"`
}

// Assignment describes an assignment to a property pointed to by a JSONPath.
type Assignment struct {
	JSONPath *JSONPath   `@@ "="`
	Value    *Expression `@@`
}

// While loop.
type While struct {
	Condition *Expression `While @@ Do`
	Block     *Block      `@@ End`
}

// For loop with assignment, condition, and step.
type For struct {
	Var        *string     `For @Ident "="`
	StartValue *Expression `@@ ";"`
	Condition  *Expression `@@`
	Step       *Expression `(";" @@)?`
	Block      *Block      `Do @@ End`
}

// ForEach loop with iterator(s).
type ForEach struct {
	Key   *string     `@Ident`
	Value *string     `("," @Ident)?`
	In    *Expression `In @@ Do`
	Block *Block      `@@ End`
}

// FunctionDefinition describes the definition of function.
type FunctionDefinition struct {
	JSONPath *JSONPath     `Function @@`
	Body     *FunctionBody `@@`
}

// IfElifElse is the main construct which defines a if-elif-else statement.
type IfElifElse struct {
	IfCondition *Expression `If @@ Then`
	IfBlock     *Block      `@@`
	Elifs       []*Elif     `@@*`
	Else        *Block      `@@? End`
}

// Elif is used within IfElifElse to match Elif branches.
type Elif struct {
	Condition *Expression `Elif @@ Then`
	Block     *Block      `@@`
}

// Block describes a "block" of statements which might end with a return statement.
type Block struct {
	Statements []*Statement     `@@*`
	Return     *ReturnStatement `@@?`
}

// Program describes an entire program which is just a Block of statements.
type Program struct {
	Block *Block `@@`
}
