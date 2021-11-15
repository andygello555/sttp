package eval

import (
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"strings"
)

// oFunc is a dummy function which exists within the operatorTable to represent an operation which cannot be made.
func oFunc(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return nil, nil
}

// muString: Multiply Object. Will repeat the String on the left if there is a number on the right.
func muString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	if op2.Type == data.Number {
		return nil, &data.Symbol{
			Value: strings.Repeat(op1.Value.(string), int(op2.Value.(float64))),
			Type:  data.String,
			Scope: op1.Scope,
		}
	}
	return errors.InvalidOperation.Errorf("*", op1.Type, op2.Type), nil
}

// muNumber: Multiply Number.
func muNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2Number *data.Symbol
	err, op2Number = Cast(op2, data.Number)
	if err != nil {
		return err, nil
	}
	return nil, &data.Symbol{
		Value: op1.Value.(float64) * op2Number.Value.(float64),
		Type:  data.Number,
		Scope: op1.Scope,
	}
}

// and performs a logical AND after casting the RHS to a Boolean.
func and(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2Bool *data.Symbol
	err, op2Bool = Cast(op2, data.Boolean)
	if err != nil {
		return err, nil
	}
	return nil, &data.Symbol{
		Value: op1.Value.(bool) && op2Bool.Value.(bool),
		Type:  data.Boolean,
		Scope: op1.Scope,
	}
}

// muBoolean: Multiply Boolean. Casts rhs to Boolean then performs a logical AND operation.
func muBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return and(op1, op2)
}

// anBoolean: And Boolean. Casts rhs to Boolean then performs a logical AND operation.
func anBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return and(op1, op2)
}
