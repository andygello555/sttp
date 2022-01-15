package parser

import (
	"container/heap"
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/alecthomas/participle/v2/lexer"
	"io"
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
	GetCallStack() CallStack
	errors.VM
	SetPos(position lexer.Position)
	GetScope() *int
	GetParentStatement() interface{}
	CheckTestResults() bool
	CreateTestResults()
	GetTestResults() TestResults
	GetStdout() io.Writer
	GetStderr() io.Writer
	SetStdout(stdout io.Writer)
	SetStderr(stderr io.Writer)
	GetDebug() (io.Writer, bool)
	WriteDebug(format string, a... interface{})
	GetBatch() (BatchSuite, heap.Interface)
	DeleteBatch()
	CreateBatch(statement *Batch)
	ExecuteBatch()
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

// Result represents a result that can occur for any evaluation within sttp.
type Result interface {
	GetErr() error
	GetValue() *data.Value
}

// BatchResult represents a result that can occur for a batched MethodCall.
type BatchResult interface {
	Result
	GetMethodCall() *MethodCall
}

// BatchSuite represents the suite that is used to execute a Batch statement.
type BatchSuite interface {
	AddWork(method *MethodCall, args... *data.Value)
	Work() int
	GetStatement() *Batch
	Execute(workers int) heap.Interface
}
