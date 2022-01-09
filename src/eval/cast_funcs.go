package eval

import (
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"strconv"
)

// eFunc is a dummy function which exists within the castTable to represent a cast which cannot be made.
func eFunc(symbol *data.Value) (err error, cast *data.Value) {
	return nil, nil
}

// same for the diagonals of the castTable matrix.
func same(symbol *data.Value) (err error, cast *data.Value) {
	return nil, symbol
}

// s marshals the JSON value and then creates a String data.Value.
func s(symbol *data.Value) (err error, cast *data.Value) {
	return err, &data.Value{
		Value: symbol.String(),
		Type:  data.String,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// length takes the length of strings, []interface{}, and map[string]interface{}. Returns an error if the value does not
// have one of these as an underlying type.
func length(value interface{}) (err error, n float64) {
	var nI int
	switch value.(type) {
	case string:
		nI = len(value.(string))
	case []interface{}:
		nI = len(value.([]interface{}))
	case map[string]interface{}:
		nI = len(value.(map[string]interface{}))
	default:
		err = errors.CannotFindLength.Errorf(value)
	}
	return nil, float64(nI)
}

// l will take the length of the JSON value. Suitable for strings, arrays and objects.
func l(symbol *data.Value) (err error, cast *data.Value) {
	var n float64
	err, n = length(symbol.Value)
	return err, &data.Value{
		Value: n,
		Type:  data.Number,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// lBool will take the length of the JSON value and convert it to a boolean. Suitable for strings, arrays, and objects.
func lBool(symbol *data.Value) (err error, cast *data.Value) {
	var n float64
	err, n = length(symbol.Value)
	return err, &data.Value{
		Value: n > 0,
		Type:  data.Boolean,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// obSing constructs a singleton object where the key is the string representation of the value of the symbol, and the
// value is null.
func obSing(symbol *data.Value) (err error, cast *data.Value) {
	ob := make(map[string]interface{})
	ob[symbol.String()] = nil
	return nil, &data.Value{
		Value: ob,
		Type:  data.Object,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// arSing constructs a singleton array where the first and only element is the symbol value.
func arSing(symbol *data.Value) (err error, cast *data.Value) {
	array := make([]interface{}, 1)
	array[0] = symbol.Value
	return nil, &data.Value{
		Value: array,
		Type:  data.Array,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// obArray: from Object to Array. Will extract keys from Object.
func obArray(symbol *data.Value) (err error, cast *data.Value) {
	ob := symbol.Value.(map[string]interface{})
	array := make([]interface{}, len(ob))
	i := 0
	for _, v := range ob {
		array[i] = v
		i ++
	}

	return nil, &data.Value{
		Value: array,
		Type:  data.Array,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// arObject: from Array to Object. Will create values from elements where keys are the index of each element.
func arObject(symbol *data.Value) (err error, cast *data.Value) {
	array := symbol.Value.([]interface{})
	ob := make(map[string]interface{})
	for i, v := range array {
		ob[strconv.Itoa(i)] = v
	}
	return nil, &data.Value{
		Value: ob,
		Type:  data.Object,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// stringTo calls ConstructSymbol on the symbol Value, which is assumed to be a string, and checks whether the returned
// symbol is of type t.
func stringTo(symbol *data.Value, to data.Type) (err error, cast *data.Value) {
	err, cast = data.ConstructSymbol(symbol.StringLit(), symbol.Global)
	if err != nil || cast.Type != to {
		return errors.CannotCast.Errorf(symbol.Type.String(), to.String()), nil
	}
	return nil, cast
}

// stObject: from String to Object.
func stObject(symbol *data.Value) (err error, cast *data.Value) {
	return stringTo(symbol, data.Object)
}

// stArray: from String to Array.
func stArray(symbol *data.Value) (err error, cast *data.Value) {
	return stringTo(symbol, data.Array)
}

// stNumber: from String to Number. Will try to parse the value to an integer, then a float, and finally will just take
// the length of the string.
func stNumber(symbol *data.Value) (err error, cast *data.Value) {
	str := symbol.StringLit()
	var num interface{}

	// First check if parsable to an integer
	if num, err = strconv.Atoi(str); err != nil {
		// If not then we'll check if parsable to a float64
		if num, err = strconv.ParseFloat(str, 64); err != nil {
			// If not then we'll return the length of the string
			return l(symbol)
		}
	} else {
		// If so then we'll convert the number to a float64
		num = float64(num.(int))
	}

	return nil, &data.Value{
		Value:    num,
		Type:     data.Number,
		Global:   symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// nuBoolean: from Number to Boolean. Checks whether the number is greater than 0.
func nuBoolean(symbol *data.Value) (err error, cast *data.Value) {
	return nil, &data.Value{
		Value: symbol.Value.(float64) > 0,
		Type:  data.Boolean,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// boNumber: from Boolean to Number. If false then 0, otherwise 1.
func boNumber(symbol *data.Value) (err error, cast *data.Value) {
	n := float64(0)
	if symbol.Value.(bool) {
		n = float64(1)
	}
	return nil, &data.Value{
		Value: n,
		Type:  data.Number,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// nlNumber: from Null to Number. Returns a Number symbol with value 0.
func nlNumber(symbol *data.Value) (err error, cast *data.Value) {
	return nil, &data.Value{
		Value: float64(0),
		Type:  data.Number,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}

// nlBoolean: from Null to Boolean. Returns a Boolean symbol with value false.
func nlBoolean(symbol *data.Value) (err error, cast *data.Value) {
	return nil, &data.Value{
		Value: false,
		Type:  data.Boolean,
		Global: symbol.Global,
		ReadOnly: symbol.ReadOnly,
	}
}
