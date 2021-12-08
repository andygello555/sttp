package parser

import (
	"fmt"
	"reflect"
)

// MarshalJSON is used for marshalling FunctionDefinitions to JSON strings as they appear in the data.Heap. The returned
// byte string is in the format:
//  "function:RAW_JSON_PATH:FUNCTION_DEF_UINTPTR"
func (f *FunctionDefinition) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"function:%s:%d\"", f.JSONPath.String(0), reflect.ValueOf(f).Pointer())), nil
} 
