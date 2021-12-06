package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"reflect"
	"testing"
)

// term describes the signature that all terms share. Each term has a left-hand side with a single higher precedence 
// term. As well as an array of factors on the right-hand side. They can also be evaluated.
type term interface {
	evalNode
	// Gets left-hand side node which can be evaluated.
	left() evalNode
	// Gets right-hand side nodes which can be evaluated and have an operator.
	right() []factor
}

// factor describes the signature that all factors share. Each factor has an operator. They can also be evaluated.
type factor interface {
	evalNode
	// Gets the operator tied to the factor.
	operator() eval.Operator
	// Gets the inner term of the factor.
	inner() term
}

// protoEvalNode implements evalNode but has a modifiable evalMethod. This means that a structure implementing 
// protoEvalNode can be defined anonymously, along with an Eval method defined then as well. This is used in the Factor 
// left and right referrers so that tokens passed straight from the lexer can be wrapped with an Eval referrer that 
// returns the token as a data.Value.
type protoEvalNode struct {
	evalMethod func(vm VM) (err error, result *data.Value)
}

// Eval calls the stored evalMethod.
func (p *protoEvalNode) Eval(vm VM) (err error, result *data.Value) { return p.evalMethod(vm) }

// tEval evaluates an AST node which implements the term interface.
func tEval(t term, vm VM) (err error, result *data.Value) {
	err, result = t.left().Eval(vm)
	if err != nil {
		return err, nil
	}

	if testing.Verbose() {
		fmt.Printf("\tLHS (%s) = %v\n", reflect.TypeOf(t.left()).String(), result)
	}

	for _, r := range t.right() {
		var right *data.Value
		err, right = r.Eval(vm)

		if testing.Verbose() {
			fmt.Printf("\tRHS (%s) = %v\n", reflect.TypeOf(r).String(), right)
		}

		if err == nil {
			err, result = eval.Compute(r.operator(), result, right)

			if testing.Verbose() {
				fmt.Println("\tnew LHS =", result)
			}
			if err == nil {
				continue
			}
		}
		return err, nil
	}
	return nil, result
}

// fEval evaluates an AST node which implements the factor interface.
func fEval(f factor, vm VM) (err error, result *data.Value) {
	return f.inner().Eval(vm)
}

func (e *Expression) left() evalNode { return e.Left }
func (e *Expression) right() []factor {f := make([]factor, len(e.Right)); for i, v := range e.Right {f[i] = v}; return f}

func (p5 *Prec5) operator() eval.Operator { return p5.Operator }
func (p5 *Prec5) inner() term { return p5.Factor }

func (p5t *Prec5Term) left() evalNode { return p5t.Left }
func (p5t *Prec5Term) right() []factor {f := make([]factor, len(p5t.Right)); for i, v := range p5t.Right {f[i] = v}; return f}

func (p4 *Prec4) operator() eval.Operator { return p4.Operator }
func (p4 *Prec4) inner() term { return p4.Factor }

func (p4t *Prec4Term) left() evalNode { return p4t.Left }
func (p4t *Prec4Term) right() []factor {f := make([]factor, len(p4t.Right)); for i, v := range p4t.Right {f[i] = v}; return f}

func (p3 *Prec3) operator() eval.Operator { return p3.Operator }
func (p3 *Prec3) inner() term { return p3.Factor }

func (p3t *Prec3Term) left() evalNode { return p3t.Left }
func (p3t *Prec3Term) right() []factor {f := make([]factor, len(p3t.Right)); for i, v := range p3t.Right {f[i] = v}; return f}

func (p2 *Prec2) operator() eval.Operator { return p2.Operator }
func (p2 *Prec2) inner() term { return p2.Factor }

func (p2t *Prec2Term) left() evalNode { return p2t.Left }
func (p2t *Prec2Term) right() []factor {f := make([]factor, len(p2t.Right)); for i, v := range p2t.Right {f[i] = v}; return f}

func (p1 *Prec1) operator() eval.Operator { return p1.Operator }
func (p1 *Prec1) inner() term { return p1.Factor }

func (p1t *Prec1Term) left() evalNode { return p1t.Left }
func (p1t *Prec1Term) right() []factor {f := make([]factor, len(p1t.Right)); for i, v := range p1t.Right {f[i] = v}; return f}

func (p0 *Prec0) operator() eval.Operator { return p0.Operator }
func (p0 *Prec0) inner() term { return p0.Factor }

func (f *Factor) left() evalNode {
	var n evalNode
	switch {
	case f.Null != nil:
		n = f.Null
	case f.Boolean != nil:
		n = f.Boolean
	case f.Number != nil:
		en := struct { protoEvalNode }{}
		en.evalMethod = func(vm VM) (err error, result *data.Value) {
			return nil, &data.Value{
				Value:  *f.Number,
				Type:   data.Number,
				Global: *vm.GetScope() == 0,
			}
		}
		n = &en
	case f.StringLit != nil:
		en := struct { protoEvalNode }{}
		en.evalMethod = func(vm VM) (err error, result *data.Value) {
			return nil, &data.Value{
				Value:  *f.StringLit,
				Type:   data.String,
				Global: *vm.GetScope() == 0,
			}
		}
		n = &en
	case f.JSONPath != nil:
		n = f.JSONPath
	case f.JSON != nil:
		n = f.JSON
	case f.FunctionCall != nil:
		n = f.FunctionCall
	case f.MethodCall != nil:
		n = f.MethodCall
	case f.SubExpression != nil:
		n = f.SubExpression
	}
	return n
}
func (f *Factor) right() []factor { return make([]factor, 0) }

func (e *Expression) Eval(vm VM) (err error, result *data.Value)  { return tEval(e, vm) }
func (p5 *Prec5) Eval(vm VM) (err error, result *data.Value)      { return fEval(p5, vm) }
func (p5t *Prec5Term) Eval(vm VM) (err error, result *data.Value) { return tEval(p5t, vm) }
func (p4 *Prec4) Eval(vm VM) (err error, result *data.Value)      { return fEval(p4, vm) }
func (p4t *Prec4Term) Eval(vm VM) (err error, result *data.Value) { return tEval(p4t, vm) }
func (p3 *Prec3) Eval(vm VM) (err error, result *data.Value)      { return fEval(p3, vm) }
func (p3t *Prec3Term) Eval(vm VM) (err error, result *data.Value) { return tEval(p3t, vm) }
func (p2 *Prec2) Eval(vm VM) (err error, result *data.Value)      { return fEval(p2, vm) }
func (p2t *Prec2Term) Eval(vm VM) (err error, result *data.Value) { return tEval(p2t, vm) }
func (p1 *Prec1) Eval(vm VM) (err error, result *data.Value)      { return fEval(p1, vm) }
func (p1t *Prec1Term) Eval(vm VM) (err error, result *data.Value) { return tEval(p1t, vm) }
func (p0 *Prec0) Eval(vm VM) (err error, result *data.Value)      { return fEval(p0, vm) }
func (f *Factor) Eval(vm VM) (err error, result *data.Value)      { return tEval(f, vm) }
