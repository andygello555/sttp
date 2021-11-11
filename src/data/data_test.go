package data

import (
	"fmt"
	"reflect"
	"testing"
)

func TestHeap_Assign(t *testing.T) {
	for testNo, test := range []struct{
		toAdd    []*Symbol
		names    []string
		expected Heap
	}{
		{
			toAdd: []*Symbol{
				{
					Value: "null",
					Type:  Null,
					Scope: 0,
				},
				{
					Value: "null",
					Type: Null,
					Scope: 3,
				},
			},
			names: []string{"a", "a"},
			expected: map[string][]*Symbol{
				"a": {
					{
						Value: nil,
						Type: Null,
						Scope: 0,
					},
					NullSymbol,
					NullSymbol,
					{
						Value: nil,
						Type: Null,
						Scope: 3,
					},
				},
			},
		},
		{
			toAdd: []*Symbol{
				{
					Value: "null",
					Type: Null,
					Scope: 0,
				},
				{
					Value: "null",
					Type: Null,
					Scope: 1,
				},
				{
					Value: "null",
					Type: Null,
					Scope: 2,
				},
			},
			names: []string{"a", "b", "c"},
			expected: map[string][]*Symbol{
				"a": {
					{
						Value: nil,
						Type: Null,
						Scope: 0,
					},
				},
				"b": {
					NullSymbol,
					{
						Value: nil,
						Type: Null,
						Scope: 1,
					},
				},
				"c": {
					NullSymbol,
					NullSymbol,
					{
						Value: nil,
						Type: Null,
						Scope: 2,
					},
				},
			},
		},
		{
			toAdd: []*Symbol{
				{
					Value: "null",
					Type: Null,
					Scope: 0,
				},
				{
					Value: "null",
					Type: Null,
					Scope: 2,
				},
				{
					Value: "null",
					Type: Null,
					Scope: 4,
				},
				{
					Value: "\"overridden\"",
					Type: String,
					Scope: 0,
				},
				{
					Value: "\"overridden\"",
					Type: String,
					Scope: 1,
				},
				{
					Value: "[\"overridden\"]",
					Type: Array,
					Scope: -1,
				},
			},
			names: []string{"a", "a", "a", "a", "a", "a"},
			expected: map[string][]*Symbol{
				"a": {
					{
						Value: "overridden",
						Type: String,
						Scope: 0,
					},
					{
						Value: "overridden",
						Type: String,
						Scope: 1,
					},
					{
						Value: nil,
						Type: Null,
						Scope: 2,
					},
					NullSymbol,
					{
						Value: []interface{}{"overridden"},
						Type: Array,
						Scope: 4,
					},
				},
			},
		},
	}{
		h := make(Heap)
		for i, add := range test.toAdd {
			h.Assign(test.names[i], add.Value, add.Scope)
		}
		for k, v := range test.expected {
			if _, ok := h[k]; ok {
				if len(v) == len(h[k]) {
					for i, symbol := range v {
						if symbol.Scope != h[k][i].Scope || symbol.Type != h[k][i].Type || !reflect.DeepEqual(symbol.Value, h[k][i].Value) {
							t.Errorf("%d symbol in scope list: %v does not match expected symbol: %v", i, symbol, h[k][i])
						}
					}
				} else {
					t.Errorf("Scope list for variable: \"%s\" is not the same size as the expected scope list for the same variable (%d vs. %d)", k, len(v), len(h[k]))
				}
			} else {
				t.Errorf("Heap for test %d does not include the scope list for the variable \"%s\"", testNo + 1, k)
			}
		}
	}
}

func TestHeap_Delete(t *testing.T) {
	type args struct {
		name  string
		scope int
	}

	for testNo, test := range []struct{
		startHeap Heap
		toDelete  []args
		expected  Heap
		errors    []error
	}{
		{
			startHeap: Heap{
				"a": {
					{
						Value: "null",
						Type: Null,
						Scope: 0,
					},
					NullSymbol,
					NullSymbol,
					{
						Value: "null",
						Type: Null,
						Scope: 3,
					},
				},
			},
			toDelete: []args{
				{"a", 0},
				{"a", -1},
			},
			expected: Heap{},
			errors: []error{},
		},
		{
			startHeap: Heap{
				// No Heap should ever end up looking like this as NullSymbol references should only be inserted between
				// or before non-null symbols. If this does occur for whatever reason the NullSymbols will be removed.
				"a": {
					{
						Value: "null",
						Type: Null,
						Scope: 0,
					},
					NullSymbol,
					NullSymbol,
					NullSymbol,
				},
			},
			toDelete: []args{
				{"a", -1},
			},
			expected: Heap{
				"a": {
					{
						Value: "null",
						Type: Null,
						Scope: 0,
					},
				},
			},
		},
		{
			startHeap: Heap{
				"a": {
					{
						Value: "null",
						Type: Null,
						Scope: 0,
					},
					{
						Value: "null",
						Type: Null,
						Scope: 1,
					},
				},
			},
			toDelete: []args{
				{"a", -1},
			},
			expected: Heap{
				"a": {
					{
						Value: "null",
						Type: Null,
						Scope: 0,
					},
				},
			},
		},
		{
			startHeap: Heap{
				"a": {
					{
						Value: "null",
						Type: Null,
						Scope: 0,
					},
					NullSymbol,
					{
						Value: "null",
						Type: Null,
						Scope: 2,
					},
				},
			},
			toDelete: []args{
				{"b", 0},
				{"a", 1},
				{"a", 3},
			},
			expected: Heap{
				"a": {
					{
						Value: "null",
						Type: Null,
						Scope: 0,
					},
					NullSymbol,
					{
						Value: "null",
						Type: Null,
						Scope: 2,
					},
				},
			},
			errors: []error{
				fmt.Errorf("cannot delete %s (scope: %d), as \"%s\" is not an entry in symbol table", "b", 0, "b"),
				fmt.Errorf("cannot delete %s (scope: %d), as scope: %d does not exist in the scope list for the symbol \"%s\"", "a", 1, 1, "a"),
				fmt.Errorf("cannot delete %s (scope: %d), as scope: %d does not exist in the scope list for the symbol \"%s\"", "a", 3, 3, "a"),
			},
		},
	}{
		for _, del := range test.toDelete {
			err := test.startHeap.Delete(del.name, del.scope)
			if len(test.errors) > 0 {
				found := false
				for _, check := range test.errors {
					if check.Error() == err.Error() {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Cannot find error: \"%s\" in expected errors", err.Error())
				}
			}
		}

		for k, v := range test.startHeap {
			if _, ok := test.expected[k]; ok {
				if len(v) == len(test.expected[k]) {
					for i, symbol := range v {
						if symbol.Scope != test.expected[k][i].Scope || symbol.Type != test.expected[k][i].Type || symbol.Value != test.expected[k][i].Value {
							t.Errorf("%d symbol in scope list: %v does not match expected symbol: %v", i, symbol, test.expected[k][i])
						}
					}
				} else {
					t.Errorf("Scope list for variable: \"%s\" is not the same size as the expected scope list for the same variable (%d vs. %d)", k, len(v), len(test.expected[k]))
				}
			} else {
				t.Errorf("Heap for test %d does not include the scope list for the variable \"%s\"", testNo + 1, k)
			}
		}
	}
}
