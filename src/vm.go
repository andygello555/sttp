package main

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
	"io"
	"io/ioutil"
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
	// Debug is the io.Writer to write debugging information to. If this is ioutil.Discard then there will be no 
	// debugging information written (or evaluated).
	Debug           io.Writer
}

func New(testResults *TestResults, stdout io.Writer, stderr io.Writer, debug io.Writer) *VM {
	h := make(data.Heap)
	cs := make(CallStack, 0)

	if stdout == nil { stdout = os.Stdout }
	if stderr == nil { stderr = os.Stderr }
	if debug  == nil { debug  = ioutil.Discard }
	return &VM{
		Symbols: &h,
		Scope: 0,
		ParentStatement: nil,
		CallStack: &cs,
		TestResults: testResults,
		Stdout: stdout,
		Stderr: stderr,
		Debug: debug,
	}
}

func (vm *VM) Eval(filename, s string) (err error, result *data.Value) {
	// We start a panic catcher to give us more helpful error messages
	defer func() {
		if p := recover(); p != nil {
			switch p.(type) {
			case struct { errors.ProtoSttpError }:
				err = p.(struct { errors.ProtoSttpError })
			default:
				err = fmt.Errorf("%v", p)
			}
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

// GetDebug will return the io.Writer used for debugging. If the io.Writer is equal to ioutil.Discard, then false will 
// be returned, otherwise true will be returned.
func (vm *VM) GetDebug() (io.Writer, bool) {
	return vm.Debug, vm.Debug != ioutil.Discard
}
