package eval

import (
	"encoding/json"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
)

type Heap map[string][]*Symbol

// Exists will check whether the variable of the given name is on the heap.
func (h *Heap) Exists(name string) bool {
	_, ok := (*h)[name]; return ok
}

// New creates a new symbol for the given variable name, type and scope.
func (h *Heap) New(name string, value interface{}, scope int) {
	var symbol *Symbol
	var err error
	if err, symbol = ConstructSymbol(value, scope); err != nil {
		panic(err)
	}

	var start int
	if !h.Exists(name) {
		// Create the list if it doesn't exist
		(*h)[name] = make([]*Symbol, 0)
		start = 0
	} else {
		start = (*h)[name][len((*h)[name])-1].Scope + 1
	}

	// Append any number of NullSymbol to fill the gap
	for i := start; i < scope; i += 1 {
		(*h)[name] = append((*h)[name], NullSymbol)
	}

	// Append the symbol to the end of the symbol list
	(*h)[name] = append((*h)[name], symbol)
}

// Delete will delete the symbol of the given name in the given scope. If scope is negative then the most recent symbol
// will be deleted. Will return an error if it cannot be found.
func (h *Heap) Delete(name string, scope int) error {
	if !h.Exists(name) {
		return errors.HeapEntryDoesNotExist.Errorf("delete", name, scope, name)
	}

	scopes := len((*h)[name])
	toRemove := scope
	if scope >= 0 {
		// If the scope exceeds the limits of the list or the element at the scope points to the NullSymbol then we will return an error
		if scope >= scopes || (*h)[name][scope] == NullSymbol {
			return errors.HeapScopeDoesNotExist.Errorf("delete", name, scope, scope, name)
		}
	} else {
		// Remove the last element
		toRemove = scopes - 1
	}

	// Set the symbol to the NullSymbol
	(*h)[name][toRemove] = NullSymbol

	// If we have "removed" the last element then we need to iterate through the scope list in reverse deleting each
	// NullSymbol until we get to a regular symbol.
	if toRemove == scopes - 1 {
		var i int
		for i = toRemove; i >= 0; i-- {
			if (*h)[name][i] != NullSymbol {
				break
			}
		}
		(*h)[name] = (*h)[name][:i + 1]
	}

	// If the scope list is now empty we will remove the entry from the heap
	if len((*h)[name]) == 0 {
		delete(*h, name)
	}
	return nil
}

// Assign will create a new entry in the heap if the variable does not exist yet. A new entry in the scope list for the
// variable is created if that scoped symbol does not exist. Otherwise, will assign the new value and type to the
// existing symbol. If scope is negative then the most recent symbol is chosen. If t is NoType then the type will not
// be overridden.
func (h *Heap) Assign(name string, value interface{}, scope int) {
	// If the symbol list for the variable doesn't exist or the scope exceeds the limits of the scope list
	if !h.Exists(name) || scope >= 0 && scope >= len((*h)[name]) {
		h.New(name, value, scope)
	} else {
		override := scope
		if scope < 0 {
			override = len((*h)[name]) - 1
		}

		var symbol *Symbol
		var err error
		if err, symbol = ConstructSymbol(value, override); err != nil {
			panic(err)
		}
		(*h)[name][override] = symbol
	}
}

// Get will retrieve the symbol of the given name in the given scope. If scope is negative then the most recent symbol
// will be fetched. Will return an error if it cannot be found.
func (h *Heap) Get(name string, scope int) (error, *Symbol) {
	if !h.Exists(name) {
		return errors.HeapEntryDoesNotExist.Errorf("get", name, scope, name), nil
	}
	scopes := len((*h)[name])
	get := scope
	if scope < 0 {
		scope = scopes - 1
	} else if scope >= scopes {
		// Scope exceeds the limits of the scope list for the entry
		return errors.HeapScopeDoesNotExist.Errorf("get", name, scope, scope, name), nil
	}

	symbol := (*h)[name][get]
	// If the symbol is the NullSymbol then we will return an error as well as the NullSymbol
	if symbol == NullSymbol {
		return errors.HeapScopeDoesNotExist.Errorf("get", name, scope, scope, name), symbol
	}
	return nil, symbol
}

type Symbol struct {
	Value interface{}
	Type  Type
	Scope int
}

func ConstructSymbol(value interface{}, scope int) (err error, symbol *Symbol) {
	var jsonVal interface{}
	var t Type

	switch value.(type) {
	case string:
		// If the value is a string we unmarshal it and check the unmarshalled value's type
		err = json.Unmarshal([]byte(value.(string)), &jsonVal)
		if err != nil {
			return err, nil
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

	return nil, &Symbol{
		Value: jsonVal,
		Type:  t,
		Scope: scope,
	}
}

// NullSymbol is used to fill the empty gaps between symbols of different scopes. This makes symbol getting within the
// Heap O(1) instead of O(n). However, it will make space complexity worse.
var NullSymbol = &Symbol{
	Value: nil,
	Type:  NoType,
	Scope: -1,
}

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
		*t = NoType
		return errors.CannotFindType.Errorf(value)
	}
	return nil
}

const (
	// NoType is used for logic within the Heap referrers.
	NoType = -1
	// Object is a standard JSON object.
	Object Type = iota
	// Array is a standard JSON array.
	Array
	String
	Number
	Boolean
	Null
	// Function cannot be stored in a variable per-say but is put on the heap as a symbol. A symbol which has a Function
	// type has a value which points to a FunctionBody struct.
	Function
)

var Types = map[Type]bool{
	Object: true,
	Array: true,
	String: true,
	Number: true,
	Boolean: true,
	Null: true,
	Function: true,
}
