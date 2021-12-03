package eval

import (
	"encoding/json"
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/andygello555/gotils/slices"
	"math"
	"math/big"
	"strings"
)

// EqualInterface will check if the two operands (interface{}) are Equal and returns a boolean. Internally this will 
// Marshal both operands to JSON and compare the produced strings.
func EqualInterface(op1 interface{}, op2 interface{}) (err error, equal bool) {
	var a, b []byte
	a, err = json.Marshal(op1)
	b, err = json.Marshal(op2)
	return err, string(a) == string(b)
}

// Equal will check if the two operands are Equal and return a boolean.
func Equal(op1 *data.Symbol, op2 *data.Symbol) (err error, equal bool) {
	return EqualInterface(op1.Value, op2.Value)
}

// equalSymbol will check if the two operands are Equal and return a boolean Symbol.
func equalSymbol(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var ok bool
	err, ok = Equal(op1, op2)
	return nil, &data.Symbol{
		Value: ok,
		Type:  data.Boolean,
		Scope: op1.Scope,
	}
}

// nequalSymbol will check if the two operands are Equal and return a boolean Symbol.
func nequalSymbol(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var ok bool
	err, ok = Equal(op1, op2)
	return nil, &data.Symbol{
		Value: !ok,
		Type:  data.Boolean,
		Scope: op1.Scope,
	}
}

// oFunc is a dummy function which exists within the operatorTable to represent an operation which cannot be made.
func oFunc(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return nil, nil
}

// op1 returns op1.
func op1(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return nil, op1
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

// number evaluates all operations with number on LHS.
func number(op1 *data.Symbol, op2 *data.Symbol, operator Operator) (err error, result *data.Symbol) {
	var op2Number *data.Symbol
	err, op2Number = Cast(op2, data.Number)
	if err != nil {
		return err, nil
	}

	a := op1.Value.(float64)
	b := op2Number.Value.(float64)
	var c float64
	switch operator {
	case Mul:
		c = a * b
	case Div:
		c = a / b
	case Mod:
		c = math.Mod(a, b)
	case Add:
		c = a + b
	case Sub:
		c = a - b
	default:
		return errors.InvalidOperation.Errorf(operator.String(), op1.Type.String(), op2.Type.String()), nil
	}
	return nil, &data.Symbol{
		Value: c,
		Type:  data.Number,
		Scope: op1.Scope,
	}
}

// muNumber: Multiply Number.
func muNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return number(op1, op2, Mul)
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

// or performs a logical OR after casting the RHS to a Boolean.
func or(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2Bool *data.Symbol
	err, op2Bool = Cast(op2, data.Boolean)
	if err != nil {
		return err, nil
	}
	return nil, &data.Symbol{
		Value: op1.Value.(bool) || op2Bool.Value.(bool),
		Type:  data.Boolean,
		Scope: op1.Scope,
	}
}

// difference denotes a set difference between two objects. Assumes that both inputs are objects.
func difference(a *data.Symbol, b *data.Symbol) (err error, result *data.Symbol) {
	aO := a.Value.(map[string]interface{})
	bO := b.Value.(map[string]interface{})
	cO := make(map[string]interface{})
	for k, v := range aO {
		// If the key does not exist in b then we add the pair to c
		if _, ok := bO[k]; !ok {
			cO[k] = v
		}
	}
	return nil, &data.Symbol{
		Value: cO,
		Type:  data.Object,
		Scope: a.Scope,
	}
}

// diObject: Divide Object. Right associative set difference. op2 - op1. Will Cast rhs to Object first.
func diObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2Object *data.Symbol
	err, op2Object = Cast(op2, data.Object)
	if err != nil {
		return err, nil
	}
	return difference(op2Object, op1)
}

// diNumber: Divide Number. Cast rhs to Number.
func diNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return number(op1, op2, Div)
}

// diBoolean: Divide Boolean. op1 NAND op2. Cast rhs to Boolean first.
func diBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	err, result = and(op1, op2)
	if err != nil {
		return err, nil
	}
	result.Value = !result.Value.(bool)
	return nil, result
}

// moString: Mod String. Casts rhs to Array then performs a string format using the verbs available in Go.
func moString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2Array *data.Symbol
	err, op2Array = Cast(op2, data.Array)
	if err != nil {
		return err, nil
	}
	return nil, &data.Symbol{
		Value: fmt.Sprintf(op1.Value.(string), op2Array.Value.([]interface{})...),
		Type:  data.String,
		Scope: op1.Scope,
	}
}

// moNumber: Mod Number.
func moNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return number(op1, op2, Mod)
}

// moBoolean: Mod Number. Casts rhs to Boolean and performs the material conditional (implies that). op1 -> op2. 
func moBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2Boolean *data.Symbol
	err, op2Boolean = Cast(op1, data.Boolean)
	if err != nil {
		return err, nil
	}
	return nil, &data.Symbol{
		Value: (!op1.Value.(bool)) || op2Boolean.Value.(bool),
		Type:  data.Boolean,
		Scope: op1.Scope,
	}
}

// adObject: Add Object. Will merge the RHS into the left overriding any values with the same key. Casts RHS to Object.
func adObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2Object *data.Symbol
	err, op2Object = Cast(op2, data.Object)
	if err != nil {
		return err, nil
	}

	a := op1.Value.(map[string]interface{})
	b := op2Object.Value.(map[string]interface{})
	c := make(map[string]interface{})
	for k, v := range a { c[k] = v }
	for k, v := range b { c[k] = v }
	return nil, &data.Symbol{
		Value: c,
		Type:  data.Object,
		Scope: op1.Scope,
	}
}

// adArray: Add to Array. Append RHS to new Array.
func adArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	a := op1.Value.([]interface{})
	b := op2.Value
	c := make([]interface{}, len(a) + 1)
	copy(c, a)
	c[len(a)] = b
	return nil, &data.Symbol{
		Value: c,
		Type:  data.Array,
		Scope: op1.Scope,
	}
}

// adString: Add Strings. Casts RHS to String then concatenates the two strings.
func adString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2String *data.Symbol
	err, op2String = Cast(op2, data.String)
	if err != nil {
		return err, nil
	}
	return nil, &data.Symbol{
		Value: op1.Value.(string) + op2String.Value.(string),
		Type:  data.String,
		Scope: op1.Scope,
	}
}

// adNumber: Add Number.
func adNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return number(op1, op2, Add)
}

// suObject: Subtract Objects. Performs a set difference between op1 and op2. op1 - op2.
func suObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2Object *data.Symbol
	err, op2Object = Cast(op2, data.Object)
	if err != nil {
		return err, nil
	}
	return difference(op1, op2Object)
}

// suArray: Subtract elements from Array. Casts RHS to Array. All elements on the LHS Equal to the elements in the RHS
// will be removed. If the element is null then the head of the Array is removed.
func suArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	var op2Array *data.Symbol
	err, op2Array = Cast(op2, data.Array)
	if err != nil {
		return err, nil
	}

	elementsToRemove := make([]int, 0)
	remove := func(i int) {
		elementsToRemove = append(elementsToRemove, i)
	}

	a := op1.Value.([]interface{})
	b := op2Array.Value.([]interface{})
	for _, v := range b {
		if v == nil {
			// Remove first element
			remove(0)
		} else {
			for i, w := range a {
				var ok bool
				err, ok = EqualInterface(v, w)
				if err != nil {
					return err, nil
				}
				if ok {
					remove(i)
				}
			}
		}
	}

	return nil, &data.Symbol{
		Value: slices.RemoveElems(a, elementsToRemove...),
		Type:  data.Array,
		Scope: op1.Scope,
	}
}

// suString: Subtract String. Depending on the type of the RHS the following will happen:
//
// - Number: Remove the last n digits from the String.
//
// - String: Remove all occurrences of the RHS from the LHS.
//
// - Object: Replaces all occurrences of RHS's string keys with String versions of their values.
//
// - Array: Casts each element in the array to a string and removes all occurrences of each from the LHS.
//
// - Default: Casts RHS to string and removes all occurrences of each from the LHS.
func suString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	op1Str := op1.Value.(string)
	var op2Str string

	switch op2.Type {
	case data.Number:
		n := len(op1Str) - int(op2.Value.(float64))
		if n > 0 && n < len(op1Str) {
			i := 0
			for j := range op1Str {
				if i == n {
					op1Str = op1Str[:j]
					break
				}
				i++
			}
		} else {
			// We cannot take off more characters than the length of the string, and we cannot add on more
			return errors.StringManipulationError.Errorf(
				op1Str,
				fmt.Sprintf(
					"cannot remove %d characters from string (len(%s) - %d = %d)",
					int(op2.Value.(float64)),
					op1Str,
					int(op2.Value.(float64)), 
					n,
				),
			), nil
		}
	case data.Object:
		obj := op2.Value.(map[string]interface{})
		pairs := make([]string, len(obj) * 2)
		i := 0
		for k, v := range obj {
			// Construct a new symbol for the value and cast it into a string then add the key and that converted value
			// to the pairs array
			var t data.Type
			err = t.Get(v)
			if err == nil {
				vSym := &data.Symbol{
					Value: v,
					Type:  t,
					Scope: 0,
				}
				err, vSym = Cast(vSym, data.String)
				if err == nil {
					pairs[i] = k
					pairs[i + 1] = vSym.Value.(string)
					i += 2
					continue
				}
			}
			return errors.StringManipulationError.Errorf(op1Str, err.Error()), nil
		}

		replacer := strings.NewReplacer(pairs...)
		op1Str = replacer.Replace(op1Str)
	case data.Array:
		arr := op2.Value.([]interface{})
		pairs := make([]string, len(arr) * 2)
		for i, v := range arr {
			// Construct a new symbol for the element and cast it into a string then add the converted element and an 
			// empty string to the pairs array
			var t data.Type
			err = t.Get(v)
			if err == nil {
				vSym := &data.Symbol{
					Value: v,
					Type:  t,
					Scope: 0,
				}
				err, vSym = Cast(vSym, data.String)
				if err == nil {
					pairs[i * 2] = vSym.Value.(string)
					pairs[i * 2 + 1] = ""
					continue
				}
			}
			return errors.StringManipulationError.Errorf(op1Str, err.Error()), nil
		}

		replacer := strings.NewReplacer(pairs...)
		op1Str = replacer.Replace(op1Str)
	default:
		// We default to Casting op2 to a string. Then we fallthrough to the String case to avoid code duplication.
		var op2StrSym *data.Symbol
		err, op2StrSym = Cast(op2, data.String)
		if err != nil {
			return err, nil
		}
		op2Str = op2StrSym.Value.(string)
		fallthrough
	case data.String:
		if op2.Type == data.String {
			op2Str = op2.Value.(string)
		}
		op1Str = strings.ReplaceAll(op1Str, op2Str, "")
	}

	return nil, &data.Symbol{
		Value: op1Str,
		Type:  data.String,
		Scope: op1.Scope,
	}
}

// suNumber: Subtract Number.
func suNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return number(op1, op2, Sub)
}

// suBoolean: Subtract Boolean. op1 NOR op2. Casts RHS to Boolean first.
func suBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	err, result = or(op1, op2)
	if err != nil {
		return err, nil
	}
	result.Value = !result.Value.(bool)
	return nil, result
}

// boolean operator logic. Convert LHS and RHS to boolean and compare them. This encapsulates the default logic.
func boolean(op1 *data.Symbol, op2 *data.Symbol, operator Operator) (err error, result *data.Symbol) {
	var op1Bool, op2Bool *data.Symbol
	err, op1Bool = Cast(op1, data.Boolean)
	if err != nil {
		return err, nil
	}
	err, op2Bool = Cast(op2, data.Boolean)
	if err != nil {
		return err, nil
	}

	a := op1Bool.Value.(bool)
	b := op2Bool.Value.(bool)
	c := false
	switch operator {
	case Eq:
		c = a == b
	case Ne:
		c = !(a == b)
	case And:
		c = a && b
	case Or:
		c = a || b
	default:
		return errors.InvalidOperation.Errorf(operator.String(), op1.Type.String(), op2.Type.String()), nil
	}
	return nil, &data.Symbol{
		Value: c,
		Type:  data.Boolean,
		Scope: op1.Scope,
	}
}

// comparison operator logic. For Numbers and Strings (the only two directly comparable types). If op1 is not a Number 
// or a String it will check if op1 can be converted to a Number or a String and if so will convert it. op2 will then be
// cast to whatever type op1 now is. Then op1 and op2 will be compared. If either op1 or op2 cannot be cast then we will
// return an error. op1 will be checked to be Number first and then a String.
func comparison(op1 *data.Symbol, op2 *data.Symbol, operator Operator) (err error, result *data.Symbol) {
	op1New := op1
	op2New := op2

	// We first check if op1 is a type which can be compared. If not we check if it can be cast and then cast it.
	if op1.Type != data.Number && op1.Type != data.String {
		if Castable(op1, data.Number) {
			err, op1New = Cast(op1, data.Number)
		} else if Castable(op1, data.String) {
			err, op1New = Cast(op1, data.String)
		} else {
			return errors.InvalidOperation.Errorf(operator.String(), op1.Type.String(), op2.Type.String()), nil
		}

		if err != nil {
			return err, nil
		}
	}

	// Then we check if we can cast op2 to the type of the newly cast op1New
	if Castable(op2, op1New.Type) {
		err, op2New = Cast(op2, op1New.Type)
		if err != nil {
			return err, nil
		}
	}

	// Then depending on whether we are comparing Numbers or Strings
	var cI int
	if op1New.Type == data.Number {
		// We use the math/big library to get either -1, 0, or 1
		cI = big.NewFloat(op1New.Value.(float64)).Cmp(big.NewFloat(op2New.Value.(float64)))
	} else if op1New.Type == data.String {
		// We use the strings.Compare to also get -1, 0, or 1
		cI = strings.Compare(op1New.Value.(string), op2New.Value.(string))
	}

	// Because we have our comparison in an intermediate format we just return true or false given the operator.
	var c bool
	switch operator {
	case Lt:
		c = cI < 0
	case Gt:
		c = cI > 0
	case Lte:
		c = cI < 0 || cI == 0
	case Gte:
		c = cI > 0 || cI == 0
	case Eq:
		c = cI == 0
	case Ne:
		c = cI != 0
	}

	return nil, &data.Symbol{
		Value: c,
		Type:  data.Boolean,
		Scope: op1.Scope,
	}
}

// ltObject: Less Than for Object.
func ltObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lt) }

// ltArray: Less Than for Array.
func ltArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lt) }

// ltString: Less Than for String.
func ltString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lt) }

// ltNumber: Less Than for Number.
func ltNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lt) }

// ltBoolean: Less Than for Boolean.
func ltBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lt) }

// gtObject: Greater Than for Object.
func gtObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gt) }

// gtArray: Greater Than for Array.
func gtArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gt) }

// gtString: Greater Than for String.
func gtString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gt) }

// gtNumber: Greater Than for Number.
func gtNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gt) }

// gtBoolean: Greater Than for Boolean.
func gtBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gt) }

// leObject: Less Than or Equal for Object.
func leObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lte) }

// leArray: Less Than or Equal for Array.
func leArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lte) }

// leString: Less Than or Equal for String.
func leString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lte) }

// leNumber: Less Than or Equal for Number.
func leNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lte) }

// leBoolean: Less Than or Equal for Boolean.
func leBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Lte) }

// geObject: Greater Than or Equal for Object.
func geObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gte) }

// geArray: Greater Than or Equal for Array.
func geArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gte) }

// geString: Greater Than or Equal for String.
func geString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gte) }

// geNumber: Greater Than or Equal for Number.
func geNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gte) }

// geBoolean: Greater Than or Equal for Boolean.
func geBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return comparison(op1, op2, Gte) }

// eqObject: Equal for Object.
func eqObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return equalSymbol(op1, op2) }

// eqArray: Equal for Array.
func eqArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return equalSymbol(op1, op2) }

// eqString: Equal for String.
func eqString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return equalSymbol(op1, op2) }

// eqNumber: Equal for Number.
func eqNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return equalSymbol(op1, op2) }

// eqBoolean: Equal for Boolean.
func eqBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return equalSymbol(op1, op2) }

// eqNull: Equal for Null.
func eqNull(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return equalSymbol(op1, op2) }

// neObject: Not Equal for Object.
func neObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return nequalSymbol(op1, op2) }

// neArray: Not Equal for Array.
func neArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return nequalSymbol(op1, op2) }

// neString: Not Equal for String.
func neString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return nequalSymbol(op1, op2) }

// neNumber: Not Equal for Number.
func neNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return nequalSymbol(op1, op2) }

// neBoolean: Not Equal for Boolean.
func neBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return nequalSymbol(op1, op2) }

// neNull: Not Equal for Null.
func neNull(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return nequalSymbol(op1, op2) }

// anObject: And Object.
func anObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, And) }

// anArray: And Array.
func anArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, And) }

// anString: And String.
func anString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, And) }

// anNumber: And Number.
func anNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, And) }

// anBoolean: And Boolean. Casts rhs to Boolean then performs a logical AND operation.
func anBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return and(op1, op2) }

// anNull: And Null.
func anNull(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, And) }

// orObject: Or Object.
func orObject(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, Or) }

// orArray: Or Array.
func orArray(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, Or) }

// orString: Or String.
func orString(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, Or) }

// orNumber: Or Number.
func orNumber(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, Or) }

// orBoolean: Or Boolean. Performs a logical OR.
func orBoolean(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return or(op1, op2) }

// orNull: Or Null.
func orNull(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) { return boolean(op1, op2, Or) }
