package eval

import (
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/data"
	"github.com/RHUL-CS-Projects/IndividualProject_2021_Jakab.Zeller/src/errors"
	"reflect"
)

// e is the "ID" for eFunc. This needs to be created so that there is an identity that can be checked for equivalence.
var e = eFunc

// castTable contains the functions that are used to cast one Value into another type. The rows represent the Type to
// cast from. Whereas, the columns represent the Type to cast to.
var castTable = [8][8]func(symbol *data.Value) (err error, cast *data.Value) {
	/*                NoType    Object    Array    String    Number    Boolean    Null    Function                    */
	/* NoType   */ {    same,        e,       e,        e,        e,         e,      e,          e},
	/* Object   */ {       e,     same, obArray,        s,        l,     lBool,      e,          e},
	/* Array    */ {       e, arObject,    same,        s,        l,     lBool,      e,          e},
	/* String   */ {       e, stObject, stArray,     same, stNumber,     lBool,      e,          e},
	/* Number   */ {       e,   obSing,  arSing,        s,     same, nuBoolean,      e,          e},
	/* Boolean  */ {       e,   obSing,  arSing,        s, boNumber,      same,      e,          e},
	/* Null     */ {       e,   obSing,  arSing,        s, nlNumber, nlBoolean,   same,          e},
	/* Function */ {       e,        e,       e,        s,        e,         e,      e,       same},
}

// Castable checks whether the given symbol can be cast to the given type. This just checks the entry in the appropriate
// cell in castTable against e. Therefore, this function does not carry out the cast function itself so will not report
// back on any errors that may happen. Returns true if the entry in the castTable is not e, false otherwise.
func Castable(symbol *data.Value, to data.Type) bool {
	return reflect.ValueOf(castTable[symbol.Type][to]).Pointer() != reflect.ValueOf(e).Pointer()
}

// Cast will cast the given Value to the given type using the castTable matrix. If you cannot cast from the given
// symbol to the given type the errors.CannotCast error will be filled and returned. Otherwise, the cast function at the
// corresponding entry in the castTable will be executed and the result will be returned.
func Cast(symbol *data.Value, to data.Type) (err error, cast *data.Value) {
	// If the entry in the matrix points to e then we will return a CannotCast error.
	if !Castable(symbol, to) {
		return errors.CannotCast.Errorf(errors.GetNullVM(), to.String()), nil
	}
	// Otherwise, we return the result of the cast method.
	return castTable[symbol.Type][to](symbol)
}

// CastInterface will construct a Value from the given interface{} and cast it to the given Type. Internally this uses
// Cast to achieve that.
func CastInterface(value interface{}, to data.Type) (err error, cast *data.Value) {
	var t data.Type
	if err = t.Get(value); err != nil {
		return err, nil
	}
	return Cast(&data.Value{
		Value: value,
		Type:  t,
	}, to)
}
