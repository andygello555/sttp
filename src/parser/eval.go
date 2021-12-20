package parser

import (
	"container/heap"
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/eval"
	"io/ioutil"
	"reflect"
	"strings"
)

type evalNode interface {
	Eval(vm VM) (err error, result *data.Value)
}

func (p *Program) Eval(vm VM) (err error, result *data.Value) {
	// We insert a nil stack frame to indicate the bottom of the stack
	err = vm.GetCallStack().Call(nil, nil, vm)
	if err != nil {
		return err, nil
	}
	if err, result = p.Block.Eval(vm); err == nil {
		if debug, ok := vm.GetDebug(); ok {
			_, _ = fmt.Fprintf(debug, "final stack frame heap: %v\n", vm.GetCallStack().Current().GetHeap())
		}
		err, _ = vm.GetCallStack().Return(vm)
	}
	return err, result
}

func (b *Block) Eval(vm VM) (err error, result *data.Value) {
	// We return the last statement or return an error if one occurred in the statement
	for _, stmt := range b.Statements {
		if err, result = stmt.Eval(vm); err != nil {
			return err, result
		}
	}

	// Then we can return either the result from the data of a ReturnStatement or a ThrowStatement
	if b.Return != nil {
		return b.Return.Eval(vm)
	} else if b.Throw != nil {
		return b.Throw.Eval(vm)
	}

	return err, result
}

func (r *ReturnStatement) Eval(vm VM) (err error, result *data.Value) {
	current := vm.GetCallStack().Current()

	// Evaluate the return value, if we have one, and copy the value into valCopy.
	var valCopy *data.Value
	if r.Value != nil {
		if err, result = r.Value.Eval(vm); err != nil {
			return err, nil
		}
		valCopy = &data.Value{
			Value:    result.Value,
			Type:     result.Type,
			Global:   result.Global,
			ReadOnly: result.ReadOnly,
		}
	} else {
		valCopy = &data.Value{
			Value:    nil,
			Type:     data.Null,
			Global:   false,
			ReadOnly: false,
		}
	}

	// Set the Return field of the current stack Frame
	*current.GetReturn() = *valCopy
	return errors.Return, result
}

func (t *ThrowStatement) Eval(vm VM) (err error, result *data.Value) {
	if t.Value != nil {
		if err, result = t.Value.Eval(vm); err != nil {
			return err, nil
		}
	}

	if result == nil {
		result = &data.Value{
			Value:  nil,
			Type:   data.Null,
			Global: false,
		}
	}
	return errors.Throw, result
}

func (s *Statement) Eval(vm VM) (err error, result *data.Value) {
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

func (a *Assignment) Eval(vm VM) (err error, result *data.Value) {
	// Then we convert the JSONPath to a Path representation which can be easily iterated over.
	var path Path
	err, path = a.JSONPath.Convert(vm)
	if err != nil {
		return err, nil
	}
	// We get the root identifier of the JSONPath. This is the variable name.
	variableName := path[0].(string)

	heap := vm.GetCallStack().Current().GetHeap()

	// Then we get value of the variable.
	variableVal := heap.Get(variableName)
	// If it cannot be found then we will set the value to be null initially.
	if variableVal == nil {
		variableVal = &data.Value{
			Value:  nil,
			Type:   data.Null,
			Global: *vm.GetScope() == 0,
		}
	}

	// Evaluate the RHS
	err, result = a.Value.Eval(vm)
	if err != nil {
		return err, nil
	}

	// If the value we are setting to is a Function, then we will create a new *FunctionDefinition. This will be 
	// composed of the body of the function and the JSONPath of the variable that we are setting.
	if result.Type == data.Function {
		// This will only create a new pointer to store the function pointer in
		oldFunction := result.Value.(*FunctionDefinition)
		// This will create a new FunctionDefinition on the heap and set newFunction to be a pointer to it
		newFunction := &FunctionDefinition{
			Pos:      oldFunction.Pos,
			JSONPath: a.JSONPath,
			Body:     oldFunction.Body,
		}
		// Finally, we set the result.Value to be the newFunction that we have just created
		result.Value = newFunction
	}

	// Then we set the current value using, the path found previously, to the value on the RHS
	var val interface{}
	err, val = path.Set(variableVal.Value, result.Value)
	if err != nil {
		return err, nil
	}

	// Finally, we assign the new value to the variable on the heap
	err = heap.Assign(variableName, val, *vm.GetScope() == 0, false)
	if err != nil {
		return err, nil
	}

	if debug, ok := vm.GetDebug(); ok {
		_, _ = fmt.Fprintf(debug, "after assignment of %s heap is: %v global: %t scope: %d\n", a.JSONPath.String(0), heap, heap.Get(variableName).Global, *vm.GetScope())
	}
	return nil, nil
}

func (m *MethodCall) Eval(vm VM) (err error, result *data.Value) {
	args := make([]*data.Value, len(m.Arguments))
	for i, arg := range m.Arguments {
		if err, args[i] = arg.Eval(vm); err != nil {
			return err, nil
		}
	}

	if batch, results := vm.GetBatch(); batch != nil && results == nil {
		// If we are currently batching MethodCalls then we will add the MethodCall to the vm.Batch and return null.
		batch.AddWork(m.Method, args...)
		return nil, &data.Value{
			Value: nil,
			Type:  data.Null,
		}
	} else if batch != nil && results != nil {
		// If we have batched results available, aka. the batch has been executed, then we will pop the next result and
		// return its error and data.Value.
		r := heap.Pop(results).(Result)
		return r.GetErr(), r.GetValue()
	} else {
		// Otherwise, we are just executing the MethodCall normally.
		return m.Method.Call(args...)
	}
}

func (t *TestStatement) Eval(vm VM) (err error, result *data.Value) {
	// We defer the addition of the test to simplify the logic within this node a bit
	if vm.GetTestResults() != nil {
		passed := false
		defer func() {
			vm.GetTestResults().AddTest(t, passed)
		}()

		if err, result = t.Expression.Eval(vm); err == nil {
			if result.Type != data.Boolean {
				if err, result = eval.Cast(result, data.Boolean); err != nil {
					return err, nil
				}
			}

			passed = result.Value.(bool) == true
			// If the test has not passed and the BreakOnFailure flag has been set in the TestConfig, then we'll set the 
			// error to FailedTest.
			if vm.GetTestResults().GetConfig().Get("BreakOnFailure").(bool) && !passed {
				err = errors.FailedTest
			}
		}
	} else {
		panic(errors.NoTestSuite.Errorf(t.String(0)))
	}
	return err, result
}

func (w *While) Eval(vm VM) (err error, result *data.Value) {
	// Panic recovery makes returning errors a bit easier
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

	evalCond := func() bool {
		// Evaluate the condition
		if err, result = w.Condition.Eval(vm); err != nil {
			panic(err)
		}

		// Cast to Boolean if it isn't already
		if result.Type != data.Boolean {
			if err, result = eval.Cast(result, data.Boolean); err != nil {
				panic(err)
			}
		}
		return result.Value.(bool)
	}

	// Then we execute the while loop
	for evalCond() {
		if err, _ = w.Block.Eval(vm); err != nil {
			panic(err)
		}
	}
	return err, nil
}

func (f *For) Eval(vm VM) (err error, result *data.Value) {
	// Evaluate the assignment
	if err, _ = f.Var.Eval(vm); err != nil {
		return err, nil
	}

	// Panic recovery makes returning errors a bit easier
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

	evalCond := func() bool {
		// Evaluate the condition
		if err, result = f.Condition.Eval(vm); err != nil {
			panic(err)
		}

		// Cast to Boolean if it isn't already
		if result.Type != data.Boolean {
			if err, result = eval.Cast(result, data.Boolean); err != nil {
				panic(err)
			}
		}
		return result.Value.(bool)
	}

	evalStep := func() {
		// Evaluate the step
		if err, _ = f.Step.Eval(vm); err != nil {
			panic(err)
		}
	}

	// Then we do our loop
	for evalCond() {
		if err, _ = f.Block.Eval(vm); err != nil {
			return err, nil
		}
		evalStep()
	}

	return err, nil
}

func (f *ForEach) Eval(vm VM) (err error, result *data.Value) {
	// Find the value we are iterating over
	if err, result = f.In.Eval(vm); err != nil {
		return err, nil
	}

	// Check if the Value is a string, object or an array. If not then we will check if the value is castable. This is 
	// done in the order:
	// - String
	// - Object
	// - Array
	if result.Type != data.String && result.Type != data.Object && result.Type != data.Array {
		// Find out what we can cast the value to
		var to data.Type
		if eval.Castable(result, data.String) {
			to = data.String
		} else if eval.Castable(result, data.Object) {
			to = data.Object
		} else if eval.Castable(result, data.Array) {
			to = data.Array
		} else {
			object := data.Object; array := data.Array; str := data.String
			return errors.CannotCast.Errorf(result.Type.String(), strings.Join([]string{object.String(), array.String(), str.String()}, ", ")), nil
		}

		// Cast the value
		if err, result = eval.Cast(result, to); err != nil {
			return err, nil
		}
	}

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

	// Construct the data.Iterator for the value
	var iterator *data.Iterator
	if err, iterator = data.Iterate(result); err != nil {
		panic(err)
	}

	// Anon func to set the key and value iterators on each iteration
	set := func(elem *data.Element) {
		if err = vm.GetCallStack().Current().GetHeap().Assign(*f.Key, elem.Key.Value, elem.Key.Global, false); err != nil {
			panic(err)
		}
		if f.Value != nil {
			if err = vm.GetCallStack().Current().GetHeap().Assign(*f.Value, elem.Val.Value, elem.Val.Global, false); err != nil {
				panic(err)
			}
		}
	}

	// Iterate over the iterator until we have nothing left.
	for iterator.Len() > 0 {
		//set(heap.Pop(iterator).(*data.Element))
		set(iterator.Next())
		if err, result = f.Block.Eval(vm); err != nil {
			panic(err)
		}
	}
	return err, nil
}

func (b *Batch) Eval(vm VM) (err error, result *data.Value) {
	batch, results := vm.GetBatch()
	if batch == nil && results == nil {
		// Replace Stdout and Stderr with ioutil.Discard
		oldStdout, oldStderr := vm.GetStdout(), vm.GetStderr()
		vm.SetStdout(ioutil.Discard); vm.SetStderr(ioutil.Discard)

		// Create a copy of the current heap
		newHeap := make(data.Heap)
		oldHeap := vm.GetCallStack().Current().GetHeap()
		for name, val := range *oldHeap {
			newHeap[name] = &data.Value{
				Value:    val.Value,
				Type:     val.Type,
				Global:   val.Global,
				ReadOnly: val.ReadOnly,
			}
		}
		// Use the copy as the new heap
		*vm.GetCallStack().Current().GetHeap() = newHeap
		// Set up the BatchSuite. vm.Batch is now not nil, but vm.BatchResults is...
		vm.CreateBatch(b)

		// Evaluate the Block for the first time. This will collect all the MethodCalls in the Batch to be executed in 
		// parallel.
		if err, result = b.Block.Eval(vm); err != nil {
			vm.DeleteBatch()
			return err, nil
		}

		// We then execute the collected MethodCalls. This will set vm.BatchResults to not be nil anymore.
		vm.ExecuteBatch()

		// Then we put back the old Heap and set the stdout and stderr back to the old ones.
		*vm.GetCallStack().Current().GetHeap() = *oldHeap
		vm.SetStdout(oldStdout); vm.SetStderr(oldStderr)

		// Then we evaluate the Block again...
		if err, result = b.Block.Eval(vm); err != nil {
			vm.DeleteBatch()
			return err, nil
		}

		// Finally, we delete the Batch, this will set both vm.Batch and vm.BatchResults back to nil.
		vm.DeleteBatch()
		return nil, nil
	}
	// We return an error if we are already in a Batch statement
	return errors.BatchWithinBatch.Errorf(), nil
}

func (tc *TryCatch) Eval(vm VM) (err error, result *data.Value) {
	if err, result = tc.Try.Eval(vm); err != nil {
		// Check if the error is user constructed
		var userErr interface{} = nil
		if result != nil {
			userErr = result.Value
		}
		// Construct the sttp error using the given err and or userErr
		errVal, ret := errors.ConstructSttpError(err, userErr)
		if ret {
			return err, result
		}

		// Place the exception on the heap with the provided identifier
		if err = vm.GetCallStack().Current().GetHeap().Assign(*tc.CatchAs, errVal, false, false); err != nil {
			return err, nil
		}

		// Execute the catch block
		if err, result = tc.Caught.Eval(vm); err != nil {
			return err, nil
		}
	}
	return nil, nil
}

func (f *FunctionDefinition) Eval(vm VM) (err error, result *data.Value) {
	// We convert the JSONPath to a Path representation which can be easily iterated over.
	var path Path
	err, path = f.JSONPath.Convert(vm)
	if err != nil {
		return err, nil
	}
	// We get the root identifier of the JSONPath. This is the variable name.
	variableName := path[0].(string)

	heap := vm.GetCallStack().Current().GetHeap()

	// Then we get value of the variable.
	variableVal := heap.Get(variableName)
	// If it cannot be found then we will set the value to be null initially.
	if variableVal == nil {
		variableVal = &data.Value{
			Value:    nil,
			Type:     data.Function,
			Global:   true,
			ReadOnly: true,
		}
	}

	// Then we set the current value using, the path found previously, to a value pointing to the FunctionDefinition
	var val interface{}
	err, val = path.Set(variableVal.Value, f)
	if err != nil {
		return err, nil
	}

	// Finally, we assign the new value to the variable on the heap.
	// NOTE: The variable's Global and ReadOnly flags are inherited from the variableVal. This means that either the 
	// function definition is stored within a fresh new Value of Type Function, or nested within another Value.
	err = heap.Assign(variableName, val, variableVal.Global, variableVal.ReadOnly)
	if err != nil {
		return err, nil
	}

	if debug, ok := vm.GetDebug(); ok {
		_, _ = fmt.Fprintf(debug,"after function definition heap is: %v\n", heap)
	}
	return nil, nil
}

func (f *FunctionCall) Eval(vm VM) (err error, result *data.Value) {
	*vm.GetScope() ++
	// We start a panic catcher to give us more helpful error messages
	defer func() {
		*vm.GetScope() --
		if p := recover(); p != nil {
			switch p.(type) {
			case struct { errors.ProtoSttpError }:
				err = p.(struct { errors.ProtoSttpError })
			default:
				err = fmt.Errorf("%v", p)
			}
		}
	}()

	if err, result = f.JSONPath.Eval(vm); err != nil {
		return err, nil
	}

	// If the value isn't a callable then we will return an error
	if result.Type != data.Function {
		// If there is a builtin with that variable name we'll retrieve it
		if CheckBuiltin(*f.JSONPath.Parts[0].Property) && len(f.JSONPath.Parts) == 1 {
			err, result = nil, GetBuiltin(*f.JSONPath.Parts[0].Property)
		} else {
			return errors.Uncallable.Errorf(result.Type.String()), nil
		}
	}

	calculateArgs := func() []*data.Value {
		// Evaluate arguments and create a list of args
		args := make([]*data.Value, len(f.Arguments))
		for i, arg := range f.Arguments {
			if err, args[i] = arg.Eval(vm); err != nil {
				panic(err)
			}
		}
		return args
	}

	// Check if the Golang type of the value
	switch result.Value.(type) {
	case *FunctionDefinition:
		args := calculateArgs()
		// Construct the new stack frame and put it on the callstack
		if err = vm.GetCallStack().Call(f, result.Value.(*FunctionDefinition), vm, args...); err != nil {
			return err, nil
		}

		if debug, ok := vm.GetDebug(); ok {
			_, _ = fmt.Fprintf(debug, "calling function %s type: %s args: %v\n", f.JSONPath.String(0), result.Type.String(), args)
			_, _ = fmt.Fprintf(debug, "new heap: %v\n", vm.GetCallStack().Current().GetHeap())
		}

		// Evaluate the Block within the definition
		if err, result = result.Value.(*FunctionDefinition).Body.Block.Eval(vm); err != nil {
			switch err.(type) {
			case errors.PurposefulError:
				// If we have a purposeful error then we will check if it is Return. If so we will set err to nil.
				if err.(errors.PurposefulError) == errors.Return {
					err = nil
					break
				}
				return err, result
			default:
				return err, nil
			}
		}


		// Return the stack frame
		var frame Frame
		if err, frame = vm.GetCallStack().Return(vm); err != nil {
			return err, nil
		}
		result = frame.GetReturn()
	case func(vm VM, args ...*data.Value) (err error, value *data.Value):
		args := calculateArgs()
		if err, result = result.Value.(func(vm VM, args ...*data.Value) (err error, value *data.Value))(vm, args...); err != nil {
			return err, result
		}
	default:
		panic(fmt.Errorf("function value has type %s", reflect.TypeOf(result.Value).String()))
	}

	return err, result
}

func (i *IfElifElse) Eval(vm VM) (err error, result *data.Value) {
	evalBool := func(e *Expression) (err error, cond bool) {
		var val *data.Value
		// Evaluate the condition
		if err, val = e.Eval(vm); err != nil {
			return err, false
		}

		// We cast the val to a Boolean if it isn't one
		if val.Type != data.Boolean {
			if err, val = eval.Cast(val, data.Boolean); err != nil {
				return err, false
			}
		}
		return nil, val.Value.(bool)
	}

	var cond bool
	if err, cond = evalBool(i.IfCondition); err != nil {
		return err, nil
	}

	if cond {
		return i.IfBlock.Eval(vm)
	} else {
		// We evaluate the condition of each Elif statement and if it evals to true then we return the evaluation of 
		// the block of that Elif
		for _, elif := range i.Elifs {
			if err, cond = evalBool(elif.Condition); err != nil {
				return err, nil
			}

			if cond {
				return elif.Block.Eval(vm)
			}
		}

		// If we haven't found any truthy Elif conditions we evaluate the Else block if we have one
		if i.Else != nil {
			return i.Else.Eval(vm)
		}
	}
	return nil, nil
}

// Eval for Null will just return a data.Value with a nil value and a data.Null type.
func (n *Null) Eval(vm VM) (err error, result *data.Value) {
	return nil, &data.Value{
		Value:  nil,
		Type:   data.Null,
		Global: *vm.GetScope() == 0,
	}
}

// Eval for Boolean will return a data.Value with the underlying boolean value and a data.Boolean type.
func (b *Boolean) Eval(vm VM) (err error, result *data.Value) {
	return nil, &data.Value{
		Value:  bool(*b),
		Type:   data.Boolean,
		Global: *vm.GetScope() == 0,
	}
}

// Eval for JSONPath calls Convert and then path.Get, to retrieve the Value at the given JSONPath. Will return data.Null
// if the JSONPath points to nothing.
func (j *JSONPath) Eval(vm VM) (err error, result *data.Value) {
	var path Path; err, path = j.Convert(vm)
	if err != nil {
		return err, nil
	}
	// We get the root identifier of the JSONPath. This is the variable name.
	variableName := path[0].(string)

	// Then we get the value of the variable from the heap so that we can set its new value appropriately.
	variableVal := vm.GetCallStack().Current().GetHeap().Get(variableName)
	// If it cannot be found then we will set the value to be null initially.
	if variableVal == nil {
		variableVal = &data.Value{
			Value:  nil,
			Type:   data.Null,
			Global: *vm.GetScope() == 0,
		}
	}

	// We get the value at the path and get the type of the value.
	var t data.Type
	val := path.Get(variableVal.Value)
	err = t.Get(val)
	if err != nil {
		return err, nil
	}

	return nil, &data.Value{
		Value:    val,
		Type:     t,
		Global:   *vm.GetScope() == 0,
		ReadOnly: t == data.Function,
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
			var key, val *data.Value; var err error
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

func (j *JSON) Eval(vm VM) (err error, result *data.Value) {
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

	return err, &data.Value{
		Value:  json,
		Type:   t,
		Global: *vm.GetScope() == 0,
	}
}
