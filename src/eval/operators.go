package eval

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"reflect"
)

type Operator int

const (
	Mul Operator = iota
	Div
	Mod
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
	"*":  Mul,
	"/":  Div,
	"%":  Mod,
	"+":  Add,
	"-":  Sub,
	"<":  Lt,
	">":  Gt,
	"<=": Lte,
	">=": Gte,
	"==": Eq,
	"!=": Ne,
	"&&": And,
	"||": Or,
}

var operatorSymbolMap = map[Operator]string{
	Mul: "*",
	Div: "/",
	Mod: "%",
	Add: "+",
	Sub: "-",
	Lt:  "<",
	Gt:  ">",
	Lte: "<=",
	Gte: ">=",
	Eq:  "==",
	Ne:  "!=",
	And: "&&",
	Or:  "||",
}

func (o *Operator) Capture(s []string) error {
	var ok bool
	*o, ok = operatorMap[s[0]]
	if !ok {
		panic(fmt.Sprintf("Unsupported operator: %s", s[0]))
	}
	return nil
}

func (o *Operator) String() string {
	return operatorSymbolMap[*o]
}

// Compute the result of the left and right operands with the referred operator. Internally this calls the package wide
// Compute method.
func (o *Operator) Compute(op1 *data.Value, op2 *data.Value) (err error, result *data.Value) {
	return Compute(*o, op1, op2)
}

// o is the "ID" for oFunc. This needs to be created so that there is an identity that can be checked for equivalence.
var o = oFunc

// operatorTable is a lookup which contains the functions which carry out operation calls. Each row represents the
// operator that is being called. Whereas, each column represents the type on the left-hand side of the expression.
var operatorTable = [13][8]func(op1 *data.Value, op2 *data.Value) (err error, result *data.Value) {
	/*           NoType    Object    Array    String    Number    Boolean    Null    Function                    */
	/* Mul */ {       o,        o,       o, muString, muNumber, anBoolean,    op1,          o},
	/* Div */ {       o, diObject,       o,        o, diNumber, diBoolean,    op1,          o},
	/* Mod */ {       o,        o,       o, moString, moNumber, moBoolean,    op1,          o},
	/* Add */ {       o, adObject, adArray, adString, adNumber, orBoolean,    op1,          o},
	/* Sub */ {       o, suObject, suArray, suString, suNumber, suBoolean,    op1,          o},
	/* Lt  */ {       o, ltObject, ltArray, ltString, ltNumber, ltBoolean,      o,          o},
	/* Gt  */ {       o, gtObject, gtArray, gtString, gtNumber, gtBoolean,      o,          o},
	/* Lte */ {       o, leObject, leArray, leString, leNumber, leBoolean,      o,          o},
	/* Gte */ {       o, geObject, geArray, geString, geNumber, geBoolean,      o,          o},
	/* Eq  */ {       o, eqObject, eqArray, eqString, eqNumber, eqBoolean, eqNull,          o},
	/* Ne  */ {       o, neObject, neArray, neString, neNumber, neBoolean, neNull,          o},
	/* And */ {       o, anObject, anArray, anString, anNumber, anBoolean, anNull,          o},
	/* Or  */ {       o, orObject, orArray, orString, orNumber, orBoolean, orNull,          o},
}

// Compute will compute the result of the given binary operation with the given left and right operands. Internally this
// uses the operatorTable and looks up the operator value and the Type of the left symbol. The right operand's type is
// not looked up because all operands are left associative. If the computation for the given operator and left-hand type
// does not exist, the errors.InvalidOperation error will be thrown. Otherwise, the result of the computation will be
// returned.
func Compute(operator Operator, left *data.Value, right *data.Value) (err error, result *data.Value) {
	// If the operatorTable entry points to o then we will return an InvalidOperation error.
	//fmt.Println(operator.String(), left.Value, left.Type.String(), reflect.TypeOf(left.Value).String())
	//fmt.Println(right.Value, right.Type.String(), reflect.TypeOf(right.Value).String())
	if reflect.ValueOf(operatorTable[operator][left.Type]).Pointer() == reflect.ValueOf(o).Pointer() {
		return errors.InvalidOperation.Errorf(operator.String(), left.Type.String(), right.Type.String()), nil
	}
	// Otherwise, we return the result of the computation method.
	return operatorTable[operator][left.Type](left, right)
}
