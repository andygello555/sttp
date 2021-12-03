package parser

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"testing"
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

	if testing.Verbose() {
		fmt.Println("after assignment heap is:", vm.GetCallStack().Current().GetHeap())
	}
	return nil, nil
}

func (f *FunctionCall) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

func (m *MethodCall) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

func (t *TestStatement) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

func (w *While) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

func (f *For) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

func (f *ForEach) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

func (b *Batch) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

func (tc *TryCatch) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

func (f *FunctionDefinition) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

func (i *IfElifElse) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, nil
}

// Eval for Null will just return a data.Symbol with a nil value and a data.Null type.
func (n *Null) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, &data.Symbol{
		Value: nil,
		Type:  data.Null,
		Scope: 0,
	}
}

// Eval for Boolean will return a data.Symbol with the underlying boolean value and a data.Boolean type.
func (b *Boolean) Eval(vm VM) (err error, result *data.Symbol) {
	return nil, &data.Symbol{
		Value: *b,
		Type:  data.Boolean,
		Scope: 0,
	}
}

// Eval for JSONPath calls Convert and then path.Get, to retrieve the Symbol at the given JSONPath.
func (j *JSONPath) Eval(vm VM) (err error, result *data.Symbol) {
	var path Path; err, path = j.Convert(vm)
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

	// We get the value at the path and get the type of the value.
	var t data.Type
	val := path.Get(variableVal.Value)
	err = t.Get(val)
	if err != nil {
		return err, nil
	}

	return nil, &data.Symbol{
		Value: path.Get(variableVal.Value),
		Type:  t,
		Scope: *vm.GetScope(),
	}
}

func jsonDeclaration(j interface{}, vm VM) interface{} {
	var out interface{}
	switch j.(type) {
	case *Array:
		arr := make([]interface{}, len(j.(*Array).Elements))
		for i, e := range j.(*Array).Elements {
			arr[i] = jsonDeclaration(e, vm)
		}
		out = arr
	case *Object:
		obj := make(map[string]interface{})
		for _, p := range j.(*Object).Pairs {
			var key, val *data.Symbol; var err error
			err, key = p.Key.Eval(vm)
			if err == nil {
				err, val = p.Value.Eval(vm)
				if err == nil {
					err, key = eval.Cast(key, data.String)
					if err == nil {
						obj[key.Value.(string)] = val.Value
						continue
					}
				}
			}
			panic(err)
		}
		out = obj
	case *Expression:
		err, result := j.(*Expression).Eval(vm)
		if err != nil {
			panic(err)
		}
		out = result.Value
	}
	return out
} 

func (j *JSON) Eval(vm VM) (err error, result *data.Symbol) {
	defer func() {
		if p := recover(); p != nil {
			err = fmt.Errorf("%v", p)
		}
	}()

	var json interface{}
	switch {
	case j.Array != nil:
		json = jsonDeclaration(j.Array, vm)
	case j.Object != nil:
		json = jsonDeclaration(j.Object, vm)
	}

	var t data.Type
	switch json.(type) {
	case map[string]interface{}:
		t = data.Object
	case []interface{}:
		t = data.Array
	default:
		t = data.NoType
	}

	return err, &data.Symbol{
		Value: json,
		Type:  t,
		Scope: *vm.GetScope(),
	}
}
