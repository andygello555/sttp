package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
)

// ASTNode is implemented by all ASTNodes
type ASTNode interface {
	indentString
	evalNode
}

// VM acts as an interface for the overarching state of the VM used for evaluation of programs.
type VM interface {
	Eval(filename, s string) (err error, result *data.Value)
	GetSymbols() *data.Heap
	GetScope() *int
	GetParentStatement() interface{}
	GetCallStack() CallStack
	GetTestResults() TestResults
}

// CallStack is implemented by the call stack that is used within the VM.
type CallStack interface {
	Call(caller *FunctionCall, current *FunctionDefinition, vm VM, args ...*data.Value) error
	Return(vm VM) (err error, frame Frame)
	Current() Frame
	fmt.Stringer
}

// Frame is an entry on the call stack.
type Frame interface {
	GetCaller()  *FunctionCall
	GetCurrent() *FunctionDefinition
	GetHeap()    *data.Heap
	GetReturn()  *data.Value
}

// TestResults is a list of test results.
type TestResults interface{
	AddTest(node *TestStatement, passed bool)
	GetConfig() Config
	CheckPassed() bool
	indentString
}

// Config is an interface for any config type.
type Config interface {
	Get(name string) interface{}
}
