package parser

import "github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"

// ASTNode is implemented by all ASTNodes
type ASTNode interface {
	indentString
	evalNode
}

// VM acts as an interface for the overarching state of the VM used for evaluation of programs.
type VM interface {
	Eval(filename, s string) (result *eval.Symbol, err error)
	GetHeap() *eval.Heap
	GetScope() *int
	GetParentStatement() interface{}
	GetCallStack() CallStack
}

// CallStack is implemented by the call stack that is used within the VM.
type CallStack interface {
	Call(caller *FunctionCall, current *FunctionDefinition) error
	Return() (err error, caller *FunctionCall, current *FunctionDefinition)
}
