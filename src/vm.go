package main

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
	"io"
	"os"
)

// VM represents the current state of the sttp virtual machines.
type VM struct {
	// Symbols contains the symbols for global variables and functions.
	Symbols         *data.Heap
	// Scope is the current scope that the VM is in.
	Scope           int
	// ParentStatement is a pointer to the first parent of the currently evaluated node.
	ParentStatement interface{}
	// CallStack contains the current call stack state.
	CallStack       *CallStack
	// TestResults contains the tests that have been run.
	TestResults     *TestResults
	// Stdout is the io.Writer written to for print calls.
	Stdout          io.Writer
	// Stderr is the io.Writer written to for error calls.
	Stderr			io.Writer
}

func New(testResults *TestResults, stdout io.Writer, stderr io.Writer) *VM {
	h := make(data.Heap)
	cs := make(CallStack, 0)

	if stdout == nil { stdout = os.Stdout }
	if stderr == nil { stderr = os.Stderr }
	return &VM{
		Symbols: &h,
		Scope: 0,
		ParentStatement: nil,
		CallStack: &cs,
		TestResults: testResults,
		Stdout: stdout,
		Stderr: stderr,
	}
}

func (vm *VM) Eval(filename, s string) (err error, result *data.Value) {
	// We start a panic catcher to give us more helpful error messages
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()

	var program *parser.Program
	err, program = parser.Parse(filename, s)
	if err != nil {
		return err, nil
	}
	return program.Eval(vm)
}

func (vm *VM) GetSymbols() *data.Heap {
	return vm.Symbols
}

func (vm *VM) GetScope() *int {
	return &vm.Scope
}

func (vm *VM) GetParentStatement() interface{} {
	return vm.ParentStatement
}

func (vm *VM) GetCallStack() parser.CallStack {
	return vm.CallStack
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
