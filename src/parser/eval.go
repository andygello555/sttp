package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
)

type evalNode interface {
	Eval(vm VM) (err error, result *data.Symbol)
}

func (p *Program) Eval(vm VM) (err error, result *data.Symbol) {
	// We insert a nil stack frame to indicate the bottom of the stack
	err = vm.GetCallStack().Call(nil, nil)
	if err != nil {
		return err, nil
	}
	err, result = p.Block.Eval(vm)
	if err == nil {
		err, _ = vm.GetCallStack().Return()
	}
	return err, result
}

func (b *Block) Eval(vm VM) (err error, result *data.Symbol) {
	// We return the last statement or return an error if one occurred in the statement
	for _, stmt := range b.Statements {
		err, result = stmt.Eval(vm)
		if err != nil {
			return err, nil
		}
	}

	// Then we can return either the result from the data of a ReturnStatement or a ThrowStatement
	if b.Return != nil {
		return b.Return.Eval(vm)
	} else if b.Throw != nil {
		return b.Throw.Eval(vm)
	}

	return nil, result
}

func (r *ReturnStatement) Eval(vm VM) (err error, result *data.Symbol) {
	// Set the Return field of the current stack Frame
	current := vm.GetCallStack().Current()
	err, result = r.Value.Eval(vm)
	*(current.GetReturn()) = *result
	return err, result
}

func (t *ThrowStatement) Eval(vm VM) (err error, result *data.Symbol) {
	err, result = t.Value.Eval(vm)
	if err == nil {
		if result == nil {
			result = &data.Symbol{
				Value: nil,
				Type:  data.Null,
				Scope: *vm.GetScope(),
			}
		}
		err = errors.Exception.Errorf(result.String(), t.Pos.String())
	}
	return err, result
}

func (s *Statement) Eval(vm VM) (err error, result *data.Symbol) {
	switch {
	case s.Assignment != nil:
		err, result = s.Assignment.Eval(vm)
	case s.FunctionCall != nil:
		err, result = s.FunctionCall.Eval(vm)
	case s.MethodCall != nil:
		err, result = s.MethodCall.Eval(vm)
	case s.Break != nil:
		return nil, nil
	case s.Test != nil:
		err, result = s.Test.Eval(vm)
	case s.While != nil:
		err, result = s.While.Eval(vm)
	case s.For != nil:
		err, result = s.For.Eval(vm)
	case s.ForEach != nil:
		err, result = s.ForEach.Eval(vm)
	case s.Batch != nil:
		err, result = s.Batch.Eval(vm)
	case s.TryCatch != nil:
		err, result = s.TryCatch.Eval(vm)
	case s.FunctionDefinition != nil:
		err, result = s.FunctionDefinition.Eval(vm)
	case s.IfElifElse != nil:
		err, result = s.IfElifElse.Eval(vm)
	default:
		err = fmt.Errorf("statement is empty")
	}
	return err, result
}

func (a *Assignment) Eval(vm VM) (err error, result *data.Symbol) {
	// Then we convert the JSONPath to a Path representation which can be easily iterated over.
	var path Path
	err, path = a.JSONPath.Convert(vm)
	if err != nil {
		return err, nil
	}
	// We get the root identifier of the JSONPath. This is the variable name.
	variableName := path[0].(string)

	// Then we get the value of the variable from the heap so that we can set its new value appropriately.
	var variableVal *data.Symbol
	err, variableVal = vm.GetCallStack().Current().GetHeap().Get(variableName, -1)
	// If it cannot be found then we will set the value to be null initially.
	if err != nil {
		variableVal = &data.Symbol{
			Value: nil,
			Type:  data.Null,
			Scope: *vm.GetScope(),
		}
	}

	// Evaluate the RHS
	err, result = a.Value.Eval(vm)
	if err != nil {
		return err, nil
	}

	// Then we set the current value using, the path found previously, to the value on the RHS
	var val interface{}
	err, val = path.Set(variableVal.Value, result.Value)
	if err != nil {
		return err, nil
	}

	// Finally, we assign the new value to the variable on the heap
	vm.GetCallStack().Current().GetHeap().Assign(variableName, val, *vm.GetScope())
	return nil, nil
}

func (e *Expression) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}
