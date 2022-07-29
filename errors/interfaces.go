package errors

import "github.com/alecthomas/participle/v2/lexer"

type SttpError interface {
	Errorf(values... interface{}) error
}

// NullVM is an implementation of errors.VM which is used in packages that cannot import the parser package and 
// therefore cannot pass a parser.VM implementation to one of the Errorf methods.
type NullVM struct {
	GetPosMethod            func() lexer.Position
	GetCallStackValueMethod func() []interface{}
}

func (vm NullVM) GetPos() lexer.Position { return vm.GetPosMethod() }
func (vm NullVM) CallStackValue() []interface{} { return vm.GetCallStackValueMethod() }

func GetNullVM() VM {
	nullVM := struct { NullVM }{}
	nullVM.GetPosMethod = func() lexer.Position { return lexer.Position{} }
	nullVM.GetCallStackValueMethod = func() []interface{} { return []interface{}{} }
	return nullVM
}

type VM interface {
	GetPos() lexer.Position
	CallStackValue() []interface{}
}

// UpdateError will update the Pos, CallStack and FromNullVM of a ProtoSttpError. If the given error is not a 
// ProtoSttpError, then the untouched error will be returned.
func UpdateError(err error, vm VM) error {
	if err != nil {
		switch err.(type) {
		case struct{ ProtoSttpError }:
			// We give the ProtoSttpError a proper VM to get the position and callstack from the given VM.
			sttpErr := err.(struct{ ProtoSttpError })
			if sttpErr.FromNullVM && vm != nil {
				sttpErr.UpdateVM(vm)
			}
			return sttpErr
		default:
			return err
		}
	}
	return nil
}
