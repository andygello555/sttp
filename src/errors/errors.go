package errors

import (
	"fmt"
)

type SttpError interface {
	Errorf(values... interface{}) error
}

type ProtoSttpError struct {
	errorMethod func() string
	Type        string
	Subset      string
}

func (p ProtoSttpError) Error() string { return p.errorMethod() }

type RuntimeError string

const (
	StackOverflow           RuntimeError = "exceeded the maximum number of stack frames (%d)"
	StackUnderFlow          RuntimeError = "exceeded the minimum number of stack frames (%d)"
	CannotFindType          RuntimeError = "cannot find type for value \"%v\""
	CannotCast              RuntimeError = "cannot cast type %s to %s"
	CannotFindLength        RuntimeError = "cannot find length of value \"%v\""
	InvalidOperation        RuntimeError = "cannot carry out operation \"%s\" for %s and %s"
	StringManipulationError RuntimeError = "error whilst manipulating \"%s\": %s"
	JSONPathError           RuntimeError = "cannot access %s with %s"
	Uncallable              RuntimeError = "cannot call value of type %s"
	MoreArgsThanParams      RuntimeError = "function %s has %d parameters, there were %d arguments provided"
	MethodParamNotOptional  RuntimeError = "method parameter \"%s\" is not optional"
)

// runtimeErrorNames contains the names of each RuntimeError enum value.
var runtimeErrorNames = map[RuntimeError]string{
	StackOverflow: "StackOverflow",
	StackUnderFlow: "StackUnderFlow",
	CannotFindType: "CannotFindType",
	CannotCast: "CannotCast",
	CannotFindLength: "CannotFindLength",
	InvalidOperation: "InvalidOperation",
	StringManipulationError: "StringManipulationError",
	JSONPathError: "JSONPathError",
	Uncallable: "Uncallable",
	MoreArgsThanParams: "MoreArgsThanParams",
	MethodParamNotOptional: "MethodParamNotOptional",
}

func (re RuntimeError) Errorf(values... interface{}) error {
	pse := struct { ProtoSttpError }{}
	pse.errorMethod = func() string { return fmt.Sprintf(string(re), values...) }
	pse.Subset = "RuntimeError"
	pse.Type = runtimeErrorNames[re]
	return pse
}

type StructureError string

const (
	ImmutableValue        StructureError = "%s is immutable, cannot write to it"
	BatchWithinBatch      StructureError = "cannot have a batch statement within a batch statement"
	NoTestSuite           StructureError = "no test suite, cannot execute test statement: \"%s\""
	HeapEntryDoesNotExist StructureError = "cannot %s %s (scope: %d), as \"%s\" is not an entry in symbol table"
	HeapScopeDoesNotExist StructureError = "cannot %s %s (scope: %d), as scope: %d does not exist in the scope list for the symbol \"%s\""
)

var structureErrorNames = map[StructureError]string{
	ImmutableValue: "ImmutableValue",
	BatchWithinBatch: "BatchWithinBatch",
	NoTestSuite: "NoTestSuite",
	HeapEntryDoesNotExist: "HeapEntryDoesNotExist",
	HeapScopeDoesNotExist: "HeapScopeDoesNotExist",
}

func (se StructureError) Errorf(values... interface{}) error {
	pse := struct { ProtoSttpError }{}
	pse.errorMethod = func() string { return fmt.Sprintf(string(se), values...) }
	pse.Subset = "StructureError"
	pse.Type = structureErrorNames[se]
	return pse
}

type PurposefulError int

const (
	Return PurposefulError = iota
	Throw
	FailedTest
)

var purposefulErrorName = map[PurposefulError]string{
	Return:     "return statement",
	Throw:      "throw statement",
	FailedTest: "failed test statement",
}

func (pe PurposefulError) Error() string { return purposefulErrorName[pe] }

// ConstructSttpError constructs a value which can be used within the sttp VM. If the error provided is Throw 
// (PurposefulError) then the returned value is the given user defined error. Otherwise, if the error is any other 
// PurposefulError, then the returned value is:
//  {
//      // The int code of the PurposefulError
//      "type": err.(PurposefulError),
//      // The error description
//      "error": err.(PurposefulError).Error(),
//      // To make sure it matches other types of errors
//      "subset": "PurposefulError",
//  }
// If the error is an anonymous struct that implements the ProtoSttpError then the error is constructed as follows 
// (sttpErr is the asserted error value):
//  {
//      "type": sttpErr.Type,
//      "error": sttpErr.Error(),
//      "subset": sttpErr.Subset,
//  }
// Finally, if the error's underlying type is none of the above, then the error is constructed as follows:
//  {
//      "type": "",
//      "error": err.Error(),
//      "subset": "go",
//  }
func ConstructSttpError(err error, userErr interface{}) (errVal interface{}, ret bool) {
	//fmt.Println(reflect.TypeOf(err).String())
	switch err.(type) {
	case PurposefulError:
		// If the error returned is an errors.Throw error. Then we'll set the errVal to be the result
		switch err.(PurposefulError) {
		case Throw:
			errVal = userErr
		case Return:
			return err, true
		default:
			errVal = map[string]interface{} {
				"type": err.(PurposefulError),
				"error": err.(PurposefulError).Error(),
				"subset": "PurposefulError",
			}
		}
		if err.(PurposefulError) == Throw {
		} else {
			errVal = map[string]interface{} {
				"type": err.(PurposefulError),
				"error": err.(PurposefulError).Error(),
				"subset": "PurposefulError",
			}
		}
	case struct { ProtoSttpError }:
		sttpErr := err.(struct { ProtoSttpError })
		errVal = map[string]interface{} {
			"type": sttpErr.Type,
			"error": sttpErr.Error(),
			"subset": sttpErr.Subset,
		}
	default:
		errVal = map[string]interface{} {
			"type": "",
			"error": err.Error(),
			"subset": "go",
		}
	}
	return errVal, false
}
