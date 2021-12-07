package data

import (
	"fmt"
	"reflect"
	"testing"
)

func TestIterate(t *testing.T) {
	for testNo, test := range []struct{
		input    *Value
		expected *Iterator
		err      error
	}{
		{
			input: &Value{
				Value:    map[string]interface{}{
					"a": float64(0),
					"g": float64(6),
					"c": float64(2),
					"e": float64(4),
					"b": float64(1),
					"f": float64(5),
					"d": float64(3),
				},
				Type:     Object,
				Global:   false,
				ReadOnly: false,
			},
			expected: &Iterator{
				{&Value{"a", String, false, true}, &Value{float64(0), Number, false, true}},
				{&Value{"b", String, false, true}, &Value{float64(1), Number, false, true}},
				{&Value{"c", String, false, true}, &Value{float64(2), Number, false, true}},
				{&Value{"d", String, false, true}, &Value{float64(3), Number, false, true}},
				{&Value{"e", String, false, true}, &Value{float64(4), Number, false, true}},
				{&Value{"f", String, false, true}, &Value{float64(5), Number, false, true}},
				{&Value{"g", String, false, true}, &Value{float64(6), Number, false, true}},
			},
			err: nil,
		},
		{
			input: &Value{
				Value:    []interface{}{"a", "b", "c"},
				Type:     Array,
				Global:   false,
				ReadOnly: false,
			},
			expected: &Iterator{
				{&Value{float64(0), Number, false, true}, &Value{"a", String, false, true}},
				{&Value{float64(1), Number, false, true}, &Value{"b", String, false, true}},
				{&Value{float64(2), Number, false, true}, &Value{"c", String, false, true}},
			},
			err: nil,
		},
		{
			input: &Value{
				Value:    "abc",
				Type:     String,
				Global:   false,
				ReadOnly: false,
			},
			expected: &Iterator{
				{&Value{float64(0), Number, false, true}, &Value{"a", String, false, true}},
				{&Value{float64(1), Number, false, true}, &Value{"b", String, false, true}},
				{&Value{float64(2), Number, false, true}, &Value{"c", String, false, true}},
			},
			err: nil,
		},
	}{
		err, actual := Iterate(test.input)
		if testing.Verbose() {
			fmt.Printf("Test no: %d - %v\n", testNo + 1, test.input)
		}

		if test.err != nil {
			if err.Error() != test.err.Error() {
				t.Errorf("error \"%s\" for testNo: %d does not match the required error: \"%s\"", err.Error(), testNo + 1, test.err.Error())
			}
		} else if err != nil {
			t.Errorf("error \"%s\" should not have occurred (testNo: %d)", err.Error(), testNo + 1)
		}

		if actual.Len() == test.expected.Len() {
			i := 0
			for actual.Len() > 0 {
				elemActual := actual.Next()
				elemExpected := test.expected.Next()
				if testing.Verbose() {
					fmt.Printf("\t%d: (k: %v, v: %v) vs (k: %v, v: %v)\n", i, elemActual.Key.Value, elemActual.Val.Value, elemExpected.Key.Value, elemExpected.Val.Value)
				}
				if !reflect.DeepEqual(elemActual.Key.Value, elemExpected.Key.Value) || !reflect.DeepEqual(elemActual.Val.Value, elemExpected.Val.Value) {
					t.Errorf("element %d: (k: %v, v: %v) does not match expected: (k: %v, v: %v)", i, elemActual.Key.Value, elemActual.Val.Value, elemExpected.Key.Value, elemExpected.Val.Value)
				}
				i ++
			}
		}
	}
}
