package main

import (
	"container/heap"
	"fmt"
	"github.com/alecthomas/participle/v2/lexer"
	"github.com/andygello555/data"
	"github.com/andygello555/errors"
	"github.com/andygello555/parser"
	"io"
	"io/ioutil"
	"os"
)

// VM represents the current state of the sttp virtual machines.
type VM struct {
	// Pos is the current position of the interpreter. This is used when throwing errors.
	Pos lexer.Position
	// Scope is the current scope that the VM is in. It is used to check if a data.Value should be defined as global.
	Scope int
	// CallStack contains the current call stack state.
	CallStack *CallStack
	// TestResults contains the tests that have been run.
	TestResults *TestResults
	// Stdout is the io.Writer written to for print calls.
	Stdout io.Writer
	// Stderr is the io.Writer written to for error calls.
	Stderr io.Writer
	// Debug is the io.Writer to write debugging information to. If this is ioutil.Discard then there will be no
	// debugging information written (or evaluated).
	Debug io.Writer
	// The BatchSuite used for parser.Batch statements. If nil then the VM is not currently in a parser.Batch statement.
	Batch parser.BatchSuite
	// BatchResults contains the results of the executed BatchSuite. If nil then the VM is not currently in a
	// parser.Batch statement, or the BatchSuite has not yet been executed.
	BatchResults heap.Interface
	// All the environments passed to this VM when evaluating a script. This is so that inheritance can take place
	// within TestSuites.
	Environments []parser.Env
	// Whether the VM is running in REPL mode. This will not remove the bottommost stack frame at the end of
	// parser.Program Eval().
	REPL bool
}

func New(repl bool, testResults *TestResults, stdout io.Writer, stderr io.Writer, debug io.Writer, envs ...parser.Env) *VM {
	cs := make(CallStack, 0)

	if stdout == nil {
		stdout = os.Stdout
	}
	if stderr == nil {
		stderr = os.Stderr
	}
	if debug == nil {
		debug = ioutil.Discard
	}
	return &VM{
		Scope:        0,
		CallStack:    &cs,
		TestResults:  testResults,
		Stdout:       stdout,
		Stderr:       stderr,
		Debug:        debug,
		Batch:        nil,
		BatchResults: nil,
		Environments: envs,
		REPL:         repl,
	}
}

func (vm *VM) Eval(filename, s string) (err error, result *data.Value) {
	// We start a panic catcher to give us more helpful error messages
	defer func() {
		if p := recover(); p != nil {
			switch p.(type) {
			case struct{ errors.ProtoSttpError }:
				err = p.(struct{ errors.ProtoSttpError })
			default:
				err = fmt.Errorf("%v", p)
			}
		}
	}()

	// Parse the script
	var program *parser.Program
	if err, program = parser.Parse(filename, s); err != nil {
		return err, nil
	}

	// We execute the Program that was parsed
	return program.Eval(vm)
}

func (vm *VM) GetPos() lexer.Position {
	return vm.Pos
}

func (vm *VM) SetPos(position lexer.Position) {
	vm.Pos = position
}

func (vm *VM) GetScope() *int {
	return &vm.Scope
}

func (vm *VM) GetCallStack() parser.CallStack {
	return vm.CallStack
}

func (vm *VM) CallStackValue() []interface{} {
	return vm.CallStack.Value()
}

func (vm *VM) CheckTestResults() bool {
	return vm.TestResults != nil
}

func (vm *VM) CreateTestResults() {
	vm.TestResults = &TestResults{
		Results: make([]*TestResult, 0),
		Config:  defaultTestConfig,
	}
}

func (vm *VM) GetTestResults() parser.TestResults {
	return vm.TestResults
}

func (vm *VM) GetStdout() io.Writer {
	return vm.Stdout
}

func (vm *VM) GetStderr() io.Writer {
	return vm.Stderr
}

func (vm *VM) SetStdout(stdout io.Writer) {
	vm.Stdout = stdout
}

func (vm *VM) SetStderr(stderr io.Writer) {
	vm.Stderr = stderr
}

// GetDebug will return the io.Writer used for debugging. If the io.Writer is equal to ioutil.Discard, then false will
// be returned, otherwise true will be returned.
func (vm *VM) GetDebug() (io.Writer, bool) {
	return vm.Debug, vm.Debug != ioutil.Discard
}

// WriteDebug will write to the Debug io.Writer if it exists, otherwise will be ignored.
func (vm *VM) WriteDebug(format string, a ...interface{}) {
	if debug, ok := vm.GetDebug(); ok {
		_, _ = fmt.Fprintf(debug, format, a...)
	}
}

func (vm *VM) GetBatch() (parser.BatchSuite, heap.Interface) {
	return vm.Batch, vm.BatchResults
}

// DeleteBatch will stop the workers, then nullify the Batch.
func (vm *VM) DeleteBatch() {
	vm.Batch.Stop()
	vm.Batch = nil
	vm.BatchResults = nil
}

func (vm *VM) CreateBatch(statement *parser.Batch) {
	vm.Batch = Batch(statement)
}

func (vm *VM) StartBatch() {
	vm.Batch.Start(-1)
}

func (vm *VM) StopBatch() {
	vm.BatchResults = vm.Batch.Stop()
}

func (vm *VM) GetEnvironment() (err error, env parser.Env) {
	if len(vm.Environments) == 0 {
		return nil, nil
	} else if len(vm.Environments) == 1 {
		return nil, vm.Environments[0]
	} else {
		// We will merge the environments together into an empty environment
		env = EmptyEnv()
		if err = env.MergeN(vm.Environments...); err != nil {
			return err, nil
		}
		vm.Environments = []parser.Env{env}
		return nil, vm.Environments[0]
	}
}

func (vm *VM) CheckREPL() bool {
	return vm.REPL
}
