package eval

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"testing"
)

func TestCompute(t *testing.T) {
	for testNo, test := range []struct{
		op1 *data.Symbol
		op2 *data.Symbol
		operator Operator
		result *data.Symbol
		err error
	}{
		// Unsupported operation
		{
			op1: &data.Symbol{
				Value: nil,
				Type:  data.Object,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: nil,
				Type:  data.Null,
				Scope: 0,
			},
			operator: Mul,
			result: &data.Symbol{
				Value: nil,
				Type:  0,
				Scope: 0,
			},
			err: errors.InvalidOperation.Errorf("*", "object", "null"),
		},

		// Array manipulation

		{
			op1: &data.Symbol{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: 4,
				Type:  data.Number,
				Scope: 0,
			},
			operator: Add,
			result: &data.Symbol{
				Value: []interface{}{1, 2, 3, 4},
				Type:  data.Array,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: map[string]interface{}{"a": 1, "b": 2},
				Type:  data.Object,
				Scope: 0,
			},
			operator: Add,
			result: &data.Symbol{
				Value: []interface{}{1, 2, 3, map[string]interface{}{"a": 1, "b": 2}},
				Type:  data.Array,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: []interface{}{[]interface{}{1, 2, 3}, 2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: nil,
				Type:  data.Null,
				Scope: 0,
			},
			operator: Sub,
			result: &data.Symbol{
				Value: []interface{}{2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: []interface{}{[]interface{}{1, 2, 3}, 2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: []interface{}{[]interface{}{1, 2, 3}},
				Type:  data.Array,
				Scope: 0,
			},
			operator: Sub,
			result: &data.Symbol{
				Value: []interface{}{2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: []interface{}{nil, 2},
				Type:  data.Array,
				Scope: 0,
			},
			operator: Sub,
			result: &data.Symbol{
				Value: []interface{}{3},
				Type:  data.Array,
				Scope: 0,
			},
			err: nil,
		},

		// Object manipulation

		{
			op1: &data.Symbol{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			operator: Div,
			result: &data.Symbol{
				Value: map[string]interface{}{"0": 1},
				Type:  data.Object,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			operator: Sub,
			result: &data.Symbol{
				Value: map[string]interface{}{"3": 3},
				Type:  data.Object,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: "{\"hello\":\"world\"}",
				Type:  data.String,
				Scope: 0,
			},
			operator: Add,
			result: &data.Symbol{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3, "hello": "world"},
				Type:  data.Object,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3},
				Type:  data.Object,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: 4,
				Type:  data.Number,
				Scope: 0,
			},
			operator: Add,
			result: &data.Symbol{
				Value: map[string]interface{}{"1": 1, "2": 2, "3": 3, "4": nil},
				Type:  data.Object,
				Scope: 0,
			},
			err: nil,
		},

		// String manipulation

		{
			op1: &data.Symbol{
				Value: "abc",
				Type:  data.String,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: map[string]interface{}{"a": float64(1), "b": float64(2), "c": float64(3)},
				Type:  data.Object,
				Scope: 0,
			},
			operator: Sub,
			result: &data.Symbol{
				Value: "123",
				Type:  data.String,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: "moomoo cow is here",
				Type:  data.String,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: []interface{}{"moo", "is"},
				Type:  data.Array,
				Scope: 0,
			},
			operator: Sub,
			result: &data.Symbol{
				Value: " cow  here",
				Type:  data.String,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: "123456",
				Type:  data.String,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: float64(3),
				Type:  data.Number,
				Scope: 0,
			},
			operator: Sub,
			result: &data.Symbol{
				Value: "123",
				Type:  data.String,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: "moomoocow",
				Type:  data.String,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: "moo",
				Type:  data.String,
				Scope: 0,
			},
			operator: Sub,
			result: &data.Symbol{
				Value: "cow",
				Type:  data.String,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: "is null nullable?",
				Type:  data.String,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: nil,
				Type:  data.Null,
				Scope: 0,
			},
			operator: Sub,
			result: &data.Symbol{
				Value: "is  able?",
				Type:  data.String,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: "Result is: ",
				Type:  data.String,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			operator: Add,
			result: &data.Symbol{
				Value: "Result is: [1,2,3]",
				Type:  data.String,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: "Result is: [%d, %d, %d]",
				Type:  data.String,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: []interface{}{1, 2, 3},
				Type:  data.Array,
				Scope: 0,
			},
			operator: Mod,
			result: &data.Symbol{
				Value: "Result is: [1, 2, 3]",
				Type:  data.String,
				Scope: 0,
			},
			err: nil,
		},
		{
			op1: &data.Symbol{
				Value: "Result is: [%d]",
				Type:  data.String,
				Scope: 0,
			},
			op2: &data.Symbol{
				Value: map[string]interface{}{"1": 1},
				Type:  data.Object,
				Scope: 0,
			},
			operator: Mod,
			result: &data.Symbol{
				Value: "Result is: [1]",
				Type:  data.String,
				Scope: 0,
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
