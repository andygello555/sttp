package main

import (
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
)

// VM represents the current state of the sttp virtual machines.
type VM struct {
	// Heap contains the variables currently on the heap
	Heap            *eval.Heap
	// Scope is the current scope that the VM is in
	Scope           int
	// ParentStatement is a pointer to the first parent of the currently evaluated node
	ParentStatement interface{}
}

func New() *VM {
	h := make(eval.Heap)
	return &VM{
		Heap: &h,
		Scope: 0,
		ParentStatement: nil,
	}
}

func (vm *VM) Eval(filename, s string) (result *eval.Symbol, err error) {
	var program *parser.Program
	err, program = parser.Parse(filename, s)
	if err != nil {
		return nil, err
	}
	result = program.Eval(vm)
	return result, nil
}

func (vm *VM) GetHeap() *eval.Heap {
	return vm.Heap
}

func (vm *VM) GetScope() *int {
	return &vm.Scope
}

func (vm *VM) GetParentStatement() interface{} {
	return vm.ParentStatement
}
