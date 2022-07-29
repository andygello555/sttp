package errors

import (
	"fmt"
	"github.com/alecthomas/participle/v2/lexer"
)

// ProtoSttpError is what all errors in sttp should use as a prototype.
type ProtoSttpError struct {
	// errorMethod will be called by Error to implement the error interface.
	errorMethod func() string
	// Type is the name of the error.
	Type        string
	// Subset is the name of the subset this error is contained within.
	Subset      string
	// Pos is the position at which the error occurred. This can be empty.
	Pos         lexer.Position
	// CallStack is the result of calling the CallStack.Value method.
	CallStack   []interface{}
	// FromNullVM is a flag that is set when no VM was supplied when creating the error, and the interpreter fell back
	// to the NullVM. A VM instance must be given to retrieve the Pos that the error occurred.
	FromNullVM  bool
}

func (p ProtoSttpError) Error() string { return p.errorMethod() }

// UpdateVM will update the ProtoSttpError with the given VM. This will get the new position and callstack value and 
// set the corresponding fields within the ProtoSttpError.
func (p ProtoSttpError) UpdateVM(vm VM) {
	p.Pos = vm.GetPos()
	p.CallStack = vm.CallStackValue()
	switch vm.(type) {
	case struct { NullVM }:
		p.FromNullVM = true
	default:
		p.FromNullVM = false
	}
}

type RuntimeError string

const (
	StackOverflow             RuntimeError = "exceeded the maximum number of stack frames (%d)"
	StackUnderFlow            RuntimeError = "exceeded the minimum number of stack frames (%d)"
	CannotFindType            RuntimeError = "cannot find type for value \"%v\""
	CannotCast                RuntimeError = "cannot cast type %s to %s"
	CannotFindLength          RuntimeError = "cannot find length of value \"%v\""
	InvalidOperation          RuntimeError = "cannot carry out operation \"%s\" for %s and %s"
	StringManipulationError   RuntimeError = "error whilst manipulating \"%s\": %s"
	JSONPathError             RuntimeError = "cannot access %s with %s"
	Uncallable                RuntimeError = "cannot call value of type %s"
	MoreArgsThanParams        RuntimeError = "function %s has %d parameters, there were %d arguments provided"
	MethodParamNotOptional    RuntimeError = "method parameter \"%s\" is not optional"
	MethodCallMismatchInBatch RuntimeError = "pointer to result for method call: \"%s\" does not match current method call: \"%s\""
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
	MethodCallMismatchInBatch: "MethodCallMismatchInBatch",
}

// Errorf will return an anonymous struct implementing ProtoSttpError with an error method that returns the format 
// string of the RuntimeError filled with the given values.
func (re RuntimeError) Errorf(vm VM, values... interface{}) error {
	return Errorf(vm, "RuntimeError", runtimeErrorNames[re], string(re), values...)
}

// Errorf constructs a custom ProtoSttpError with the given arguments.
func Errorf(vm VM, subset string, t string, format string, values ... interface{}) error {
	pse := struct { ProtoSttpError }{}
	pse.Subset = subset
	pse.Type = t

	if vm != nil {
		pse.Pos = vm.GetPos()
		pse.CallStack = vm.CallStackValue()
		// Decide whether FromNullVM should be set or not
		switch vm.(type) {
		case struct{ NullVM }:
			pse.FromNullVM = true
		default:
			pse.FromNullVM = false
		}
	}

	pse.errorMethod = func() string {
		main := fmt.Sprintf(format, values...)
		if pse.Pos != (lexer.Position{}) {
			main = fmt.Sprintf("%s: %s", pse.Pos.String(), main)
		}
		return main
	}
	return pse
}

type StructureError string

const (
	ImmutableValue        StructureError = "%s is immutable, cannot write to it"
	BatchWithinBatch      StructureError = "cannot have a batch statement within a batch statement"
	NoTestSuite           StructureError = "no test suite, cannot execute test statement: \"%s\""
	HeapEntryDoesNotExist StructureError = "cannot %s %s (scope: %d), as \"%s\" is not an entry in symbol table"
	HeapScopeDoesNotExist StructureError = "cannot %s %s (scope: %d), as scope: %d does not exist in the scope list for the symbol \"%s\""
	BreakOutsideLoop      StructureError = "break statement is outside of loop"
)

var structureErrorNames = map[StructureError]string{
	ImmutableValue: "ImmutableValue",
	BatchWithinBatch: "BatchWithinBatch",
	NoTestSuite: "NoTestSuite",
	HeapEntryDoesNotExist: "HeapEntryDoesNotExist",
	HeapScopeDoesNotExist: "HeapScopeDoesNotExist",
	BreakOutsideLoop: "BreakOutsideLoop",
}

func (se StructureError) Errorf(vm VM, values... interface{}) error {
	return Errorf(vm, "StructureError", structureErrorNames[se], string(se), values...)
}

type PurposefulError int

const (
	Return PurposefulError = iota
	Throw
	FailedTest
	Break
)

var purposefulErrorName = map[PurposefulError]string{
	Return:     "return statement",
	Throw:      "throw statement",
	FailedTest: "failed test statement",
	Break:      "break statement",
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
//      // The below two keys will only be added to the value if the 
//      "pos": {
//          // The column number that the error occurred on
//          "col": 0.0,
//          // The sttp file in which the error occurred in
//          "filename": "*.sttp",
//          // The line number that the error occurred on
//          "line": 0.0,
//      },
//      // The callstack is converted to an sttp value using CallStack.Value. It only converts the most recent couple 
//      // of stack frames
//      "callstack": [
//          {
//              "parent": {
//                  "pos": {"line": ..., "col": ..., "filename": ...},
//					"function": "FunctionDeclaration pointer",
//					"string": "Pretty printed sttp code",
//              },
//				"caller": {
//					"pos": {"line": ..., "col": ..., "filename": ...},
//					"string": "Pretty printed sttp code",
//				},
//          },
//          ...
//      ]
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
		case FailedTest: fallthrough
		case Return: fallthrough
		case Break:
			return err, true
		default:
			errVal = map[string]interface{} {
				"type": err.(PurposefulError),
				"error": err.(PurposefulError).Error(),
				"subset": "PurposefulError",
			}
		}
	case struct { ProtoSttpError }:
		sttpErr := err.(struct { ProtoSttpError })
		errMap := map[string]interface{} {
			"type": sttpErr.Type,
			"error": sttpErr.Error(),
			"subset": sttpErr.Subset,
		}
		if !sttpErr.FromNullVM {
			errMap["pos"] = map[string]interface{} {
				"line": float64(sttpErr.Pos.Line),
				"col": float64(sttpErr.Pos.Column),
				"filename": sttpErr.Pos.Filename,
			}
			errMap["callstack"] = sttpErr.CallStack
		}
		errVal = errMap
	default:
		errVal = map[string]interface{} {
			"type": "",
			"error": err.Error(),
			"subset": "go",
		}
	}
	return errVal, false
}
