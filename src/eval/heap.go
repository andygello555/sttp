package eval

import (
	"fmt"
)

type Heap map[string][]*Symbol

// Exists will check whether the variable of the given name is on the heap.
func (h *Heap) Exists(name string) bool {
	_, ok := (*h)[name]; return ok
}

// New creates a new symbol for the given variable name, type and scope.
func (h *Heap) New(name string, value interface{}, t Type, scope int) {
	// Check if t is a valid type
	if !Types[t] {
		panic(fmt.Sprintf("type: %d, is not a valid type", t))
	}

	// Create the list if it doesn't exist
	if !h.Exists(name) {
		(*h)[name] = make([]*Symbol, 0)
	}

	if scope != 0 {
		// Append any number of NullSymbol to fill the gap
		for i := (*h)[name][len((*h)[name])-1].Scope + 1; i < scope; i += 1 {
			(*h)[name] = append((*h)[name], NullSymbol)
		}
	}

	// Append the symbol to the end of the symbol list
	(*h)[name] = append((*h)[name], &Symbol{
		Value: value,
		Type:  t,
		Scope: scope,
	})
}

// Delete will delete the symbol of the given name in the given scope. If scope is negative then the most recent symbol
// will be deleted. Will return an error if it cannot be found.
func (h *Heap) Delete(name string, scope int) error {
	if !h.Exists(name) {
		return HeapEntryDoesNotExist.Errorf(name, scope, name)
	}

	scopes := len((*h)[name])
	toRemove := scope
	if scope >= 0 {
		// If the scope exceeds the limits of the list or the element at the scope points to the NullSymbol then we will return an error
		if scope >= scopes || (*h)[name][scope] == NullSymbol {
			return HeapScopeDoesNotExist.Errorf(name, scope, scope, name)
		}
	} else {
		// Remove the last element
		toRemove = scopes - 1
	}

	// We will remove that element using copy
	copy((*h)[name][toRemove:], (*h)[name][toRemove + 1:])
	(*h)[name] = (*h)[name][:scopes- 1]
	return nil
}

// Assign will create a new entry in the heap if the variable does not exist yet. A new entry in the scope list for the
// variable is created if that scoped symbol does not exist. Otherwise, will assign the new value and type to the
// existing symbol. If scope is negative then the most recent symbol is chosen. If t is NoType then the type will not
// be overridden.
func (h *Heap) Assign(name string, value interface{}, t Type, scope int) {
	// If the symbol list for the variable doesn't exist or the scope exceeds the limits of the scope list
	if !h.Exists(name) || scope >= 0 && scope >= len((*h)[name]) {
		h.New(name, value, t, scope)
	} else {
		override := scope
		if scope < 0 {
			override = len((*h)[name])
		}

		symbol := (*h)[name][override]
		symbol.Value = value
		symbol.Scope = override
		if t != NoType {
			symbol.Type = t
		}
	}
}

// Get will retrieve the symbol of the given name in the given scope. If scope is negative then the most recent symbol
// will be fetched. Will return an error if it cannot be found.
func (h *Heap) Get(name string, scope int) (error, *Symbol) {
	if !h.Exists(name) {
		return HeapEntryDoesNotExist.Errorf(name, scope, name), nil
	}
	scopes := len((*h)[name])
	get := scope
	if scope < 0 {
		scope = scopes - 1
	} else if scope >= scopes {
		// Scope exceeds the limits of the scope list for the entry
		return HeapScopeDoesNotExist.Errorf(name, scope, scope, name), nil
	}

	symbol := (*h)[name][get]
	// If the symbol is the NullSymbol then we will return an error as well as the NullSymbol
	if symbol == NullSymbol {
		return HeapScopeDoesNotExist.Errorf(name, scope, scope, name), symbol
	}
	return nil, symbol
}

type Symbol struct {
	Value interface{}
	Type  Type
	Scope int
}

// NullSymbol is used to fill the empty gaps between symbols of different scopes. This makes symbol getting within the
// Heap O(1) instead of O(n).
var NullSymbol = &Symbol{
	Value: nil,
	Type:  NoType,
	Scope: -1,
}

type Type int

const (
	// NoType is used for logic within the Heap referrers
	NoType = -1
	Object Type = iota
	Array
	String
	Number
	Boolean
	Null
)

var Types = map[Type]bool{
	Object: true,
	Array: true,
	String: true,
	Number: true,
	Boolean: true,
	Null: true,
}
