package errors

import "fmt"

type SttpError interface {
	Errorf(values... interface{}) error
}

func errorf(error string, values... interface{}) error {
	return fmt.Errorf(error, values...)
}

type RuntimeError string

const (
	StackOverflow           RuntimeError = "exceeded the maximum number of stack frames (%d)"
	StackUnderFlow          RuntimeError = "exceeded the minimum number of stack frames (%d)"
	CannotFindType          RuntimeError = "cannot find type for value \"%v\""
	CannotCast              RuntimeError = "cannot cast type %s to %s"
	CannotFindLength        RuntimeError = "cannot find length of value \"%v\""
	InvalidOperation        RuntimeError = "cannot carry out operation \"%s\" for %s and %s"
	StringManipulationError RuntimeError = "error whilst manipulating \"%s\": %s"
	Exception               RuntimeError = "exception: %s, was thrown on %s"
	JSONPathError           RuntimeError = "cannot access %s with %s"
)

func (re RuntimeError) Errorf(values... interface{}) error { return errorf(string(re), values...) }

type StructureError string

const (
	ImmutableValue        StructureError = "%s is immutable, cannot write to it"
	HeapEntryDoesNotExist StructureError = "cannot %s %s (scope: %d), as \"%s\" is not an entry in symbol table"
	HeapScopeDoesNotExist StructureError = "cannot %s %s (scope: %d), as scope: %d does not exist in the scope list for the symbol \"%s\""
)

func (se StructureError) Errorf(values... interface{}) error { return errorf(string(se), values...) }
