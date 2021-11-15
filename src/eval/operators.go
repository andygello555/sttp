package eval

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
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
func (o *Operator) Compute(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return Compute(*o, op1, op2)
}

// o is the "ID" for oFunc. This needs to be created so that there is an identity that can be checked for equivalence.
var o = oFunc

// operatorTable is a lookup which contains the functions which carry out operation calls. Each row represents the
// operator that is being called. Whereas, each column represents the type on the left-hand side of the expression.
var operatorTable = [13][8]func(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	/*           NoType    Object    Array    String    Number    Boolean    Null    Function                    */
	/* Mul */ {       o,        o,       o, muString, muNumber,   },
	/* Div */ {       o},
	/* Mod */ {       o},
	/* Add */ {       o},
	/* Sub */ {       o},
	/* Lt  */ {       o},
	/* Gt  */ {       o},
	/* Lte */ {       o},
	/* Gte */ {       o},
	/* Eq  */ {       o},
	/* Ne  */ {       o},
	/* And */ {       o},
	/* Or  */ {       o},
}

// Compute will compute the result of the given binary operation with the given left and right operands. Internally this
// uses the operatorTable and looks up the operator value and the Type of the left symbol. The right operand's type is
// not looked up because all operands are left associative. If the computation for the given operator and left-hand type
// does not exist, the errors.InvalidOperation error will be thrown. Otherwise, the result of the computation will be
// returned.
func Compute(operator Operator, left *data.Symbol, right *data.Symbol) (err error, result *data.Symbol) {
	// If the operatorTable entry points to o then we will return an InvalidOperation error.
	if &operatorTable[operator][left.Type] == &o {
		return errors.InvalidOperation.Errorf(operator.String(), left.Type.String(), right.Type.String()), nil
	}
	// Otherwise, we return the result of the computation method.
	return operatorTable[operator][left.Type](left, right)
}
