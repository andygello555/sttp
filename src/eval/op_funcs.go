package eval

import "github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"

// oFunc is a dummy function which exists within the operatorTable to represent an operation which cannot be made.
func oFunc(op1 *data.Symbol, op2 *data.Symbol) (err error, result *data.Symbol) {
	return nil, nil
}
