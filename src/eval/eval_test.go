package eval

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"testing"
)

func TestCompute(t *testing.T) {
	for testNo, test := range []struct{
		op1 *data.Value
		op2 *data.Value
		operator Operator
		result *data.Value
		err error
	}{
		// Unsupported operation
		{
			op1: &data.Value{
				Value: nil,
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: nil,
				Type:  data.Null,
				Global: false,
			},
			operator: Mul,
			result: &data.Value{
				Value: nil,
				Type:  0,
				Global: false,
			},
			err: errors.InvalidOperation.Errorf("*", "object", "null"),
		},

		// Array manipulation

		{
			op1: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: 4,
				Type:  data.Number,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: []interface{}{1, 2, 3, 4},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: map[string]interface{}{"a": 1, "b": 2},
				Type:  data.Object,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: []interface{}{1, 2, 3, map[string]interface{}{"a": 1, "b": 2}},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{[]interface{}{1, 2, 3}, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: nil,
				Type:  data.Null,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: []interface{}{2, 3},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{[]interface{}{1, 2, 3}, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{[]interface{}{1, 2, 3}},
				Type:  data.Array,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: []interface{}{2, 3},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{nil, 2},
				Type:  data.Array,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: []interface{}{3},
				Type:  data.Array,
				Global: false,
			},
			err: nil,
		},

		// Object manipulation

		{
			op1: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			operator: Div,
			result: &data.Value{
				Value: map[string]interface{}{"0": 1},
				Type:  data.Object,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: map[string]interface{}{"3": 3},
				Type:  data.Object,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: "{\"hello\":\"world\"}",
				Type:  data.String,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3, "hello": "world"},
				Type:  data.Object,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Global: false,
			},
			op2: &data.Value{
				Value: 4,
				Type:  data.Number,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3, "4": nil},
				Type:  data.Object,
				Global: false,
			},
			err: nil,
		},

		// String manipulation

		{
			op1: &data.Value{
				Value: "abc",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: map[string]interface{}{"a": float64(1), "b": float64(2), "c": float64(3)},
				Type:  data.Object,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: "123",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "moomoo cow is here",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{"moo", "is"},
				Type:  data.Array,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: " cow  here",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "123456",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: float64(3),
				Type:  data.Number,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: "123",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "moomoocow",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: "moo",
				Type:  data.String,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: "cow",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "is null nullable?",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: nil,
				Type:  data.Null,
				Global: false,
			},
			operator: Sub,
			result: &data.Value{
				Value: "is  able?",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "Result is: ",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			operator: Add,
			result: &data.Value{
				Value: "Result is: [1,2,3]",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "Result is: [%d, %d, %d]",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Global: false,
			},
			operator: Mod,
			result: &data.Value{
				Value: "Result is: [1, 2, 3]",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
		{
			op1: &data.Value{
				Value: "Result is: [%d]",
				Type:  data.String,
				Global: false,
			},
			op2: &data.Value{
				Value: map[string]interface{}{"1": 1},
				Type:  data.Object,
				Global: false,
			},
			operator: Mod,
			result: &data.Value{
				Value: "Result is: [1]",
				Type:  data.String,
				Global: false,
			},
			err: nil,
		},
	}{
		var ok bool
		err, result := Compute(test.operator, test.op1, test.op2)
		// Check if the actual result is Equal to the expected result only if there is no error.
		if err == nil {
			err, ok = Equal(result, test.result)
		}

		if testing.Verbose() && result != nil {
			fmt.Printf("%d: %v %s %v = %v\n", testNo + 1, test.op1.String(), test.operator.String(), test.op2.String(), result.String())
		}

		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("error \"%s\" for testNo: %d does not match the required error: \"%s\"", err.Error(), testNo + 1, test.err.Error())
			}
		} else if !ok {
			t.Errorf("result \"%v\" for testNo: %d does not match the required result: \"%v\"", result, testNo + 1, test.result)
		}
	}
}
