package parser

import (
	"container/heap"
	"fmt"
	"github.com/andygello555/src/data"
	"github.com/andygello555/src/errors"
	"github.com/andygello555/src/eval"
	"reflect"
	"strings"
)

type evalNode interface {
	positionable
	Eval(vm VM) (err error, result *data.Value)
}

// Eval for program will...
//
// • Push a new nil Frame to the callstack.
//
// • The Block will be evaluated.
//
// • The Frame will be returned from.
//
// Any errors or results that have bubbled up from lower AST nodes will be returned.
func (p *Program) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(p.GetPos())
	// We insert a nil stack frame to indicate the bottom of the stack. We check if stack size is zero because if the
	// VM is in REPL mode, we do not want to add another bottommost stack frame onto of the original.
	if vm.GetCallStack().Size() == 0 {
		err = vm.GetCallStack().Call(nil, nil, vm)
	}

	// We insert the environment (if we have one)
	var env Env
	if err, env = vm.GetEnvironment(); err != nil {
		return err, nil
	} else if env != nil {
		// Assign the "env" variable to the environment
		if err = vm.GetCallStack().Current().GetHeap().Assign(
			"env",
			env.GetValue().Value,
			true,
			true,
		); err != nil {
			return err, nil
		}
		if debug, ok := vm.GetDebug(); ok {
			_, _ = fmt.Fprintf(debug, "environment was given: %s\n", env.String())
		}
	} else {
		if debug, ok := vm.GetDebug(); ok {
			_, _ = fmt.Fprint(debug, "environment was not given\n")
		}
	}

	// Evaluate the inner Block
	if err, result = p.Block.Eval(vm); err == nil {
		if debug, ok := vm.GetDebug(); ok {
			_, _ = fmt.Fprintf(debug, "final stack frame heap: %v\n", vm.GetCallStack().Current().GetHeap())
		}

		// When we are not in REPL mode, we will pop the bottommost stack frame
		if !vm.CheckREPL() {
			err, _ = vm.GetCallStack().Return(vm)
		}
	} else if purposeful, ok := err.(errors.PurposefulError); ok {
		// We check if the error is a purposeful error
		switch purposeful {
		case errors.Break:
			// Exchange the error for a more informative one
			err = errors.BreakOutsideLoop.Errorf(vm)
		case errors.Throw:
			// Wrap the user error within a go error
			errVal, _ := errors.ConstructSttpError(err, result.Value)
			err = errors.Errorf(vm, "RuntimeError", "ThrowError", "throw not caught: %s", errVal)
		case errors.Return:
			// Nilify the error in case of Return
			err = nil
		default:
			break
		}
	}
	return err, result
}

// Eval for Block will evaluate each Statement within it. A Block can end with either a ReturnStatement or a
// ThrowStatement. If either exist they will also be evaluated and returned before the Statements are returned.
func (b *Block) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(b.GetPos())
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

// Eval for ReturnStatement, will evaluate the Value if it exists. If it doesn't, then the Value returned will be
// data.Null. This data.Value will then be returned with an errors.Return purposeful error.
func (r *ReturnStatement) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(r.GetPos())
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

// Eval for ThrowStatement works in a similar way to ReturnStatement.Eval. It will calculate the Value (or data.Null)
// and return errors.Throw, which is a purposeful error.
func (t *ThrowStatement) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(t.GetPos())
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

// Eval for Statement will evaluate the matched non-terminal.
func (s *Statement) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(s.GetPos())
	switch {
	case s.Assignment != nil:
		err, result = s.Assignment.Eval(vm)
	case s.FunctionCall != nil:
		err, result = s.FunctionCall.Eval(vm)
	case s.MethodCall != nil:
		err, result = s.MethodCall.Eval(vm)
	case s.Break != nil:
		return errors.Break, nil
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

// Eval for Assignment. Will execute the following steps:
//
// 1. Converts the JSONPath to a Path and gets the root property from it to find the root property's value. If the root
//    property's value is not found, we default to data.Null.
//
// 2. The RHS (Value) of the Assignment is then evaluated. If the RHS is a data.Function, then we will create a copy of
//    the FunctionDefinition, changing the JSONPath to the JSONPath of the Assignment.
//
// 3. We then set this evaluated value on the RHS using the Path we converted earlier.
func (a *Assignment) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(a.GetPos())
	// Then we convert the JSONPath to a Path representation which can be easily iterated over.
	var path Path
	err, path = a.JSONPath.Convert(vm)
	if err != nil {
		return err, nil
	}
	// We get the root identifier of the JSONPath. This is the variable name.
	variableName := path[0].(string)

	hp := vm.GetCallStack().Current().GetHeap()

	// Then we get value of the variable.
	variableVal := hp.Get(variableName)
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
	err, val = path.Set(vm, variableVal.Value, result.Value)
	if err != nil {
		return err, nil
	}

	// Finally, we assign the new value to the variable on the heap
	err = hp.Assign(variableName, val, *vm.GetScope() == 0, false)
	if err != nil {
		return errors.UpdateError(err, vm), nil
	}

	if debug, ok := vm.GetDebug(); ok {
		_, _ = fmt.Fprintf(debug, "after assignment of %s heap is: %v global: %t scope: %d\n", a.JSONPath.String(0), hp, hp.Get(variableName).Global, *vm.GetScope())
	}
	return nil, nil
}

// Eval for MethodCall will first evaluate all Arguments given to it. Then, depending on whether we are currently within
// Batch and in our first pass, currently within a Batch and in the second pass, or not in a Batch at all.
//
// First pass of Batch: add the MethodCall as work to the work queue.
//
// Second pass of Batch: pop the next result in the BatchSuite's result queue. If there isn't a result to pop then
// return an errors.MethodCallMismatchInBatch. If there is one but the pointer does not point to the current MethodCall,
// then also errors.MethodCallMismatchInBatch. Otherwise, the result and error from the freshly popped result will be
// returned.
//
// Not in Batch: returns the synchronously evaluated MethodCall.
func (m *MethodCall) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(m.GetPos())
	args := make([]*data.Value, len(m.Arguments))
	for i, arg := range m.Arguments {
		if err, args[i] = arg.Eval(vm); err != nil {
			return err, nil
		}
	}

	if batch, results := vm.GetBatch(); batch != nil && results == nil {
		// If we are currently batching MethodCalls then we will add the MethodCall to the vm.Batch and return null.
		batch.AddWork(m, args...)
		if debug, ok := vm.GetDebug(); ok {
			_, _ = fmt.Fprintf(debug, "adding %s %v to work queue\n", m.String(0), args)
		}
		return nil, &data.Value{
			Value: nil,
			Type:  data.Null,
		}
	} else if batch != nil && results != nil {
		if results.Len() > 0 {
			// If we have batched results available, aka. the batch has been executed, then we will pop the next result and
			// return its error and data.Value.
			r := heap.Pop(results).(BatchResult)
			// If the current result's MethodCall pointer does not match the pointer to the current MethodCall then we will
			// return the appropriate error.
			if debug, ok := vm.GetDebug(); ok {
				_, _ = fmt.Fprintf(debug, "popped %s from result queue\n", r.GetMethodCall().String(0))
			}
			if r.GetMethodCall() != m {
				return errors.MethodCallMismatchInBatch.Errorf(vm, r.GetMethodCall().String(0), m.String(0)), nil
			}
			return errors.UpdateError(r.GetErr(), vm), r.GetValue()
		} else {
			// If we have not got anymore results then we have a mismatch of batched MethodCalls.
			return errors.MethodCallMismatchInBatch.Errorf(vm, "null (no more results)", m.String(0)), nil
		}
	} else {
		// Otherwise, we are just executing the MethodCall normally.
		err, result = m.Method.Call(args...)
		return errors.UpdateError(err, vm), result
	}
}

// Eval for TestStatement will first check if there are TestResults defined within the VM, if not then fresh TestResults
// will be created just for the execution of this script. It's worth noting that if there are TestResults defined within
// the VM, this means that either TestResults have been generated earlier in the script, or the script is being executed
// as part of a TestSuite. After this, a function is deferred to add the result of the test to the TestResults. Then,
// the Expression is evaluated and cast to a data.Boolean, if it is not already. If the BreakOnFailure flag is set on
// the TestResults' config, and the test does not pass, then we will return an errors.FailedTest.
func (t *TestStatement) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(t.GetPos())
	// If we don't have any TestResults, this means that we are not running in a test suite. We still want to create
	// TestResults to store results in just for this script.
	if !vm.CheckTestResults() {
		vm.CreateTestResults()
	}

	// We defer the addition of the test to simplify the logic within this node a bit
	passed := false
	defer func() {
		fmt.Println("adding", t, "to results")
		vm.GetTestResults().AddTest(t, passed)
	}()

	if err, result = t.Expression.Eval(vm); err == nil {
		if result.Type != data.Boolean {
			if err, result = eval.Cast(result, data.Boolean); err != nil {
				return errors.UpdateError(err, vm), nil
			}
		}

		passed = result.Value.(bool) == true
		// If the test has not passed and the BreakOnFailure flag has been set in the TestConfig, then we'll set the
		// error to FailedTest.
		if vm.GetTestResults().GetConfig().Get("BreakOnFailure").(bool) && !passed {
			err = errors.FailedTest
		}
	}
	return err, result
}

// Eval for While loop. Will evaluate the Block until the Condition does not hold.
func (w *While) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(w.GetPos())
	// Panic recovery makes returning errors a bit easier
	defer func() {
		if p := recover(); p != nil {
			switch p.(type) {
			case struct{ errors.ProtoSttpError }:
				err = p.(struct{ errors.ProtoSttpError })
			case errors.PurposefulError:
				// We ignore any errors thrown by the break statement
				if p.(errors.PurposefulError) == errors.Break {
					err = nil
				}
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
				panic(errors.UpdateError(err, vm))
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

// Eval for For loop. Will evaluate the Block until the condition does not hold. Will also apply the Step assignment
// after evaluating each iteration.
func (f *For) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(f.GetPos())
	// Evaluate the assignment
	if err, _ = f.Var.Eval(vm); err != nil {
		return err, nil
	}

	// Panic recovery makes returning errors a bit easier
	defer func() {
		if p := recover(); p != nil {
			switch p.(type) {
			case struct{ errors.ProtoSttpError }:
				err = p.(struct{ errors.ProtoSttpError })
			case errors.PurposefulError:
				// We ignore any errors thrown by the break statement
				if p.(errors.PurposefulError) == errors.Break {
					err = nil
				}
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
				panic(errors.UpdateError(err, vm))
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
			panic(err)
		}
		evalStep()
	}

	return err, nil
}

// Eval for ForEach loop. Will iterate over each value in the In value. If the In value is not a data.String,
// data.Object, or data.Array, then we will first try to eval.Cast In into a data.String, then a data.Object, and
// finally data.Array. If we cannot cast In to any of these, we will return an errors.CannotCast error. A data.Iterator
// will then be constructed to iterate over the values in the In value.
func (f *ForEach) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(f.GetPos())
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
			object := data.Object
			array := data.Array
			str := data.String
			return errors.CannotCast.Errorf(vm, result.Type.String(), strings.Join([]string{object.String(), array.String(), str.String()}, ", ")), nil
		}

		// Cast the value
		if err, result = eval.Cast(result, to); err != nil {
			return errors.UpdateError(err, vm), nil
		}
	}

	// Panic recovery makes returning errors a bit easier
	defer func() {
		if p := recover(); p != nil {
			switch p.(type) {
			case struct{ errors.ProtoSttpError }:
				err = p.(struct{ errors.ProtoSttpError })
			case errors.PurposefulError:
				// We ignore any errors thrown by the break statement
				if p.(errors.PurposefulError) == errors.Break {
					err = nil
				}
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

// Eval for Batch follows the following set of steps:
//
// 1. If there is a BatchSuite or results for the BatchSuite set up already then we will assume that there is a Batch
//    within a Batch. This means we will return an errors.BatchWithinBatch.
//
// 2. We cache the stdout and stderr file handlers, and set the interpreter to use temporary string buffers instead.
//    Then we create a deep copy of the current Frame's data.Heap, and set this copy as the new data.Heap for the
//    current Frame.
//
// 3. A BatchSuite is created, and its workers are started. The first pass of the Block is then initiated.
//
// 4. After this succeeds, we set the stdout and stderr file handlers to the ones cached before the first pass was
//    initiated. We also wait for the BatchSuite to execute all the work it was given just now.
//
// 5. If the BatchSuite has not executed any work (aka. there were no MethodCall(s) within the Batch) we will write the
//    temporary stdout and stderr to the cached stdout and stderr and set them back as the defaults. This effectively
//    just skips the second pass as it is unnecessary.
//
// 6. Otherwise, we will set the heap back to the old one, and then execute the second pass of the Block. If we still
//    have results in the BatchSuite result queue after executing the second pass, we will return an
//    errors.MethodCallMismatchInBatch.
//
// If an error occurs at any point in these steps, we will first set the stdout and stderr back to the cached ones, if
// we haven't done already. We will also delete the BatchSuite, so that the interpreter knows to not Batch anymore.
func (b *Batch) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(b.GetPos())
	batch, results := vm.GetBatch()
	if batch == nil && results == nil {
		// Replace Stdout and Stderr with temporary string buffers
		oldStdout, oldStderr := vm.GetStdout(), vm.GetStderr()
		var newStdout, newStderr strings.Builder
		vm.SetStdout(&newStdout)
		vm.SetStderr(&newStderr)

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
		// Start the BatchSuite workers...
		vm.StartBatch()

		// Evaluate the Block for the first time. This will enqueue work to the worker goroutines running within the
		// BatchSuite.
		if err, result = b.Block.Eval(vm); err != nil {
			// We also have to set the stdout and stderr back to their originals as well as stopping and deleting the
			// batch altogether
			vm.SetStdout(oldStdout)
			vm.SetStderr(oldStderr)
			vm.DeleteBatch()
			return err, nil
		}

		// Then we set the stdout and stderr back to the old ones.
		vm.SetStdout(oldStdout)
		vm.SetStderr(oldStderr)
		// Then we wait for all the work to be processed by stopping the BatchSuite...
		vm.StopBatch()

		// We find the number of results that have been processed. If nothing has been processed, then we optimise by
		// not running the Block again. We also keep the current copied over heap.
		_, results = vm.GetBatch()
		work := results.Len()

		if work > 0 {
			// We set the current heap back to the old one...
			*vm.GetCallStack().Current().GetHeap() = *oldHeap
			// Then we evaluate the Block again...
			if err, result = b.Block.Eval(vm); err != nil {
				vm.DeleteBatch()
				return err, nil
			}

			// If we still have results in the Batch that we haven't attached to a MethodCall. Then we will throw an
			// error.
			if _, results = vm.GetBatch(); results.Len() > 0 {
				vm.DeleteBatch()
				return errors.MethodCallMismatchInBatch.Errorf(vm, "null (too many results)", "null"), nil
			}
		} else {
			// If we have no work then we will write the new stdout and stderr into their respective io.Writers
			_, _ = fmt.Fprint(vm.GetStdout(), newStdout.String())
			_, _ = fmt.Fprint(vm.GetStderr(), newStderr.String())
		}

		// Finally, we delete the Batch, this will set both vm.Batch and vm.BatchResults back to nil.
		vm.DeleteBatch()
		return nil, nil
	}
	// We return an error if we are already in a Batch statement
	return errors.BatchWithinBatch.Errorf(vm), nil
}

// Eval for TryCatch will first execute the Block pointed to by the Try field. If Try returns an error then we will
// check if the error is user constructed by testing if the result returned by Try is not nil. If so we will construct
// a user defined error, otherwise we will construct a sttp error. This error will then be placed on the current heap
// as the CatchAs identifier. The Caught Block will then be executed.
func (tc *TryCatch) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(tc.GetPos())
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

// Eval for FunctionDefinition will place the pointer to this AST node on the heap at the JSONPath. The
// FunctionDefinition data.Value can only be Global and ReadOnly if the variable does not exist in the heap. Global is
// only set when the current scope is 0, and ReadOnly is only set if the FunctionDefinition is being set to a root
// property. Otherwise, if the variable can be found then the Global and ReadOnly flags are inherited from that
// variable.
func (f *FunctionDefinition) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(f.GetPos())
	// We convert the JSONPath to a Path representation which can be easily iterated over.
	var path Path
	err, path = f.JSONPath.Convert(vm)
	if err != nil {
		return err, nil
	}
	// We get the root identifier of the JSONPath. This is the variable name.
	variableName := path[0].(string)

	hp := vm.GetCallStack().Current().GetHeap()

	// Then we get value of the variable.
	variableVal := hp.Get(variableName)
	// If it cannot be found then we will set the value to be null initially.
	if variableVal == nil {
		variableVal = &data.Value{
			Value:  nil,
			Type:   data.Function,
			Global: *vm.GetScope() == 0,
			// The Value is only ReadOnly if the FunctionDefinition is being set to a root property.
			ReadOnly: len(path) == 0,
		}
	}

	// Then we set the current value using, the path found previously, to a value pointing to the FunctionDefinition
	var val interface{}
	err, val = path.Set(vm, variableVal.Value, f)
	if err != nil {
		return err, nil
	}

	// Finally, we assign the new value to the variable on the heap.
	// NOTE: The variable's Global and ReadOnly flags are inherited from the variableVal. This means that either the
	// function definition is stored within a fresh new Value of Type Function, or nested within another Value.
	err = hp.Assign(variableName, val, variableVal.Global, variableVal.ReadOnly)
	if err != nil {
		return err, nil
	}

	if debug, ok := vm.GetDebug(); ok {
		_, _ = fmt.Fprintf(debug, "after function definition heap is: %v\n", hp)
	}
	return nil, nil
}

// Eval for FunctionCall will have the following steps of execution:
//
// 1. Increment the VM scope. This will be decremented in a deferred function. Then the JSONPath is evaluated.
//
// 2. If the JSONPath returns a data.Value that is not of type data.Function, then the builtins will be checked if there
//    is only a root property. If neither of these conditions holds, then we will return an errors.Uncallable error.
//
// 3. If the value found is a pointer to a FunctionDefinition then we will evaluate all the Arguments, push a new Frame,
//    evaluate the FunctionDefinition's body and then return from the new Frame. The result returned will be the return
//    value from the popped frame.
//
// 4. If the value found is a BuiltinFunction, then we will call the BuiltinFunction by passing all uncomputed Arguments
//    to it.
func (f *FunctionCall) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(f.GetPos())
	*vm.GetScope()++
	// We start a panic catcher to give us more helpful error messages
	defer func() {
		*vm.GetScope()--
		if p := recover(); p != nil {
			switch p.(type) {
			case struct{ errors.ProtoSttpError }:
				err = p.(struct{ errors.ProtoSttpError })
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
			return errors.Uncallable.Errorf(vm, result.Type.String()), nil
		}
	}

	// Check if the Golang type of the value
	switch result.Value.(type) {
	case *FunctionDefinition:
		var args []*data.Value
		if err, args = computeArgs(vm, f.Arguments...); err != nil {
			return err, nil
		}
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

		if debug, ok := vm.GetDebug(); ok {
			_, _ = fmt.Fprintf(debug, "returned from %s with return value %s\n", f.JSONPath.String(0), frame.GetReturn().String())
		}

		result = frame.GetReturn()
	case BuiltinFunction:
		if debug, ok := vm.GetDebug(); ok {
			_, _ = fmt.Fprintf(debug, "calling builtin function %s args: %v\n", *f.JSONPath.Parts[0].Property, f.Arguments)
		}
		if err, result = result.Value.(BuiltinFunction)(vm, f.Arguments...); err != nil {
			return err, result
		}
	default:
		panic(fmt.Errorf("function value has type %s", reflect.TypeOf(result.Value).String()))
	}

	return err, result
}

// Eval for IfElifElse will first evaluate the first IfCondition, if truthy, will then evaluate the IfBlock and return
// it. Otherwise, we will start evaluating the Elifs to see if any have a truthy condition. If not, we will evaluate the
// Else block if we have one.
func (i *IfElifElse) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(i.GetPos())
	evalBool := func(e *Expression) (err error, cond bool) {
		var val *data.Value
		// Evaluate the condition
		if err, val = e.Eval(vm); err != nil {
			return err, false
		}

		// We cast the val to a Boolean if it isn't one
		if val.Type != data.Boolean {
			if err, val = eval.Cast(val, data.Boolean); err != nil {
				return errors.UpdateError(err, vm), false
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
	vm.SetPos(n.GetPos())
	return nil, &data.Value{
		Value: nil,
		Type:  data.Null,
	}
}

// Eval for Boolean will return a data.Value with the underlying boolean value and a data.Boolean type.
func (b *Boolean) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(b.GetPos())
	return nil, &data.Value{
		Value: bool(*b),
		Type:  data.Boolean,
	}
}

// jsonPath is an interface that instances of both *JSONPath, and *JSONPathFactor implement.
type jsonPath interface {
	Pathable
	ASTNode
}

// jsonPathEval is called by both JSONPath.Eval and JSONPathFactor.Eval as the behaviour of both can be described as
// agnostic using the above interface and the "reflect" package.
func jsonPathEval(j jsonPath, vm VM) (err error, result *data.Value) {
	vm.SetPos(j.GetPos())
	var path Path
	err, path = j.Convert(vm)
	if err != nil {
		return err, nil
	}

	var variableVal *data.Value
	rootPropertyField := reflect.ValueOf(j).Elem().FieldByName("RootProperty")
	// If the type of j is *JSONPath OR the RootProperty field exists within the value of j, then we will consider j as
	// having a root property that is a string (variable identifier).
	if reflect.TypeOf(j) == reflect.TypeOf((*JSONPath)(nil)) || (rootPropertyField.IsValid() && !rootPropertyField.IsNil()) {
		// We get the root identifier of the JSONPath. This is the variable name.
		variableName := path[0].(string)
		// Then we get the value of the variable from the heap so that we can set its new value appropriately.
		variableVal = vm.GetCallStack().Current().GetHeap().Get(variableName)
	} else {
		// The root property in a Path of j is a Value
		variableVal = path[0].(*data.Value)
	}

	// If it cannot be found then we will set the value to be null initially.
	if variableVal == nil {
		variableVal = &data.Value{
			Value: nil,
			Type:  data.Null,
		}
	}

	// We get the value at the path and get the type of the value.
	var t data.Type
	var val interface{}
	if err, val = path.Get(vm, variableVal.Value); err != nil {
		return err, nil
	}
	err = t.Get(val)

	if debug, ok := vm.GetDebug(); ok {
		_, _ = fmt.Fprintf(debug, "getting %s from %s = %v\n", j.String(0), variableVal.String(), val)
	}
	if err != nil {
		return errors.UpdateError(err, vm), nil
	}

	return nil, &data.Value{
		Value:    val,
		Type:     t,
		ReadOnly: t == data.Function,
	}
}

// Eval for JSONPath calls Convert and then path.Get, to retrieve the Value at the given JSONPath. Will return data.Null
// if the JSONPath points to nothing.
func (j *JSONPath) Eval(vm VM) (err error, result *data.Value) {
	return jsonPathEval(j, vm)
}

// Eval for JSONPathFactor is similar to JSONPath.Eval. If the root property is not a root property that points to a
// variable we will Get the value requested from that root value.
func (j *JSONPathFactor) Eval(vm VM) (err error, result *data.Value) {
	return jsonPathEval(j, vm)
}

// jsonDeclaration recursively generates a sttp value from a JSON declaration.
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
			var key, val *data.Value
			var err error
			err, key = p.Key.Eval(vm)
			if err == nil {
				err, val = p.Value.Eval(vm)
				if err == nil {
					err, key = eval.Cast(key, data.String)
					if err == nil {
						obj[key.StringLit()] = val.Value
						continue
					}
				}
			}
			panic(errors.UpdateError(err, vm))
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

// Eval for JSON will use jsonDeclaration to construct the value and will then set the type appropriately.
func (j *JSON) Eval(vm VM) (err error, result *data.Value) {
	vm.SetPos(j.GetPos())
	defer func() {
		if p := recover(); p != nil {
			switch p.(type) {
			case struct{ errors.ProtoSttpError }:
				err = p.(struct{ errors.ProtoSttpError })
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
		Value: json,
		Type:  t,
	}
}
