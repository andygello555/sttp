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
			path: Path{"json", "hello", 0},
			current: "hello",
			to: true,
			expected: map[string]interface{}{
				"": "hello",
				"hello": []interface{}{true},
			},
			err: nil,
		},
		{
			path: Path{"json", 0, "hello"},
			current: "hello",
			to: true,
			expected: []interface{}{
				map[string]interface{}{
					"hello": true,
				},
				"hello",
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
					"hello": "world",
				},
			},
			to: map[string]interface{}{"hello": "world"},
			expected: []interface{}{
				map[string]interface{}{
					"hello": map[string]interface{}{
						"": "world",
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
	}{
		var equal bool
		err, result := test.path.Set(test.current, test.to)
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
			t.Errorf("error \"%s\" should not have occurred", err.Error())
		} else if !equal {
			t.Errorf("result \"%v\" for testNo: %d does not match the required result: \"%v\"", result, testNo + 1, test.expected)
		}
	}
}
