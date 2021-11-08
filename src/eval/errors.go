package eval

import "fmt"

type ErrorError string

const (
	HeapEntryDoesNotExist ErrorError = "cannot %s %s (scope: %d), as \"%s\" is not an entry in symbol table"
	HeapScopeDoesNotExist ErrorError = "cannot %s %s (scope: %d), as scope: %d does not exist in the scope list for the symbol \"%s\""
)

func (ee ErrorError) Errorf(values ... interface{}) error {
	return fmt.Errorf(string(ee), values...)
}
