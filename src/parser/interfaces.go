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
	IndentString
	evalNode
}

// VM acts as an interface for the overarching state of the VM used for evaluation of programs.
type VM interface {
	// Eval will parse and evaluate the given sttp script as a string. Filename, can also be given to give errors more 
	// context.
	Eval(filename, s string) (err error, result *data.Value)
	// GetCallStack will return the parser.CallStack bound to the VM.
	GetCallStack() CallStack
	// errors.VM also implement the GetPos and CallStackValue instance methods.
	errors.VM
	// SetPos will set the position state to the given position. This is used to give context to the VM, so should be 
	// used whenever possible (within reason) to give errors the best possible context. 
	SetPos(position lexer.Position)
	// GetScope will return a pointer to an integer representing the current scope of execution. This is incremented 
	// whenever a FunctionCall is started to be evaluated, and decremented whenever a FunctionCall has stopped 
	// evaluation.
	GetScope() *int
	// CheckTestResults checks whether there are TestResults to add a result to.
	CheckTestResults() bool
	// CreateTestResults will create a new TestResults container for test results.
	CreateTestResults()
	// GetTestResults will return the TestResults within the VM state.
	GetTestResults() TestResults
	// GetStdout will return the io.Writer for the currently set stdout file.
	GetStdout() io.Writer
	// GetStderr will return the io.Writer for the currently set stderr file.
	GetStderr() io.Writer
	// SetStdout will set the stdout io.Writer to the one provided.
	SetStdout(stdout io.Writer)
	// SetStderr will set the stderr io.Writer to the one provided.
	SetStderr(stderr io.Writer)
	// GetDebug will return the io.Writer for the file used for debugging and whether that file is ioutil.Discard.
	GetDebug() (io.Writer, bool)
	// WriteDebug will write the format string and its arguments to the debug io.Writer.
	WriteDebug(format string, a... interface{})
	// GetBatch will return the BatchSuite as well as the batch results if there are any.
	GetBatch() (BatchSuite, heap.Interface)
	// DeleteBatch will set both the BatchSuite and the batch results to be nil. Forcing them to be garbage collected.
	DeleteBatch()
	// CreateBatch will create a new BatchSuite for the given Batch AST node.
	CreateBatch(statement *Batch)
	// StartBatch will start the worker threads for the batch, ready to execute any MethodCall(s) enqueued as work.
	StartBatch()
	// StopBatch will stop indicate to the internal Batch that there is no more work to execute and that we want to wait
	// for the workers to be finish. It should also set the BatchResults field to the results of this Batch.
	StopBatch()
	// GetEnvironment will return the currently used environment, or nil if there is no environment.
	GetEnvironment() (err error, env Env)
}

// CallStack is implemented by the call stack that is used within the VM.
type CallStack interface {
	// Call will add a new stack frame to the call stack with the given fields. It will also create a new Heap 
	// accordingly with the given computed arguments as values on the Heap.
	Call(caller *FunctionCall, current *FunctionDefinition, vm VM, args ...*data.Value) error
	// Return will remove the top frame from the call stack and return it.
	Return(vm VM) (err error, frame Frame)
	// Current will return the top of the call stack but not return it.
	Current() Frame
	// Stringer interface is used so that we can stringify the top couple of stack frames.
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
	CheckPass() bool
	IndentString
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
	GetStatement() *Batch
	Start(workers int)
	Stop() heap.Interface
}

// Env represents an environment variable that can be passed to a VM to set a global constant.
type Env interface {
	fmt.Stringer
	Merge(env Env) (err error)
	MergeN(envs... Env) (err error)
	GetPaths() []string
	GetValue() *data.Value
}
