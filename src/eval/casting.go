package eval

import (
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
)

// e is the "ID" for eFunc. This needs to be created so that there is an identity that can be checked for equivalence.
var e = eFunc

// castTable contains the functions that are used to cast one Symbol into another type. The rows represent the Type to
// cast from. Whereas, the columns represent the Type to cast to.
var castTable = [8][8]func(symbol *data.Symbol) (err error, cast *data.Symbol) {
	/*                NoType    Object    Array    String    Number    Boolean    Null    Function                    */
	/* NoType   */ {       e,        e,       e,        e,        e,         e,      e,          e},
	/* Object   */ {       e,     same, obArray,        s,        l,     lBool,      e,          e},
	/* Array    */ {       e, arObject,    same,        s,        l,     lBool,      e,          e},
	/* String   */ {       e, stObject, stArray,     same,        l,     lBool,      e,          e},
	/* Number   */ {       e,   obSing,  arSing,        s,     same, nuBoolean,      e,          e},
	/* Boolean  */ {       e,   obSing,  arSing,        s, boNumber,      same,      e,          e},
	/* Null     */ {       e,   obSing,  arSing,        s, nlNumber, nlBoolean,   same,          e},
	/* Function */ {       e,        e,       e,        e,        e,         e,      e,       same},
}

// Castable checks whether the given symbol can be cast to the given type. This just checks the entry in the appropriate
// cell in castTable against e. Therefore, this function does not carry out the cast function itself so will not report
// back on any errors that may happen. Returns true if the entry in the castTable is not e, false otherwise.
func Castable(symbol *data.Symbol, to data.Type) bool {
	return &castTable[symbol.Type][to] != &e
}

// Cast will cast the given Symbol to the given type using the castTable matrix. If you cannot cast from the given
// symbol to the given type the errors.CannotCast error will be filled and returned. Otherwise, the cast function at the
// corresponding entry in the castTable will be executed and the result will be returned.
func Cast(symbol *data.Symbol, to data.Type) (err error, cast *data.Symbol) {
	// If the entry in the matrix points to e then we will return a CannotCast error.
	if !Castable(symbol, to) {
		return errors.CannotCast.Errorf(symbol.Type.String(), to.String()), nil
	}
	// Otherwise, we return the result of the cast method.
	return castTable[symbol.Type][to](symbol)
}
