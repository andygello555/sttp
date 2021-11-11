package parser

import (
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
)

type evalNode interface {
	Eval(vm VM) (err error, result *eval.Symbol)
}

func (p *Program) Eval(vm VM) (err error, result *eval.Symbol) {
	return p.Block.Eval(vm)
}

func (b *Block) Eval(vm VM) (err error, result *eval.Symbol) {
	// We return the last statement or return an error if one occurred in the statement
	for _, stmt := range b.Statements {
		err, result = stmt.Eval(vm)
		if err != nil {
			return err, nil
		}
	}

	// Then we can return either the result from the eval of a ReturnStatement or a ThrowStatement
	if b.Return != nil {
		return b.Return.Eval(vm)
	} else if b.Throw != nil {
		return b.Throw.Eval(vm)
	}

	return nil, result
}

func (s *Statement) Eval(vm VM) (err error, result *eval.Symbol) {
	return nil, nil
}

func (r *ReturnStatement) Eval(vm VM) (err error, result *eval.Symbol) {
	return nil, nil
}

func (t *ThrowStatement) Eval(vm VM) (err error, result *eval.Symbol) {
	return nil, nil
}
