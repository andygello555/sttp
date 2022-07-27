package data

import (
	"encoding/json"
	"github.com/andygello555/src/errors"
	"reflect"
	"strconv"
	"strings"
)

// Heap stores variable values on a stack frame.
type Heap map[string]*Value

// Exists will check whether the variable of the given name is on the heap.
func (h *Heap) Exists(name string) bool {
	_, ok := (*h)[name]
	return ok
}

// Delete will delete the symbol of the given name in the given scope. If scope is negative then the most recent symbol
// will be deleted.
func (h *Heap) Delete(name string) {
	delete(*h, name)
}

// Assign will create a new entry in the heap if the variable does not exist yet. Otherwise, will assign the new value
// and type to the existing symbol. The type of the symbol will be decided by Type.Get
func (h *Heap) Assign(name string, value interface{}, global bool, ro bool) (err error) {
	var t Type
	err = t.Get(value)
	if err != nil {
		return err
	}

	// Check if there is an existing value to set
	existing := h.Get(name)
	if existing != nil {
		// If it is immutable, then we return an error
		if existing.ReadOnly {
			return errors.ImmutableValue.Errorf(errors.GetNullVM(), name)
		}
		existing.Value = value
		existing.Type = t
		existing.Global = global
		existing.ReadOnly = ro
	} else {
		(*h)[name] = &Value{
			Value:    value,
			Type:     t,
			Global:   global,
			ReadOnly: ro,
		}
	}
	return nil
}

// Get will retrieve the symbol of the given name in the given scope. Nil will be returned if the variable of the given
// name does not exist.
func (h *Heap) Get(name string) *Value {
	return (*h)[name]
}

// Value represents a value stored on the Heap. It can take any value that is capable of being marshalled to JSON. The
// Global flag indicates whether to reference the Value on all frames that are added to the stack. The ReadOnly flag
// indicates whether or not the value is mutable.
type Value struct {
	Value    interface{} `json:"value"`
	Type     Type        `json:"type"`
	Global   bool        `json:"global"`
	ReadOnly bool        `json:"readOnly"`
}

func (v *Value) String() string {
	ss, err := json.Marshal(v.Value)
	if err != nil {
		panic(err)
	}
	return string(ss)
}

func (v *Value) Float64() float64            { return v.Value.(float64) }
func (v *Value) Int() int                    { return int(v.Float64()) }
func (v *Value) StringLit() string           { return v.Value.(string) }
func (v *Value) Map() map[string]interface{} { return v.Value.(map[string]interface{}) }
func (v *Value) Array() []interface{}        { return v.Value.([]interface{}) }
func (v *Value) Len() int {
	switch v.Type {
	case Object:
		return len(v.Map())
	case Array:
		return len(v.Array())
	case String:
		return len(v.StringLit())
	default:
		return 0
	}
}

func ConstructSymbol(value interface{}, global bool) (err error, symbol *Value) {
	var jsonVal interface{}
	var t Type

	switch value.(type) {
	case string:
		// If the value is a string we unmarshal it and check the unmarshalled value's type
		if err = json.Unmarshal([]byte(value.(string)), &jsonVal); err != nil {
			// If we cannot unmarshal it we can just wrap it in quotations and treat it as a string
			if err = json.Unmarshal([]byte(strconv.Quote(value.(string))), &jsonVal); err != nil {
				return err, nil
			}
		}

		err = t.Get(jsonVal)
		if err != nil {
			return err, nil
		}
	default:
		// We assume that the value is a pointer to a FunctionDefinition
		jsonVal = value
		t = Function
	}

	return nil, &Value{
		Value:  jsonVal,
		Type:   t,
		Global: global,
	}
}

// Type denotes the type of a value. It is defined by a set of constants.
type Type int

func (t *Type) Get(value interface{}) (err error) {
	if value == nil {
		*t = Null
		return nil
	}

	switch value.(type) {
	case bool:
		*t = Boolean
	case float64:
		*t = Number
	case string:
		*t = String
	case []interface{}:
		*t = Array
	case map[string]interface{}:
		*t = Object
	default:
		// Using reflection we find the name of the value's type to see if it is a FunctionDefinition
		if strings.Contains(reflect.TypeOf(value).String(), "FunctionDefinition") {
			*t = Function
		} else {
			return errors.CannotFindType.Errorf(errors.GetNullVM(), value)
		}
	}
	return nil
}

func (t *Type) String() string {
	return typeNames[*t]
}

const (
	// NoType is used for logic within the Heap referrers.
	NoType Type = iota
	// Object is a standard JSON object.
	Object
	// Array is a standard JSON array.
	Array
	String
	// Number can be either floating point or an integer.
	Number
	Boolean
	// Null is a falsy value that indicates nothing.
	Null
	// Function cannot be stored in a variable per-say but is put on the heap as a symbol. A symbol which has a Function
	// type has a value which points to a FunctionBody struct.
	Function
)

var Types = map[Type]bool{
	NoType:   true,
	Object:   true,
	Array:    true,
	String:   true,
	Number:   true,
	Boolean:  true,
	Null:     true,
	Function: true,
}

var typeNames = map[Type]string{
	NoType:   "<error>",
	Object:   "object",
	Array:    "array",
	String:   "string",
	Number:   "number",
	Boolean:  "bool",
	Null:     "null",
	Function: "function",
}
