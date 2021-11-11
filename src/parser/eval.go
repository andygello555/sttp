package parser

import (
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
)

type evalNode interface {
	Eval(vm VM) (err error, result *eval.Symbol)
}

func (p *Program) Eval(vm VM) (err error, result *eval.Symbol) {
	return err, result
}
