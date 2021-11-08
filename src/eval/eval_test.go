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
