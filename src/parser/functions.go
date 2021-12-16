package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"reflect"
	"strings"
)

// MarshalJSON is used for marshalling FunctionDefinitions to JSON strings as they appear in the data.Heap. The returned
// byte string is in the format:
//  "function:RAW_JSON_PATH:FUNCTION_DEF_UINTPTR"
func (f *FunctionDefinition) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"function:%s:%d\"", f.JSONPath.String(0), reflect.ValueOf(f).Pointer())), nil
}

// builtins contains all builtins in sttp. All builtins take a list of computed arguments.
var builtins = map[string]func(vm VM, args ...*data.Value) (err error, value *data.Value) {
	"print": func(vm VM, args ...*data.Value) (err error, value *data.Value) {
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