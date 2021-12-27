package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"reflect"
	"strings"
)

// MarshalJSON is used for marshalling FunctionDefinitions to JSON strings as they appear in the data.Heap. The returned
// byte string is in the format:
//  "function:RAW_JSON_PATH:FUNCTION_DEF_UINTPTR"
func (f *FunctionDefinition) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"function:%s:%d\"", f.JSONPath.String(0), reflect.ValueOf(f).Pointer())), nil
}

// BuiltinFunction denotes the signature of each builtin function in the builtins table.
type BuiltinFunction func(vm VM, uncomputedArgs... *Expression) (err error, value *data.Value)

// String is used for marshalling BuiltinFunctions to strings and JSON strings. The name of the builtin will be found 
// by using reflection to get the pointer to the builtin function. The returned string is in the format:
//  "builtin:NAME:BUILTIN_FUNCTION_UINTPTR"
func (b BuiltinFunction) String() string {
	ptr := reflect.ValueOf(b).Pointer()
	var name string; var v BuiltinFunction
	for name, v = range builtins {
		currPtr := reflect.ValueOf(v).Pointer()
		if ptr == currPtr {
			break
		}
	}
	return fmt.Sprintf("\"builtin:%s:%d\"", name, ptr)
}

func (b BuiltinFunction) MarshalJSON() ([]byte, error) {
	return []byte(b.String()), nil
}

func computeArgs(vm VM, uncomputedArgs... *Expression) (err error, args []*data.Value) {
	// Evaluate arguments and create a list of args
	args = make([]*data.Value, len(uncomputedArgs))
	for i, arg := range uncomputedArgs {
		if err, args[i] = arg.Eval(vm); err != nil {
			return err, nil
		}
	}
	return nil, args
}

// builtins contains all builtins in sttp. All builtins take a list of uncomputed arguments. These are uncomputed as 
// there might be special use cases.
var builtins = map[string]BuiltinFunction {
	"print": func(vm VM, uncomputedArgs ...*Expression) (err error, value *data.Value) {
		var args []*data.Value
		if err, args = computeArgs(vm, uncomputedArgs...); err != nil {
			return err, nil
		}

		var b strings.Builder
		for i, arg := range args {
			if arg.Type == data.String {
				b.WriteString(arg.Value.(string))
			} else {
				b.WriteString(arg.String())
			}
			if i != len(args) - 1 {
				b.WriteString(" ")
			}
		}
		_, err = fmt.Fprintf(vm.GetStdout(), "%s\n", b.String())

		return err, &data.Value{
			Value:    nil,
			Type:     data.Null,
			Global:   false,
			ReadOnly: false,
		}
	},
	"free": func(vm VM, uncomputedArgs ...*Expression) (err error, value *data.Value) {
		// For each uncomputed arg, we will find if there is a JSONPath factor terminal contained in the left-hand side
		// of the expression.
		for _, uncomputedArg := range uncomputedArgs {
			// We look down the left-hand side of the expression tree and see if the terminal factor is a JSONPath. We 
			// will only consider this JSONPath, even 
			var t term = uncomputedArg
			var e evalNode
			for {
				// We first get the evalNode from the left() method of the current argument.
				e = t.left()
				if t == nil {
					// We return an error if we cannot continue down the left path before finding a JSONPath terminal.
					return errors.InvalidOperation.Errorf("builtin:delete", fmt.Sprintf("non-JSONPath value: \"%s\"", uncomputedArg.String(0)), "delete"), nil
				} else {
					// We do a type switch for the evalNode to find out if the underlying type is a JSONPath. If so we
					// can stop iteration. Otherwise, we cast the evalNode to a term interface and assign the t var.
					stop := false
					switch e.(type) {
					case *JSONPath:
						stop = true
					case *Null, *Boolean, *JSON, *FunctionCall, *MethodCall, *Expression, *struct { protoEvalNode }:
						// We found an expression terminal/factor before finding a JSONPath.
						return errors.InvalidOperation.Errorf("builtin:delete", fmt.Sprintf("non-JSONPath value: \"%s\"", e.(ASTNode).String(0)), "delete"), nil
					default:
						break
					}
					if stop {
						break
					}
					t = e.(term)
				}
			}

			// e should be a *JSONPath value, so we convert it into a Path.
			var path Path
			if err, path = e.(*JSONPath).Convert(vm); err != nil {
				return err, nil
			}

			// We get the name of the variable as well as the value of the variable.
			variableName := path[0].(string)
			heap := vm.GetCallStack().Current().GetHeap()

			if len(path) > 1 {
				variableVal := heap.Get(variableName)
				if variableVal == nil {
					variableVal = &data.Value{
						Value:  nil,
						Type:   data.Null,
					}
				}

				// We set the value at the path to nil if we have more than element in the path.
				if err, variableVal.Value = path.Set(variableVal.Value, nil); err != nil {
					return err, nil
				}
				if err = heap.Assign(variableName, variableVal.Value, variableVal.Global, variableVal.ReadOnly); err != nil {
					return err, nil
				}
			} else {
				// If we only have one element then we will delete the Value from the heap.
				heap.Delete(variableName)
			}
		}
		return nil, &data.Value{
			Value:    nil,
			Type:     data.Null,
			Global:   false,
			ReadOnly: false,
		}
	},
}

// CheckBuiltin will check if the function of the given name exists as a builtin.
func CheckBuiltin(name string) bool {
	_, ok := builtins[name]
	return ok
}

// GetBuiltin will return the builtin function encapsulated in a data.Value.
func GetBuiltin(name string) *data.Value {
	if CheckBuiltin(name) {
		return &data.Value{
			Value:    builtins[name],
			Type:     data.Function,
			Global:   true,
			ReadOnly: true,
		}
	}
	return nil
}