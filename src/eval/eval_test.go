package eval

import "testing"

func TestHeap_Assign(t *testing.T) {
	for testNo, test := range []struct{
		toAdd    []*Symbol
		names    []string
		expected Heap
	}{
		{
			toAdd: []*Symbol{
				{
					Value: nil,
					Type:  Object,
					Scope: 0,
				},
				{
					Value: nil,
					Type: Boolean,
					Scope: 3,
				},
			},
			names: []string{"a", "a"},
			expected: map[string][]*Symbol{
				"a": {
					{
						Value: nil,
						Type: Object,
						Scope: 0,
					},
					NullSymbol,
					NullSymbol,
					{
						Value: nil,
						Type: Boolean,
						Scope: 3,
					},
				},
			},
		},
		{
			toAdd: []*Symbol{
				{
					Value: nil,
					Type: Object,
					Scope: 0,
				},
				{
					Value: nil,
					Type: String,
					Scope: 1,
				},
				{
					Value: nil,
					Type: Array,
					Scope: 2,
				},
			},
			names: []string{"a", "b", "c"},
			expected: map[string][]*Symbol{
				"a": {
					{
						Value: nil,
						Type: Object,
						Scope: 0,
					},
				},
				"b": {
					NullSymbol,
					{
						Value: nil,
						Type: String,
						Scope: 1,
					},
				},
				"c": {
					NullSymbol,
					NullSymbol,
					{
						Value: nil,
						Type: Array,
						Scope: 2,
					},
				},
			},
		},
		{
			toAdd: []*Symbol{
				{
					Value: nil,
					Type: String,
					Scope: 0,
				},
				{
					Value: nil,
					Type: Object,
					Scope: 2,
				},
				{
					Value: nil,
					Type: Number,
					Scope: 4,
				},
				{
					Value: "overridden",
					Type: Number,
					Scope: 0,
				},
				{
					Value: "overridden",
					Type: Null,
					Scope: 1,
				},
				{
					Value: "overridden",
					Type: Array,
					Scope: -1,
				},
			},
			names: []string{"a", "a", "a", "a", "a", "a"},
			expected: map[string][]*Symbol{
				"a": {
					{
						Value: "overridden",
						Type: Number,
						Scope: 0,
					},
					{
						Value: "overridden",
						Type: Null,
						Scope: 1,
					},
					{
						Value: nil,
						Type: Object,
						Scope: 2,
					},
					NullSymbol,
					{
						Value: "overridden",
						Type: Array,
						Scope: 4,
					},
				},
			},
		},
	}{
		h := make(Heap)
		for i, add := range test.toAdd {
			h.Assign(test.names[i], add.Value, add.Type, add.Scope)
		}
		for k, v := range test.expected {
			if _, ok := h[k]; ok {
				if len(v) == len(h[k]) {
					for i, symbol := range v {
						if symbol.Scope != h[k][i].Scope || symbol.Type != h[k][i].Type || symbol.Value != h[k][i].Value {
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

	for _, test := range []struct{
		startHeap Heap
		toDelete  []args
		expected  Heap
		errors    []error
	}{
		{
			startHeap: Heap{
				"a": {
					{
						Value: nil,
						Type: String,
						Scope: 0,
					},
					NullSymbol,
					NullSymbol,
					{
						Value: nil,
						Type: Object,
						Scope: 3,
					},
				},
			},
			toDelete: []args{
				{"a", 0},
				{"a", -1},
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
		//for k, v := range test.expected {
		//	if _, ok := h[k]; ok {
		//		if len(v) == len(h[k]) {
		//			for i, symbol := range v {
		//				if symbol.Scope != h[k][i].Scope || symbol.Type != h[k][i].Type || symbol.Value != h[k][i].Value {
		//					t.Errorf("%d symbol in scope list: %v does not match expected symbol: %v", i, symbol, h[k][i])
		//				}
		//			}
		//		} else {
		//			t.Errorf("Scope list for variable: \"%s\" is not the same size as the expected scope list for the same variable (%d vs. %d)", k, len(v), len(h[k]))
		//		}
		//	} else {
		//		t.Errorf("Heap for test %d does not include the scope list for the variable \"%s\"", testNo + 1, k)
		//	}
		//}
	}
}
