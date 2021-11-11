package eval

import (
	"encoding/json"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"strconv"
)

// eFunc is a dummy function which exists within the castTable to represent a cast which cannot be made.
func eFunc(symbol *data.Symbol) (err error, cast *data.Symbol) {
	return nil, nil
}

// same for the diagonals of the castTable matrix.
func same(symbol *data.Symbol) (err error, cast *data.Symbol) {
	return nil, symbol
}

// str will marshal the JSON value back to a JSON string.
func str(value interface{}) string {
	b, err := json.Marshal(value)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// s marshals the JSON value and then creates a String symbol.
func s(symbol *data.Symbol) (err error, cast *data.Symbol) {
	return err, &data.Symbol{
		Value: str(symbol.Value),
		Type:  data.String,
		Scope: symbol.Scope,
	}
}

// length takes the length of strings, []interface{}, and map[string]interface{}. Returns an error if the value does not
// have one of these as an underlying type.
func length(value interface{}) (err error, n int) {
	switch value.(type) {
	case string:
		n = len(value.(string))
	case []interface{}:
		n = len(value.([]interface{}))
	case map[string]interface{}:
		n = len(value.(map[string]interface{}))
	default:
		err = errors.CannotFindLength.Errorf(value)
	}
	return nil, n
}

// l will take the length of the JSON value. Suitable for strings, arrays and objects.
func l(symbol *data.Symbol) (err error, cast *data.Symbol) {
	var n int
	err, n = length(symbol.Value)
	return err, &data.Symbol{
		Value: n,
		Type:  data.Number,
		Scope: symbol.Scope,
	}
}

// lBool will take the length of the JSON value and convert it to a boolean. Suitable for strings, arrays, and objects.
func lBool(symbol *data.Symbol) (err error, cast *data.Symbol) {
	var n int
	err, n = length(symbol.Value)
	return err, &data.Symbol{
		Value: n > 0,
		Type:  data.Boolean,
		Scope: symbol.Scope,
	}
}

// obSing constructs a singleton object where the key is the string representation of the value of the symbol, and the
// value is null.
func obSing(symbol *data.Symbol) (err error, cast *data.Symbol) {
	ob := make(map[string]interface{})
	ob[str(symbol.Value)] = nil
	return nil, &data.Symbol{
		Value: ob,
		Type:  data.Object,
		Scope: symbol.Scope,
	}
}

// arSing constructs a singleton array where the first and only element is the symbol value.
func arSing(symbol *data.Symbol) (err error, cast *data.Symbol) {
	array := make([]interface{}, 1)
	array[1] = symbol.Value
	return nil, &data.Symbol{
		Value: array,
		Type:  data.Array,
		Scope: symbol.Scope,
	}
}

// obArray: from Object to Array. Will extract keys from Object.
func obArray(symbol *data.Symbol) (err error, cast *data.Symbol) {
	ob := symbol.Value.(map[string]interface{})
	array := make([]interface{}, len(ob))
	i := 0
	for _, v := range ob {
		array[i] = v
		i ++
	}
	return nil, &data.Symbol{
		Value: array,
		Type:  data.Array,
		Scope: symbol.Scope,
	}
}

// arObject: from Array to Object. Will create values from elements where keys are the index of each element.
func arObject(symbol *data.Symbol) (err error, cast *data.Symbol) {
	array := symbol.Value.([]interface{})
	ob := make(map[string]interface{})
	for i, v := range array {
		ob[strconv.Itoa(i)] = v
	}
	return nil, &data.Symbol{
		Value: ob,
		Type:  data.Object,
		Scope: symbol.Scope,
	}
}

// stringTo calls ConstructSymbol on the symbol Value, which is assumed to be a string, and checks whether the returned
// symbol is of type t.
func stringTo(symbol *data.Symbol, to data.Type) (err error, cast *data.Symbol) {
	err, cast = data.ConstructSymbol(symbol.Value.(string), symbol.Scope)
	if cast.Type != to {
		return errors.CannotCast.Errorf(symbol.Type.String(), to.String()), nil
	}
	return nil, cast
}

// stObject: from String to Object.
func stObject(symbol *data.Symbol) (err error, cast *data.Symbol) {
	return stringTo(symbol, data.Object)
}

// stArray: from String to Array.
func stArray(symbol *data.Symbol) (err error, cast *data.Symbol) {
	return stringTo(symbol, data.Array)
}

// nuBoolean: from Number to Boolean. Checks whether the number is greater than 0.
func nuBoolean(symbol *data.Symbol) (err error, cast *data.Symbol) {
	return nil, &data.Symbol{
		Value: symbol.Value.(float64) > 0,
		Type:  data.Boolean,
		Scope: symbol.Scope,
	}
}

// boNumber: from Boolean to Number. If false then 0, otherwise 1.
func boNumber(symbol *data.Symbol) (err error, cast *data.Symbol) {
	n := float64(0)
	if symbol.Value.(bool) {
		n = float64(1)
	}
	return nil, &data.Symbol{
		Value: n,
		Type:  data.Number,
		Scope: symbol.Scope,
	}
}

// nlNumber: from Null to Number. Returns a Number symbol with value 0.
func nlNumber(symbol *data.Symbol) (err error, cast *data.Symbol) {
	return nil, &data.Symbol{
		Value: float64(0),
		Type:  data.Number,
		Scope: symbol.Scope,
	}
}

// nlBoolean: from Null to Boolean. Returns a Boolean symbol with value false.
func nlBoolean(symbol *data.Symbol) (err error, cast *data.Symbol) {
	return nil, &data.Symbol{
		Value: false,
		Type:  data.Boolean,
		Scope: symbol.Scope,
	}
}
