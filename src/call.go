package main

import (
	"fmt"
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

// CallStack represents a stack of frames each of which representing a function call.
type CallStack []*Frame

// Frame is a stack frame which is allocated on the call stack.
type Frame struct {
	Caller  *parser.FunctionCall
	Current *parser.FunctionDefinition
}

// Call allocates a new stack Frame with the given caller and function definition and adds it to the top of the stack.
// Returns an error if there is a stack overflow as well as the allocated stack frame.
func (cs *CallStack) Call(caller *parser.FunctionCall, current *parser.FunctionDefinition) error {
	if len(*cs) == MaxStackFrames {
		return errors.StackOverflow.Errorf(MaxStackFrames)
	}

	*cs = append(*cs, &Frame{
		Caller:  caller,
		Current: current,
	})
	return nil
}

// Return pops off the topmost stack Frame and returns it. Also returns an error if there is a stack underflow.
func (cs *CallStack) Return() (err error, caller *parser.FunctionCall, current *parser.FunctionDefinition) {
	if len(*cs) == MinStackFrames {
		return errors.StackUnderFlow.Errorf(MinStackFrames), nil, nil
	}

	// Pop off the topmost frame
	frame := (*cs)[len(*cs) - 1]
	*cs = (*cs)[:len(*cs)]
	return nil, frame.Caller, frame.Current
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
		sb.WriteString(
			fmt.Sprintf(
				"File \"%s\", position %d:%d, in %s\n\t%s\n",
				frame.Caller.Pos.Filename,
				frame.Caller.Pos.Line,
				frame.Caller.Pos.Column,
				parentFrame.Current.JSONPath.String(0),
				frame.Caller.String(0),
			),
		)
	}
	return sb.String()
}
