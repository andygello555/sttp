package main

import (
	"fmt"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/parser"
	"strings"
)

const (
	// MaxStackFrames are the maximum number of stack frames that can exist on the stack before an overflow can occur.
	MaxStackFrames      = 500
	// MinStackFrames are the minimum number of stack frames that can exist on the stack before an underflow can occur.
	MinStackFrames      = 0
	// MaxStackFramesPrint are the maximum number of stack frames that are printed when the stack is dumped (from the
	// top).
	MaxStackFramesPrint = 25
)

// Frame is a stack frame which is allocated on the call stack.
type Frame struct {
	// Caller is the reference to the FunctionCall node in the AST.
	Caller  *parser.FunctionCall
	// Current is the reference to the FunctionDefinition node in the AST.
	Current *parser.FunctionDefinition
	// Heap contains the parameters and local variables assigned within the function.
	Heap    *data.Heap
	// Return is the value/symbol returned by the function.
	Return  *data.Value
}

func (f *Frame) GetCaller() *parser.FunctionCall {
	return f.Caller
}

func (f *Frame) GetCurrent() *parser.FunctionDefinition {
	return f.Current
}

func (f *Frame) GetHeap() *data.Heap {
	return f.Heap
}

func (f *Frame) GetReturn() *data.Value {
	return f.Return
}

// CallStack represents a stack of frames each of which representing a function call.
type CallStack []*Frame

// Call allocates a new stack Frame with the given caller and function definition and adds it to the top of the stack.
// Returns an error if there is a stack overflow as well as the allocated stack frame.
func (cs *CallStack) Call(caller *parser.FunctionCall, current *parser.FunctionDefinition, vm parser.VM, args ...*data.Value) error {
	if len(*cs) == MaxStackFrames {
		return errors.StackOverflow.Errorf(vm, MaxStackFrames)
	}

	// Put a new Frame onto the stack
	heap := make(data.Heap)
	*cs = append(*cs, &Frame{
		Caller:  caller,
		Current: current,
		Heap: &heap,
		Return: &data.Value{
			Value:  nil,
			Type:   data.NoType,
			Global: false,
		},
	})

	// We only do this if this isn't our last stack frame
	if caller != nil && current != nil {
		params := current.Body.Parameters
		previous := (*cs)[len(*cs) - 2]
		// If there are more arguments than parameters then we'll return an error
		if len(args) > len(params) {
			return errors.MoreArgsThanParams.Errorf(vm, current.JSONPath.String(0), len(params), len(args))
		}

		// Copy over global variables from the previous stack frame
		for name, val := range *previous.Heap {
			if val.Global {
				heap[name] = val
			}
		}

		// Create the self variable on the heap. We do this by finding the JSONPath on the previous frame.
		self := previous.Heap.Get(*current.JSONPath.Parts[0].Property)
		if debug, ok := vm.GetDebug(); ok {
			_, _ = fmt.Fprintf(debug, "after getting self: %s\n", self.String())
		}
		if err := heap.Assign("self", self.Value, true, false); err != nil {
			return err
		}

		// Set arguments on the heap
		for i, param := range params {
			// Get the value to set the param on the heap to
			var val *data.Value
			if i < len(args) {
				val = args[i]
			} else {
				val = &data.Value{
					Value:    nil,
					Type:     data.Null,
					Global:   false,
					ReadOnly: false,
				}
			}

			// Check whether the param's JSONPath starts with "self"
			err, path := param.Convert(vm)
			if err != nil {
				return err
			}

			var pathVal *data.Value
			if path[0].(string) == "self" {
				pathVal = heap.Get("self")
			} else {
				// If the root property doesn't exist on the heap then we will create a null value
				if !heap.Exists(path[0].(string)) {
					if err = heap.Assign(path[0].(string), nil, false, false); err != nil {
						return err
					}
				}
				pathVal = heap.Get(path[0].(string))
			}

			// Then finally we set the value of the *data.Value
			if err, pathVal.Value = path.Set(vm, pathVal.Value, val.Value); err != nil {
				return err
			}
		}
	}
	return nil
}

// Current returns the currently "running" Frame and doesn't pop it.
func (cs *CallStack) Current() parser.Frame {
	return (*cs)[len(*cs) - 1]
}

// Return pops off the topmost stack Frame and returns it. Also returns an error if there is a stack underflow.
func (cs *CallStack) Return(vm parser.VM) (err error, frame parser.Frame) {
	if len(*cs) == MinStackFrames {
		return errors.StackUnderFlow.Errorf(vm, MinStackFrames), nil
	}

	// Pop off the topmost frame
	frame = (*cs)[len(*cs) - 1]
	(*cs)[len(*cs) - 1] = nil
	*cs = (*cs)[:len(*cs) - 1]

	if len(*cs) > 0 {
		heap := cs.Current().GetHeap()
		// Copy the globals back to the current stack frame
		for name, val := range *frame.GetHeap() {
			if val.Global {
				// If the variable is self we'll copy back the value into the variable denoted by the root property of the 
				// old frame's JSONPath.
				if name == "self" {
					var path parser.Path
					if err, path = frame.GetCurrent().JSONPath.Convert(vm); err != nil {
						return err, frame
					}
					name = path[0].(string)

					if debug, ok := vm.GetDebug(); ok {
						_, _ = fmt.Fprint(debug, "SELF: ")
					}
				}
				(*heap)[name].Value = val.Value

				if debug, ok := vm.GetDebug(); ok {
					_, _ = fmt.Fprintf(debug, "copying back %s to %s in function %s\n", name, val.String(), frame.GetCurrent().JSONPath.String(0))
				}
			}
		}
	}
	return nil, frame
}

// String returns the string representation of the stack. Useful for when errors occur within the VM. Stack frames are
// added to a string builder in reverse order with the caller location and the procedure name.
func (cs *CallStack) String() string {
	// We first get the most recent stack frames
	start := 0
	if len(*cs) > MaxStackFramesPrint {
		start = len(*cs) - (MaxStackFramesPrint + 1)
	}
	top := (*cs)[start:]

	var sb strings.Builder
	for i := len(top) - 1; i > 0; i-- {
		frame := top[i]
		parentFrame := top[i - 1]
		parentCurrent := ""
		if parentFrame.Current != nil {
			parentCurrent = fmt.Sprintf(", in %s", parentFrame.Current.JSONPath.String(0))
		}

		caller := ""
		if frame.Caller != nil {
			caller = fmt.Sprintf(
				"File \"%s\", position %d:%d%s\n\t%s\n",
				frame.Caller.Pos.Filename,
				frame.Caller.Pos.Line,
				frame.Caller.Pos.Column,
				parentCurrent,
				frame.Caller.String(0),
			)
		}

		sb.WriteString(caller)
	}
	return sb.String()
}

// Value gets the sttp value of the most recent stack frames. The returned value will be a []interface{} array which is 
// supported by sttp.
func (cs *CallStack) Value() []interface{} {
	// We first get the most recent stack frames
	start := 0
	if len(*cs) > MaxStackFramesPrint {
		start = len(*cs) - (MaxStackFramesPrint + 1)
	}
	top := (*cs)[start:]

	getPosMap := func(node parser.ASTNode) map[string]interface{} {
		return map[string]interface{} {
			"line": float64(node.GetPos().Line),
			"col": float64(node.GetPos().Column),
			"filename": node.GetPos().Filename,
		}
	}

	val := make([]interface{}, 0)
	for i := len(top) - 1; i > 0; i-- {
		frame := top[i]
		parentFrame := top[i - 1]
		frameMap := make(map[string]interface{})

		if parentFrame.Current != nil {
			frameMap["current"] = map[string]interface{} {
				"pos": getPosMap(parentFrame.Current),
				"function": parentFrame.Current,
				"string": parentFrame.Current.String(0),
			}
		}

		if frame.Caller != nil {
			frameMap["caller"] = map[string]interface{} {
				"pos": getPosMap(frame.Caller),
				"string": frame.Caller.String(0),
			}
		}

		val = append(val, frameMap)
	}
	return val
}
