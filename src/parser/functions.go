package parser

import (
	"fmt"
	"reflect"
)

func (f *FunctionDefinition) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("\"function:%s:%d\"", f.JSONPath.String(0), reflect.ValueOf(f).Pointer())), nil
} 
