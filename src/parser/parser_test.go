package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"testing"
)

func TestPath_Set(t *testing.T) {
	for testNo, test := range []struct{
		path     Path
		current  interface{}
		to       interface{}
		expected interface{}
		err      error
	}{
		{
			path: Path{"json", "hello", "world"},
			current: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": false,
				},
			},
			to: true,
			expected: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": true,
				},
			},
			err: nil,
		},
		{
			path: Path{"json", "hello", "world", 0},
			current: nil,
			to: true,
			expected: map[string]interface{}{
				"hello": map[string]interface{}{
					"world": []interface{}{true},
				},
			},
			err: nil,
		},
		{
			path: Path{"json", 0, 0, 0},
			current: nil,
			to: true,
			expected: []interface{}{
				[]interface{}{
					[]interface{}{
						true,
					},
				},
			},
			err: nil,
		},
		{
			path: Path{"json", 0, 1, 2, 3},
			current: nil,
			to: true,
			expected: []interface{}{
				[]interface{}{
					nil,
					[]interface{}{
						nil,
						nil,
						[]interface{}{
							nil,
							nil,
							nil,
							true,
						},
					},
				},
			},
			err: nil,
		},
		{
			path:    Path{"json", "hello", 0},
			current: 3.142,
			to:      true,
			expected: map[string]interface{}{
				"": 3.142,
				"hello": []interface{}{true},
			},
			err: nil,
		},
		{
			path: Path{"json", 0, "hello"},
			current: 3.142,
			to: true,
			expected: []interface{}{
				map[string]interface{}{
					"hello": true,
				},
				3.142,
			},
			err: nil,
		},
		{
			path: Path{"json", 0, 1},
			current: map[string]interface{}{
				"b": 2,
				"a": 1,
				"c": 3,
			},
			to: true,
			expected: map[string]interface{}{
				"b": 2,
				"a": []interface{}{nil, true, 1},
				"c": 3,
			},
			err: nil,
		},
		{
			path: Path{"json", 0, "hello", "world"},
			current: []interface{}{
				map[string]interface{}{
					"hello": 3.142,
				},
			},
			to: map[string]interface{}{"hello": "world"},
			expected: []interface{}{
				map[string]interface{}{
					"hello": map[string]interface{}{
						"": 3.142,
						"world": map[string]interface{}{
							"hello": "world",
						},
					},
				},
			},
			err: nil,
		},
		{
			path: Path{"json", 0},
			current: []interface{}{false},
			to: true,
			expected: []interface{}{true},
			err: nil,
		},
		{
			path: Path{"json"},
			current: "hello",
			to: "world",
			expected: "world",
			err: nil,
		},
		{
			path: Path{"json", "0"},
			current: []interface{}{false},
			to: true,
			expected: nil,
			err: fmt.Errorf("cannot access array with property"),
		},
		{
			path: Path{"json", 3},
			current: map[string]interface{}{
				"b": 2,
				"a": 1,
				"c": 3,
			},
			to: true,
			expected: nil,
			err: fmt.Errorf("cannot access object with index 3"),
		},
		{
			path: Path{"json", 0, -5},
			current: map[string]interface{}{
				"b": 2,
				"a": 1,
				"c": 3,
			},
			to: true,
			expected: nil,
			err: fmt.Errorf("cannot access non-object/array type with a negative index (-5)"),
		},
		{
			path: Path{"json", 0, -1},
			current: map[string]interface{}{
				"b": 2,
				"a": []interface{}{1, 2, 3},
				"c": 3,
			},
			to: true,
			expected: map[string]interface{}{
				"b": 2,
				"a": []interface{}{1, 2, true},
				"c": 3,
			},
			err: nil,
		},
		{
			path: Path{"json", 0, -3},
			current: map[string]interface{}{
				"b": 2,
				"a": []interface{}{1, 2, 3},
				"c": 3,
			},
			to: true,
			expected: map[string]interface{}{
				"b": 2,
				"a": []interface{}{true, 2, 3},
				"c": 3,
			},
			err: nil,
		},
		{
			path: Path{"json", 0, -4},
			current: map[string]interface{}{
				"b": 2,
				"a": []interface{}{1, 2, 3},
				"c": 3,
			},
			to: true,
			expected: nil,
			err: fmt.Errorf("cannot access array with negative index that is out of array bounds (-4)"),
		},
		{
			path: Path{"json", "hello", 3},
			current: map[string]interface{}{
				"hello": "world",
			},
			to: "egg",
			expected: map[string]interface{}{
				"hello": "woreggd",
			},
			err: nil,
		},
		{
			path: Path{"json", "hello", 4},
			current: map[string]interface{}{
				"hello": "world",
			},
			to: "egg",
			expected: map[string]interface{}{
				"hello": "worlegg",
			},
			err: nil,
		},
		{
			path: Path{"json", "hello", -5},
			current: map[string]interface{}{
				"hello": "world",
			},
			to: "egg",
			expected: map[string]interface{}{
				"hello": "eggorld",
			},
			err: nil,
		},
		{
			path: Path{"json", "hello", -6},
			current: map[string]interface{}{
				"hello": "world",
			},
			to: "egg",
			expected: nil,
			err: fmt.Errorf("cannot access string with negative index that is out of string bounds (-6)"),
		},
		{
			path: Path{"json", "hello", 6},
			current: map[string]interface{}{
				"hello": "world",
			},
			to: "egg",
			expected: map[string]interface{}{
				"hello": "world egg",
			},
			err: nil,
		},
		{
			path: Path{"json", "hello", 6},
			current: map[string]interface{}{
				"hello": "world",
			},
			to: "egg",
			expected: map[string]interface{}{
				"hello": "world egg",
			},
			err: nil,
		},
		{
			path: Path{"json", "hello", 6, 0},
			current: map[string]interface{}{
				"hello": "world",
			},
			to: "egg",
			expected: map[string]interface{}{
				"hello": "world egg",
			},
			err: nil,
		},
		{
			path: Path{"json", "hello", 6, 1},
			current: map[string]interface{}{
				"hello": "world",
			},
			to: "egg",
			expected: map[string]interface{}{
				"hello": "world  egg",
			},
			err: nil,
		},
	}{
		var equal bool
		// Parsing in nil for VM parameter as we don't test filter blocks here.
		err, result := test.path.Set(nil, test.current, test.to)
		// Check if the actual result is equal to the expected result only if there is no error.
		if err == nil {
			err, equal = eval.EqualInterface(result, test.expected)
		}

		if testing.Verbose() && result != nil {
			fmt.Printf("%d: set %v to %v = %v\n", testNo + 1, test.path, test.to, result)
		}

		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("error \"%s\" for testNo: %d does not match the required error: \"%s\"", err.Error(), testNo + 1, test.err.Error())
			}
		} else if err != nil {
			t.Errorf("error \"%s\" should not have occurred (testNo: %d)", err.Error(), testNo + 1)
		} else if !equal {
			t.Errorf("result \"%v\" for testNo: %d does not match the required result: \"%v\"", result, testNo + 1, test.expected)
		}
	}
}
